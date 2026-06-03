package service

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type QueryResult struct {
	Columns      []string        `json:"columns"`
	Types        []string        `json:"types"`
	Rows         [][]interface{} `json:"rows"`
	RowsAffected int64           `json:"rowsAffected"`
	Duration     int64           `json:"duration"` // ms
	IsSelect     bool            `json:"isSelect"`
	Error        string          `json:"error,omitempty"`
}

func ExecuteQuery(db *sql.DB, query string) *QueryResult {
	start := time.Now()
	result := &QueryResult{}

	trimmed := strings.TrimSpace(strings.ToUpper(query))
	isSelect := strings.HasPrefix(trimmed, "SELECT") ||
		strings.HasPrefix(trimmed, "SHOW") ||
		strings.HasPrefix(trimmed, "DESCRIBE") ||
		strings.HasPrefix(trimmed, "EXPLAIN") ||
		strings.HasPrefix(trimmed, "WITH")

	if isSelect {
		rows, err := db.Query(query)
		if err != nil {
			result.Error = err.Error()
			result.Duration = time.Since(start).Milliseconds()
			return result
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			result.Error = err.Error()
			result.Duration = time.Since(start).Milliseconds()
			return result
		}
		result.Columns = columns
		result.IsSelect = true

		// get column types
		types, _ := rows.ColumnTypes()
		for _, t := range types {
			result.Types = append(result.Types, t.DatabaseTypeName())
		}

		for rows.Next() {
			values := make([]interface{}, len(columns))
			ptrs := make([]interface{}, len(columns))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				result.Error = err.Error()
				break
			}
			// convert []byte to string
			row := make([]interface{}, len(values))
			for i, v := range values {
				if b, ok := v.([]byte); ok {
					row[i] = string(b)
				} else {
					row[i] = v
				}
			}
			result.Rows = append(result.Rows, row)
		}
		if err := rows.Err(); err != nil {
			result.Error = err.Error()
		}
	} else {
		res, err := db.Exec(query)
		if err != nil {
			result.Error = err.Error()
			result.Duration = time.Since(start).Milliseconds()
			return result
		}
		affected, _ := res.RowsAffected()
		result.RowsAffected = affected
	}

	result.Duration = time.Since(start).Milliseconds()
	return result
}

// GetDatabases returns list of databases for the given connection
func GetDatabases(db *sql.DB, dbType string) ([]string, error) {
	var query string
	switch dbType {
	case "mysql":
		query = "SHOW DATABASES"
	case "postgresql":
		query = "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname"
	case "sqlite":
		return []string{"main"}, nil
	case "sqlserver":
		query = "SELECT name FROM sys.databases ORDER BY name"
	case "oracle":
		query = "SELECT USERNAME FROM ALL_USERS ORDER BY USERNAME"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		databases = append(databases, name)
	}
	return databases, rows.Err()
}

