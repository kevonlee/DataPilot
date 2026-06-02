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

// poolKey returns a unique key for the connection + database combination.
func poolKey(connID, dbName string) string {
	if dbName == "" {
		return connID
	}
	return connID + "/" + dbName
}

func (m *DBManager) Get(conn *model.Connection) (*sql.DB, error) {
	key := poolKey(conn.ID, conn.Database)

	m.mu.RLock()
	db, ok := m.pools[key]
	m.mu.RUnlock()

	if ok {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		// stale connection - remove it
		m.mu.Lock()
		// re-check under write lock
		if m.pools[key] == db {
			db.Close()
			delete(m.pools, key)
		}
		m.mu.Unlock()
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

	key := poolKey(conn.ID, conn.Database)
	m.mu.Lock()
	// close any existing connection for this key
	if old, ok := m.pools[key]; ok {
		old.Close()
	}
	m.pools[key] = db
	m.mu.Unlock()

	return db, nil
}

func (m *DBManager) Close(connID string) {
	m.mu.Lock()
	for key, db := range m.pools {
		if key == connID || (len(key) > len(connID)+1 && key[:len(connID)+1] == connID+"/") {
			db.Close()
			delete(m.pools, key)
		}
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

// SwitchDatabase returns a connection for the specified database.
// Uses per-database pool keys so connections are not shared across databases.
func (m *DBManager) SwitchDatabase(conn *model.Connection, dbName string) (*sql.DB, error) {
	// Create a copy with the target database
	switchConn := *conn
	switchConn.Database = dbName
	return m.Get(&switchConn)
}
