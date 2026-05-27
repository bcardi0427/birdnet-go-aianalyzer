//go:build ignore

package main

import (
	"fmt"
	"github.com/tphakala/birdnet-go/internal/classifier"
)

func main() {
	tax, sci, err := classifier.LoadTaxonomyData("")
	if err != nil {
		fmt.Printf("Error loading taxonomy data: %v\n", err)
		return
	}
	fmt.Printf("Tax map size: %d\n", len(tax))
	fmt.Printf("Sci map size: %d\n", len(sci))
	fmt.Printf("Cardinalis cardinalis: %s\n", sci["Cardinalis cardinalis"])
}
