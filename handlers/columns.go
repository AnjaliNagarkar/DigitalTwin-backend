package handlers

import (
	"database/sql"
	"fmt"
	"strings"
)

// ColumnChecker detects which columns exist in the FAMILY table at startup.
type ColumnChecker struct {
	available map[string]bool
	LatCol    string // auto-detected latitude column name
	LngCol    string // auto-detected longitude column name
}

func NewColumnChecker(db *sql.DB) *ColumnChecker {
	cc := &ColumnChecker{available: make(map[string]bool)}

	rows, err := db.Query(`
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'FAMILY'
		ORDER BY ORDINAL_POSITION
	`)
	if err != nil {
		fmt.Printf("Warning: could not read FAMILY columns: %v\n", err)
		return cc
	}
	defer rows.Close()

	var allCols []string
	for rows.Next() {
		var col string
		rows.Scan(&col)
		cc.available[strings.ToUpper(col)] = true
		allCols = append(allCols, col)
	}

	fmt.Printf("Detected %d columns in FAMILY table\n", len(allCols))

	// Auto-detect lat/lng columns:
	// 1. Prefer the exact LATITUDE/LONGITUDE column names when present.
	for _, col := range allCols {
		up := strings.ToUpper(col)
		if up == "LATITUDE" {
			cc.LatCol = col
		}
		if up == "LONGITUDE" {
			cc.LngCol = col
		}
	}

	// 2. Search by common name patterns if exact names are not present.
	for _, col := range allCols {
		up := strings.ToUpper(col)
		if cc.LatCol == "" && strings.Contains(up, "LAT") && !strings.Contains(up, "PLAT") {
			cc.LatCol = col
		}
		if cc.LngCol == "" && (strings.HasPrefix(up, "LON") || strings.HasPrefix(up, "LNG") || strings.HasSuffix(up, "LON")) {
			cc.LngCol = col
		}
	}

	// 3. Fall back to the last 2 columns in ordinal order
	if (cc.LatCol == "" || cc.LngCol == "") && len(allCols) >= 2 {
		cc.LatCol = allCols[len(allCols)-2]
		cc.LngCol = allCols[len(allCols)-1]
		fmt.Printf("Lat/Lng not found by name — using last 2 columns: %s, %s\n", cc.LatCol, cc.LngCol)
	}

	if cc.LatCol != "" {
		fmt.Printf("Using latitude column: %s\n", cc.LatCol)
	}
	if cc.LngCol != "" {
		fmt.Printf("Using longitude column: %s\n", cc.LngCol)
	}

	return cc
}

func (cc *ColumnChecker) Has(col string) bool {
	return cc.available[strings.ToUpper(col)]
}

// ColOrEmpty returns `COALESCE(col, ”)` if column exists, or `” AS alias` if not.
func (cc *ColumnChecker) ColOrEmpty(col, alias string) string {
	if cc.Has(col) {
		return fmt.Sprintf("COALESCE(f.%s, '')", col)
	}
	return fmt.Sprintf("'' /* %s not found */", alias)
}
