package service

import (
	"fmt"
	"regexp"
)

var safeIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidateIdentifier checks that a table or column name contains only safe characters.
func ValidateIdentifier(name string) error {
	if !safeIdentifier.MatchString(name) {
		return fmt.Errorf("invalid identifier: %q", name)
	}
	return nil
}

// QuoteIdentifier quotes a table/column name for the given database type.
func QuoteIdentifier(name, dbType string) string {
	if err := ValidateIdentifier(name); err != nil {
		return name // will fail validation elsewhere
	}
	switch dbType {
	case "postgresql":
		return `"` + name + `"`
	default:
		return "`" + name + "`"
	}
}
