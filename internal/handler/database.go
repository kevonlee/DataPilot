package handler

import (
	"database-manager/internal/config"
	"database-manager/internal/model"
	"database-manager/internal/service"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func handleDBRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/conn/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}

	connID := parts[0]
	cfg := config.Get()
	conn := cfg.GetConnection(connID)
	if conn == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
		return
	}

	dbm := service.GetDBManager()

	// /api/conn/{id}/databases
	if len(parts) == 2 && parts[1] == "databases" {
		db, err := dbm.Get(conn)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		databases, err := service.GetDatabases(db, string(conn.Type))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, databases)
		return
	}

	// /api/conn/{id}/query
	if len(parts) >= 2 && parts[1] == "query" {
		handleQueryRoute(w, r, conn)
		return
	}

	if len(parts) < 4 {
		http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
		return
	}

	dbName := parts[2]

	db, err := dbm.SwitchDatabase(conn, dbName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// /api/conn/{id}/databases/{db}/tables
	if len(parts) == 3 || (len(parts) == 4 && parts[3] == "tables") {
		tables, err := service.GetTables(db, string(conn.Type))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, tables)
		return
	}

	if len(parts) < 5 {
		http.Error(w, `{"error":"missing table name"}`, http.StatusBadRequest)
		return
	}

	tableName := parts[4]

	if len(parts) >= 6 {
		action := parts[5]
		switch action {
		case "columns":
			columns, err := service.GetColumns(db, string(conn.Type), tableName)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, columns)

		case "indexes":
			indexes, err := service.GetIndexes(db, string(conn.Type), tableName)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, indexes)

		case "ddl":
			ddl, err := service.GetDDL(db, string(conn.Type), tableName)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"ddl": ddl})

		case "data":
			handleTableDataRoute(w, r, db, tableName, conn.Type)

		case "export":
			handleExportRoute(w, r, db, tableName, conn.Type)

		default:
			http.Error(w, `{"error":"unknown action"}`, http.StatusBadRequest)
		}
		return
	}

	http.Error(w, `{"error":"invalid path"}`, http.StatusBadRequest)
}

func handleTableDataRoute(w http.ResponseWriter, r *http.Request, db *sql.DB, tableName string, dbType model.DBType) {
	switch r.Method {
	case http.MethodGet:
		page := 1
		pageSize := 50
		if p := r.URL.Query().Get("page"); p != "" {
			json.Unmarshal([]byte(p), &page)
		}
		if ps := r.URL.Query().Get("pageSize"); ps != "" {
			json.Unmarshal([]byte(ps), &pageSize)
		}

		offset := (page - 1) * pageSize
		query := "SELECT * FROM `" + tableName + "` LIMIT " + itoa(pageSize) + " OFFSET " + itoa(offset)
		result := service.ExecuteQuery(db, query)

		// get total count
		countResult := service.ExecuteQuery(db, "SELECT COUNT(*) FROM `"+tableName+"`")
		total := int64(0)
		if countResult.Error == "" && len(countResult.Rows) > 0 {
			if v, ok := countResult.Rows[0][0].(int64); ok {
				total = v
			}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data":  result,
			"total": total,
			"page":  page,
			"size":  pageSize,
		})

	case http.MethodPost:
		handleInsertRow(w, r, db, tableName, dbType)

	case http.MethodPut:
		handleUpdateRow(w, r, db, tableName, dbType)

	case http.MethodDelete:
		handleDeleteRow(w, r, db, tableName, dbType)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleInsertRow(w http.ResponseWriter, r *http.Request, db *sql.DB, tableName string, dbType model.DBType) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	var cols []string
	var vals []interface{}
	var placeholders []string
	i := 0
	for k, v := range req {
		cols = append(cols, "`"+k+"`")
		vals = append(vals, v)
		if dbType == model.DBTypePostgreSQL {
			placeholders = append(placeholders, "$"+itoa(i+1))
		} else {
			placeholders = append(placeholders, "?")
		}
		i++
	}

	query := "INSERT INTO `" + tableName + "` (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	_, err := db.Exec(query, vals...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "inserted"})
}

func handleUpdateRow(w http.ResponseWriter, r *http.Request, db *sql.DB, tableName string, dbType model.DBType) {
	var req struct {
		Where map[string]interface{} `json:"where"`
		Set   map[string]interface{} `json:"set"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	var setClauses []string
	var args []interface{}
	i := 0
	for k, v := range req.Set {
		if dbType == model.DBTypePostgreSQL {
			setClauses = append(setClauses, "`"+k+"` = $"+itoa(i+1))
		} else {
			setClauses = append(setClauses, "`"+k+"` = ?")
		}
		args = append(args, v)
		i++
	}

	var whereClauses []string
	for k, v := range req.Where {
		if dbType == model.DBTypePostgreSQL {
			whereClauses = append(whereClauses, "`"+k+"` = $"+itoa(i+1))
		} else {
			whereClauses = append(whereClauses, "`"+k+"` = ?")
		}
		args = append(args, v)
		i++
	}

	query := "UPDATE `" + tableName + "` SET " + strings.Join(setClauses, ", ") + " WHERE " + strings.Join(whereClauses, " AND ")
	_, err := db.Exec(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func handleDeleteRow(w http.ResponseWriter, r *http.Request, db *sql.DB, tableName string, dbType model.DBType) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	var whereClauses []string
	var args []interface{}
	i := 0
	for k, v := range req {
		if dbType == model.DBTypePostgreSQL {
			whereClauses = append(whereClauses, "`"+k+"` = $"+itoa(i+1))
		} else {
			whereClauses = append(whereClauses, "`"+k+"` = ?")
		}
		args = append(args, v)
		i++
	}

	query := "DELETE FROM `" + tableName + "` WHERE " + strings.Join(whereClauses, " AND ")
	_, err := db.Exec(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func handleExportRoute(w http.ResponseWriter, r *http.Request, db *sql.DB, tableName string, dbType model.DBType) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Format string `json:"format"` // csv, json, sql
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	query := "SELECT * FROM `" + tableName + "`"
	result := service.ExecuteQuery(db, query)
	if result.Error != "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": result.Error})
		return
	}

	var data []byte
	var err error
	var contentType string
	var filename string

	switch req.Format {
	case "csv":
		data, err = service.ExportToCSV(result)
		contentType = "text/csv"
		filename = tableName + ".csv"
	case "json":
		data, err = service.ExportToJSON(result)
		contentType = "application/json"
		filename = tableName + ".json"
	case "sql":
		data, err = service.ExportToSQL(result, tableName)
		contentType = "text/plain"
		filename = tableName + ".sql"
	default:
		data, err = service.ExportToCSV(result)
		contentType = "text/csv"
		filename = tableName + ".csv"
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(data)
}

func handleQueryRoute(w http.ResponseWriter, r *http.Request, conn *model.Connection) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SQL      string `json:"sql"`
		Database string `json:"database"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.SQL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty query"})
		return
	}

	dbm := service.GetDBManager()
	var db *sql.DB
	var err error

	if req.Database != "" {
		db, err = dbm.SwitchDatabase(conn, req.Database)
	} else {
		db, err = dbm.Get(conn)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := service.ExecuteQuery(db, req.SQL)

	// save history
	cfg := config.Get()
	cfg.AddHistory(model.QueryHistory{
		ID:         generateID(),
		ConnID:     conn.ID,
		Database:   req.Database,
		SQL:        req.SQL,
		Duration:   result.Duration,
		ExecutedAt: timeNow(),
	})

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
