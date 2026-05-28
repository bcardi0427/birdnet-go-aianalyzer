//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tphakala/birdnet-go/internal/classifier"
)

func main() {
	db, err := sql.Open("sqlite3", "birdnet.db")
	if err != nil {
		fmt.Printf("Error opening db: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, scientific_name FROM labels LIMIT 10")
	if err != nil {
		fmt.Printf("Error querying: %v\n", err)
		return
	}
	defer rows.Close()

	_, scientificIndex, err := classifier.LoadTaxonomyData("")
	if err != nil {
		fmt.Printf("LoadTaxonomyData error: %v\n", err)
	}

	taxonomyCodeMap := make(map[string]string, len(scientificIndex))
	for sciName, code := range scientificIndex {
		taxonomyCodeMap[strings.ToLower(sciName)] = code
	}

	for rows.Next() {
		var id int
		var label string
		if err := rows.Scan(&id, &label); err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}

		sci, rest, found := strings.Cut(label, "_")
		if !found {
			sci = label
		}

		code := taxonomyCodeMap[strings.ToLower(sci)]
		fmt.Printf("LabelID: %d | DB Label: %q | sci: %q | eBird Code: %q | rest: %q\n", id, label, sci, code, rest)
	}
}