// GetTables returns list of tables for the given database
func GetTables(db *sql.DB, dbType string) ([]string, error) {
	var query string
	switch dbType {
	case "mysql":
		query = "SHOW TABLES"
	case "postgresql":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name"
	case "sqlite":
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	case "sqlserver":
		query = "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME"
	case "oracle":
		query = "SELECT TABLE_NAME FROM USER_TABLES ORDER BY TABLE_NAME"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     string `json:"nullable"`
	Key          string `json:"key"`
	DefaultValue *string `json:"defaultValue"`
	Extra        string `json:"extra"`
	Comment      string `json:"comment"`
}

// GetColumns returns column information for the given table
func GetColumns(db *sql.DB, dbType, tableName string) ([]ColumnInfo, error) {
	if err := ValidateIdentifier(tableName); err != nil {
		return nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = "SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA, COLUMN_COMMENT FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = '" + tableName + "' ORDER BY ORDINAL_POSITION"
	case "postgresql":
		return getColumnsPG(db, tableName)
	case "sqlite":
		query = "PRAGMA table_info('" + tableName + "')"
	case "sqlserver":
		return getColumnsSQLServer(db, tableName)
	case "oracle":
		return getColumnsOracle(db, tableName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Key, &col.DefaultValue, &col.Extra, &col.Comment); err != nil {
			continue
		}
		columns = append(columns, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// for SQLite, parse the PRAGMA result differently
	if dbType == "sqlite" && len(columns) == 0 {
		rows2, err := db.Query("PRAGMA table_info('" + tableName + "')")
		if err != nil {
			return nil, err
		}
		defer rows2.Close()
		for rows2.Next() {
			var cid int
			var name, ctype string
			var notnull int
			var dfltValue *string
			var pk int
			if err := rows2.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
				continue
			}
			nullable := "YES"
			if notnull == 1 {
				nullable = "NO"
			}
			key := ""
			if pk == 1 {
				key = "PRI"
			}
			columns = append(columns, ColumnInfo{
				Name:         name,
				Type:         ctype,
				Nullable:     nullable,
				Key:          key,
				DefaultValue: dfltValue,
			})
		}
		if err := rows2.Err(); err != nil {
			return nil, err
		}
	}

	return columns, nil
}

func getColumnsPG(db *sql.DB, tableName string) ([]ColumnInfo, error) {
	rows, err := db.Query(`
		SELECT column_name, data_type,
			CASE WHEN is_nullable = 'YES' THEN 'YES' ELSE 'NO' END,
			CASE WHEN column_name IN (
				SELECT column_name FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
				WHERE tc.table_name = $1 AND tc.constraint_type = 'PRIMARY KEY'
			) THEN 'PRI' ELSE '' END,
			column_default, '', ''
		FROM information_schema.columns WHERE table_name = $1 ORDER BY ordinal_position`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Key, &col.DefaultValue, &col.Extra, &col.Comment); err != nil {
			continue
		}
		columns = append(columns, col)
	}
	return columns, rows.Err()
}

func getColumnsSQLServer(db *sql.DB, tableName string) ([]ColumnInfo, error) {
	rows, err := db.Query(`
		SELECT c.COLUMN_NAME, c.DATA_TYPE,
			CASE WHEN c.IS_NULLABLE = 'YES' THEN 'YES' ELSE 'NO' END,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 'PRI' ELSE '' END,
			c.COLUMN_DEFAULT, '', ''
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN (
			SELECT ku.COLUMN_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
			WHERE ku.TABLE_NAME = @p1 AND tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
		) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
		WHERE c.TABLE_NAME = @p1 ORDER BY c.ORDINAL_POSITION`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Key, &col.DefaultValue, &col.Extra, &col.Comment); err != nil {
			continue
		}
		columns = append(columns, col)
	}
	return columns, rows.Err()
}

func getColumnsOracle(db *sql.DB, tableName string) ([]ColumnInfo, error) {
	rows, err := db.Query(`
		SELECT COLUMN_NAME, DATA_TYPE, NULLABLE,
			CASE WHEN COLUMN_NAME IN (
				SELECT COLUMN_NAME FROM USER_CONS_COLUMNS WHERE CONSTRAINT_NAME IN (
					SELECT CONSTRAINT_NAME FROM USER_CONSTRAINTS WHERE TABLE_NAME = :1 AND CONSTRAINT_TYPE = 'P'
				)
			) THEN 'PRI' ELSE '' END,
			DATA_DEFAULT, '', ''
		FROM USER_TAB_COLUMNS WHERE TABLE_NAME = :1 ORDER BY COLUMN_ID`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Key, &col.DefaultValue, &col.Extra, &col.Comment); err != nil {
			continue
		}
		columns = append(columns, col)
	}
	return columns, rows.Err()
}

type IndexInfo struct {
	Name      string   `json:"name"`
	Columns   []string `json:"columns"`
	Unique    bool     `json:"unique"`
	Primary   bool     `json:"primary"`
}

