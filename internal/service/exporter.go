package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

func ExportToCSV(result *QueryResult) ([]byte, error) {
	var sb strings.Builder
	w := csv.NewWriter(&sb)

	// write header
	if err := w.Write(result.Columns); err != nil {
		return nil, err
	}

	// write rows
	for _, row := range result.Rows {
		record := make([]string, len(row))
		for i, v := range row {
			if v == nil {
				record[i] = "NULL"
			} else {
				record[i] = fmt.Sprintf("%v", v)
			}
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return []byte(sb.String()), w.Error()
}

func ExportToJSON(result *QueryResult) ([]byte, error) {
	rows := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		m := make(map[string]interface{})
		for j, col := range result.Columns {
			if j < len(row) {
				m[col] = row[j]
			}
		}
		rows[i] = m
	}
	return json.MarshalIndent(rows, "", "  ")
}

func ExportToSQL(result *QueryResult, tableName string) ([]byte, error) {
	var sb strings.Builder
	for _, row := range result.Rows {
		var cols []string
		var vals []string
		for j, col := range result.Columns {
			if j < len(row) {
				cols = append(cols, fmt.Sprintf("`%s`", col))
				if row[j] == nil {
					vals = append(vals, "NULL")
				} else {
					switch v := row[j].(type) {
					case string:
						vals = append(vals, fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")))
					case float64, float32, int, int64, int32:
						vals = append(vals, fmt.Sprintf("%v", v))
					default:
						vals = append(vals, fmt.Sprintf("'%v'", v))
					}
				}
			}
		}
		sb.WriteString(fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s);\n",
			tableName, strings.Join(cols, ", "), strings.Join(vals, ", ")))
	}
	return []byte(sb.String()), nil
}

func ExportToMarkdown(result *QueryResult) ([]byte, error) {
	var sb strings.Builder

	// header
	sb.WriteString("| " + strings.Join(result.Columns, " | ") + " |\n")
	seps := make([]string, len(result.Columns))
	for i := range seps {
		seps[i] = "---"
	}
	sb.WriteString("| " + strings.Join(seps, " | ") + " |\n")

	// rows
	for _, row := range result.Rows {
		record := make([]string, len(row))
		for i, v := range row {
			if v == nil {
				record[i] = "NULL"
			} else {
				record[i] = fmt.Sprintf("%v", v)
			}
		}
		sb.WriteString("| " + strings.Join(record, " | ") + " |\n")
	}
	return []byte(sb.String()), nil
}
