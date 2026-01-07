package testablecode

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseCSV parses a CSV file with page rankings and URLs.
// Supports both header and headerless formats:
//   - With header: rank,url (first row contains column names)
//   - Without header: 1,www.mongodb.com/docs/... (first row is data)
// Returns a slice of PageEntry structs.
func ParseCSV(path string) ([]PageEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Determine if first row is a header or data
	// Check if first column of first row is a number (data) or text (header)
	firstRow := records[0]
	if len(firstRow) < 2 {
		return nil, fmt.Errorf("CSV must have at least 2 columns (rank and URL)")
	}

	hasHeader := false
	rankIdx := 0
	urlIdx := 1

	// Try to parse first column as a number
	_, err = strconv.Atoi(strings.TrimSpace(firstRow[0]))
	if err != nil {
		// First column is not a number, so this is likely a header row
		hasHeader = true

		// Find column indices from header
		for i, col := range firstRow {
			colLower := strings.ToLower(strings.TrimSpace(col))
			switch colLower {
			case "rank", "site rank", "siterank":
				rankIdx = i
			case "url", "page", "path":
				urlIdx = i
			}
		}
	}

	// Determine starting row index
	startIdx := 0
	if hasHeader {
		startIdx = 1
	}

	if len(records) <= startIdx {
		return nil, fmt.Errorf("no data rows found in CSV")
	}

	// Parse data rows
	var entries []PageEntry
	for i, record := range records[startIdx:] {
		if len(record) <= rankIdx || len(record) <= urlIdx {
			continue // Skip malformed rows
		}

		rankStr := strings.TrimSpace(record[rankIdx])
		url := strings.TrimSpace(record[urlIdx])

		if rankStr == "" || url == "" {
			continue // Skip empty rows
		}

		rank, err := strconv.Atoi(rankStr)
		if err != nil {
			// Try to parse as float and convert
			rankFloat, err := strconv.ParseFloat(rankStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid rank value on row %d: %s", i+startIdx+1, rankStr)
			}
			rank = int(rankFloat)
		}

		entries = append(entries, PageEntry{
			Rank: rank,
			URL:  url,
		})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no valid data rows found in CSV")
	}

	return entries, nil
}

