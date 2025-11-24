package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// sum CSV columns across all CSV files in a directory.
// Usage examples:
//  go run ./tools/sum_csv_columns.go -dir ./result/pbft_6 -cols 4
//  go run ./tools/sum_csv_columns.go -dir ./result/pbft_6 -cols 4,5,6

func parseCols(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	cols := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		if v <= 0 {
			return nil, fmt.Errorf("columns are 1-based and must be > 0")
		}
		cols = append(cols, v-1) // convert to 0-based index
	}
	return cols, nil
}

func sumFileColumns(path string, cols []int) ([]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	sums := make([]float64, len(cols))

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// skip malformed lines but log
			log.Printf("warning: reading %s: %v", path, err)
			continue
		}
		for i, cidx := range cols {
			if cidx < 0 || cidx >= len(rec) {
				continue
			}
			cell := strings.TrimSpace(rec[cidx])
			if cell == "" {
				continue
			}
			// allow integers or floats
			v, err := strconv.ParseFloat(cell, 64)
			if err != nil {
				// try to remove quotes or surrounding characters
				cleaned := strings.Trim(cell, "\"' ")
				v, err = strconv.ParseFloat(cleaned, 64)
				if err != nil {
					// if can't parse, skip this cell
					continue
				}
			}
			sums[i] += v
		}
	}
	return sums, nil
}

func main() {
	dir := flag.String("dir", "./result/pbft_6", "directory containing CSV files")
	colsArg := flag.String("cols", "4", "comma-separated 1-based column indices to sum (e.g. 4 or 4,5,6)")
	pattern := flag.String("pattern", "*.csv", "filename glob pattern to match CSVs")
	flag.Parse()

	cols, err := parseCols(*colsArg)
	if err != nil {
		log.Fatalf("invalid cols: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(*dir, *pattern))
	if err != nil {
		log.Fatalf("failed to list files: %v", err)
	}
	if len(files) == 0 {
		log.Fatalf("no files found in %s matching %s", *dir, *pattern)
	}

	total := make([]float64, len(cols))
	fmt.Printf("Found %d files in %s\n", len(files), *dir)

	for _, f := range files {
		sums, err := sumFileColumns(f, cols)
		if err != nil {
			log.Printf("error processing %s: %v", f, err)
			continue
		}
		fmt.Printf("%s:\n", f)
		for i, s := range sums {
			fmt.Printf("  col %d sum = %.0f\n", cols[i]+1, s)
			total[i] += s
		}
	}

	fmt.Println("Overall totals:")
	for i, s := range total {
		// print integer-like values without decimal when possible
		if s == float64(int64(s)) {
			fmt.Printf("  col %d total = %d\n", cols[i]+1, int64(s))
		} else {
			fmt.Printf("  col %d total = %.6f\n", cols[i]+1, s)
		}
	}

	// example: also print sum of all selected columns combined
	grand := 0.0
	for _, v := range total {
		grand += v
	}
	if grand == float64(int64(grand)) {
		fmt.Printf("Grand total (all selected columns combined) = %d\n", int64(grand))
	} else {
		fmt.Printf("Grand total (all selected columns combined) = %.6f\n", grand)
	}
}
