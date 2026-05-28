//go:build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/tphakala/birdnet-go/internal/classifier"
)

func main() {
	_, scientificIndex, err := classifier.LoadTaxonomyData("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	taxonomyCodeMap := make(map[string]string, len(scientificIndex))
	for sciName, code := range scientificIndex {
		taxonomyCodeMap[strings.ToLower(sciName)] = code
	}

	// simulate what happens in the app
	labelsByID := map[uint]string{
		1: "Cardinalis cardinalis_Northern Cardinal",
		2: "Cardinalis cardinalis",
		3: "Spizella pallida_Clay-colored Sparrow",
		4: "Pandion haliaetus_Osprey",
	}

	for _, label := range labelsByID {
		sci, _, found := strings.Cut(label, "_")
		if !found {
			sci = label
		}

		labels := []string{"Cardinalis cardinalis_Northern Cardinal", "Spizella pallida_Clay-colored Sparrow", "Pandion haliaetus_Osprey"}
		
		var commonName, ebirdCode string
		for _, l := range labels {
			lsci, lrest, lfound := strings.Cut(l, "_")
			if lfound && lsci == sci {
				if idx := strings.LastIndex(lrest, "_"); idx > 0 {
					commonName = lrest[:idx]
					ebirdCode = lrest[idx+1:]
				} else {
					commonName = lrest
					ebirdCode = ""
				}
				break
			}
		}
		if commonName == "" {
			commonName = sci
			ebirdCode = ""
		}

		if ebirdCode == "" && taxonomyCodeMap != nil {
			ebirdCode = taxonomyCodeMap[strings.ToLower(sci)]
		}

		fmt.Printf("Label: %q -> sci: %q -> eBirdCode: %q (common: %q)\n", label, sci, ebirdCode, commonName)
	}
}
