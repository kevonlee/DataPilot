package model

import "time"

type DBType string

const (
	DBTypeMySQL      DBType = "mysql"
	DBTypePostgreSQL DBType = "postgresql"
	DBTypeSQLite     DBType = "sqlite"
	DBTypeSQLServer  DBType = "sqlserver"
	DBTypeOracle     DBType = "oracle"
)

type Connection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      DBType `json:"type"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Database  string `json:"database"`
	FileName  string `json:"fileName"` // for SQLite
	Group     string `json:"group"`
	Color     string `json:"color"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (c *Connection) DSN() string {
	switch c.Type {
	case DBTypeMySQL:
		return c.Username + ":" + c.Password + "@tcp(" + c.Host + ":" + itoa(c.Port) + ")/" + c.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	case DBTypePostgreSQL:
		return "host=" + c.Host + " port=" + itoa(c.Port) + " user=" + c.Username + " password=" + c.Password + " dbname=" + c.Database + " sslmode=disable"
	case DBTypeSQLite:
		return c.FileName
	case DBTypeSQLServer:
		return "server=" + c.Host + ";port=" + itoa(c.Port) + ";user id=" + c.Username + ";password=" + c.Password + ";database=" + c.Database
	case DBTypeOracle:
		return "oracle://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + itoa(c.Port) + "/" + c.Database
	default:
		return ""
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

type QueryHistory struct {
	ID         string    `json:"id"`
	ConnID     string    `json:"connId"`
	Database   string    `json:"database"`
	SQL        string    `json:"sql"`
	Rows       int       `json:"rows"`
	Duration   int64     `json:"duration"` // ms
	ExecutedAt time.Time `json:"executedAt"`
}