// GetIndexes returns index information for the given table
func GetIndexes(db *sql.DB, dbType, tableName string) ([]IndexInfo, error) {
	if err := ValidateIdentifier(tableName); err != nil {
		return nil, err
	}

	var query string
	switch dbType {
	case "mysql":
		query = "SHOW INDEX FROM `" + tableName + "`"
	case "postgresql":
		return getIndexesPG(db, tableName)
	case "sqlite":
		query = "PRAGMA index_list('" + tableName + "')"
	case "sqlserver":
		return getIndexesSQLServer(db, tableName)
	default:
		return nil, nil
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexMap := make(map[string]*IndexInfo)

	switch dbType {
	case "mysql":
		for rows.Next() {
			var table, nonUnique, seqInIndex int
			var keyName, columnName, collation string
			var cardinality, subPart, packed, null, indexType, comment, indexComment *string
			var visible *string
			if err := rows.Scan(&table, &nonUnique, &keyName, &seqInIndex, &columnName, &collation, &cardinality, &subPart, &packed, &null, &indexType, &comment, &visible, &indexComment); err != nil {
				continue
			}
			if idx, ok := indexMap[keyName]; ok {
				idx.Columns = append(idx.Columns, columnName)
			} else {
				idx := &IndexInfo{
					Name:    keyName,
					Columns: []string{columnName},
					Unique:  nonUnique == 0,
					Primary: keyName == "PRIMARY",
				}
				indexMap[keyName] = idx
			}
		}
	case "sqlite":
		for rows.Next() {
			var seq int
			var name string
			var unique int
			var origin string
			var partial int
			if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
				continue
			}
			// get columns for this index
			cols, _ := getIndexColumns(db, name)
			indexMap[name] = &IndexInfo{
				Name:    name,
				Columns: cols,
				Unique:  unique == 1,
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var indexes []IndexInfo
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}
	return indexes, nil
}

func getIndexesPG(db *sql.DB, tableName string) ([]IndexInfo, error) {
	rows, err := db.Query(`
		SELECT indexname, indexdef FROM pg_indexes
		WHERE tablename = $1 AND schemaname = 'public'`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var name, def string
		if err := rows.Scan(&name, &def); err != nil {
			continue
		}
		indexes = append(indexes, IndexInfo{Name: name, Columns: []string{}})
	}
	return indexes, rows.Err()
}

func getIndexesSQLServer(db *sql.DB, tableName string) ([]IndexInfo, error) {
	rows, err := db.Query(`
		SELECT i.name, i.is_unique, i.is_primary_key, c.name as col_name
		FROM sys.indexes i
		JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		WHERE i.object_id = OBJECT_ID(@p1) AND i.type > 0
		ORDER BY i.name, ic.key_ordinal`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexMap := make(map[string]*IndexInfo)
	for rows.Next() {
		var name string
		var isUnique, isPK bool
		var colName string
		if err := rows.Scan(&name, &isUnique, &isPK, &colName); err != nil {
			continue
		}
		if idx, ok := indexMap[name]; ok {
			idx.Columns = append(idx.Columns, colName)
		} else {
			indexMap[name] = &IndexInfo{
				Name:    name,
				Columns: []string{colName},
				Unique:  isUnique,
				Primary: isPK,
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var indexes []IndexInfo
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}
	return indexes, nil
}

func getIndexColumns(db *sql.DB, indexName string) ([]string, error) {
	rows, err := db.Query("PRAGMA index_info('" + indexName + "')")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var seqno, cid int
		var name string
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			continue
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

// GetDDL returns the DDL (CREATE TABLE statement) for the given table
func GetDDL(db *sql.DB, dbType, tableName string) (string, error) {
	if err := ValidateIdentifier(tableName); err != nil {
		return "", err
	}

	switch dbType {
	case "mysql":
		var table, createSQL string
		err := db.QueryRow("SHOW CREATE TABLE `" + tableName + "`").Scan(&table, &createSQL)
		if err != nil {
			return "", err
		}
		return createSQL, nil
	case "sqlite":
		var ddl string
		err := db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&ddl)
		if err != nil {
			return "", err
		}
		return ddl, nil
	case "postgresql":
		// Build DDL from columns
		cols, err := GetColumns(db, dbType, tableName)
		if err != nil {
			return "", err
		}
		var sb strings.Builder
		sb.WriteString("CREATE TABLE \"" + tableName + "\" (\n")
		for i, col := range cols {
			sb.WriteString("    \"" + col.Name + "\" " + col.Type)
			if col.Nullable == "NO" {
				sb.WriteString(" NOT NULL")
			}
			if col.DefaultValue != nil {
				sb.WriteString(" DEFAULT " + *col.DefaultValue)
			}
			if i < len(cols)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString(");")
		return sb.String(), nil
	case "sqlserver":
		cols, err := GetColumns(db, dbType, tableName)
		if err != nil {
			return "", err
		}
		var sb strings.Builder
		sb.WriteString("CREATE TABLE [" + tableName + "] (\n")
		for i, col := range cols {
			sb.WriteString("    [" + col.Name + "] " + col.Type)
			if col.Nullable == "NO" {
				sb.WriteString(" NOT NULL")
			}
			if col.DefaultValue != nil {
				sb.WriteString(" DEFAULT " + *col.DefaultValue)
			}
			if i < len(cols)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString(");")
		return sb.String(), nil
	default:
		return "-- DDL generation not supported for this database type", nil
	}
}
