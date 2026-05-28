//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type TaxonomyMap map[string]string

func main() {
	data, err := os.ReadFile("internal/classifier/data/eBird_taxonomy_codes_2021E.json")
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		return
	}
	var taxonomyMap TaxonomyMap
	if err := json.Unmarshal(data, &taxonomyMap); err != nil {
		fmt.Printf("Err: %v\n", err)
		return
	}

	index := make(map[string]string)
	for taxonName, taxonCode := range taxonomyMap {
		parts := strings.Split(taxonName, "_")
		sci := taxonName
		if len(parts) == 2 {
			sci = parts[0]
		}
		index[strings.ToLower(sci)] = taxonCode
	}

	fmt.Printf("Size: %d\n", len(index))
	fmt.Printf("Cardinalis cardinalis: %q\n", index["cardinalis cardinalis"])
	fmt.Printf("Dryophytes cinereus: %q\n", index["dryophytes cinereus"])
	fmt.Printf("Buteo lineatus: %q\n", index["buteo lineatus"])
	fmt.Printf("Spizella pallida: %q\n", index["spizella pallida"])
}
