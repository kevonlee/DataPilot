package service

import (
	"database-manager/internal/model"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/sijms/go-ora/v2"
)

type DBManager struct {
	mu    sync.RWMutex
	pools map[string]*sql.DB
}

var dbManager *DBManager

func init() {
	dbManager = &DBManager{
		pools: make(map[string]*sql.DB),
	}
}

func GetDBManager() *DBManager {
	return dbManager
}

func (m *DBManager) Get(conn *model.Connection) (*sql.DB, error) {
	m.mu.RLock()
	if db, ok := m.pools[conn.ID]; ok {
		m.mu.RUnlock()
		if err := db.Ping(); err == nil {
			return db, nil
		}
		// connection stale, remove and reconnect
		m.mu.Lock()
		delete(m.pools, conn.ID)
		m.mu.Unlock()
	} else {
		m.mu.RUnlock()
	}

	return m.connect(conn)
}

func (m *DBManager) connect(conn *model.Connection) (*sql.DB, error) {
	dsn := conn.DSN()
	if dsn == "" {
		return nil, fmt.Errorf("unsupported database type: %s", conn.Type)
	}

	driverName := string(conn.Type)
	if conn.Type == model.DBTypeSQLServer {
		driverName = "sqlserver"
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	m.mu.Lock()
	m.pools[conn.ID] = db
	m.mu.Unlock()

	return db, nil
}

func (m *DBManager) Close(connID string) {
	m.mu.Lock()
	if db, ok := m.pools[connID]; ok {
		db.Close()
		delete(m.pools, connID)
	}
	m.mu.Unlock()
}

func (m *DBManager) CloseAll() {
	m.mu.Lock()
	for id, db := range m.pools {
		db.Close()
		delete(m.pools, id)
	}
	m.mu.Unlock()
}

func (m *DBManager) Test(conn *model.Connection) error {
	dsn := conn.DSN()
	if dsn == "" {
		return fmt.Errorf("unsupported database type: %s", conn.Type)
	}

	driverName := string(conn.Type)
	if conn.Type == model.DBTypeSQLServer {
		driverName = "sqlserver"
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Ping()
}

// SwitchDatabase switches the connection to a different database
func (m *DBManager) SwitchDatabase(conn *model.Connection, dbName string) (*sql.DB, error) {
	db, err := m.Get(conn)
	if err != nil {
		return nil, err
	}

	// For SQLite, no need to switch
	if conn.Type == model.DBTypeSQLite {
		return db, nil
	}

	// Execute USE database command for SQL databases
	var useSQL string
	switch conn.Type {
	case model.DBTypeMySQL, model.DBTypeSQLServer:
		useSQL = "USE `" + dbName + "`"
	case model.DBTypePostgreSQL:
		// PostgreSQL doesn't support USE, need new connection
		newConn := *conn
		newConn.Database = dbName
		return m.connect(&newConn)
	case model.DBTypeOracle:
		// Oracle uses schemas, not USE
		return db, nil
	default:
		useSQL = "USE `" + dbName + "`"
	}

	if useSQL != "" {
		if _, err := db.Exec(useSQL); err != nil {
			return nil, fmt.Errorf("switch database failed: %w", err)
		}
	}

	return db, nil
}
