package testablecode

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// BuildPageReport builds a PageReport from a PageAnalysis.
func BuildPageReport(analysis *PageAnalysis) PageReport {
	report := PageReport{
		Rank:       analysis.Rank,
		URL:        analysis.URL,
		SourcePath: analysis.SourcePath,
		ContentDir: analysis.ContentDir,
		Error:      analysis.Error,
		ByProduct:  make(map[string]*ProductStats),
	}

	for _, ex := range analysis.CodeExamples {
		report.TotalExamples++
		if ex.IsInput {
			report.TotalInput++
		}
		if ex.IsOutput {
			report.TotalOutput++
		}
		if ex.IsTested {
			report.TotalTested++
		}
		if ex.IsTestable {
			report.TotalTestable++
		}
		if ex.IsMaybeTestable {
			report.TotalMaybeTestable++
		}

		// Aggregate by product
		product := ex.Product
		if product == "" {
			product = "Unknown"
		}
		stats, ok := report.ByProduct[product]
		if !ok {
			stats = &ProductStats{Product: product}
			report.ByProduct[product] = stats
		}
		stats.TotalCount++
		if ex.IsInput {
			stats.InputCount++
		}
		if ex.IsOutput {
			stats.OutputCount++
		}
		if ex.IsTested {
			stats.TestedCount++
		}
		if ex.IsTestable {
			stats.TestableCount++
		}
		if ex.IsMaybeTestable {
			stats.MaybeTestableCount++
		}
	}

	return report
}

// OutputText outputs the reports in text format.
func OutputText(w io.Writer, reports []PageReport) error {
	fmt.Fprintln(w, "="+strings.Repeat("=", 89))
	fmt.Fprintln(w, "PAGE ANALYTICS REPORT")
	fmt.Fprintln(w, "="+strings.Repeat("=", 89))
	fmt.Fprintf(w, "Total pages analyzed: %d\n\n", len(reports))

	// Summary table
	fmt.Fprintln(w, "SUMMARY")
	fmt.Fprintln(w, "-"+strings.Repeat("-", 89))
	fmt.Fprintf(w, "%-5s %-50s %6s %6s %8s %6s\n", "Rank", "URL", "Total", "Tested", "Testable", "Maybe")
	fmt.Fprintln(w, "-"+strings.Repeat("-", 89))

	for _, report := range reports {
		url := report.URL
		if len(url) > 50 {
			url = url[:47] + "..."
		}
		if report.Error != "" {
			fmt.Fprintf(w, "%-5d %-50s %s\n", report.Rank, url, "ERROR: "+report.Error)
		} else {
			fmt.Fprintf(w, "%-5d %-50s %6d %6d %8d %6d\n",
				report.Rank, url, report.TotalExamples, report.TotalTested,
				report.TotalTestable, report.TotalMaybeTestable)
		}
	}
	fmt.Fprintln(w)

	// Detailed per-page reports
	fmt.Fprintln(w, "DETAILED REPORTS")
	fmt.Fprintln(w, "="+strings.Repeat("=", 89))

	for _, report := range reports {
		if report.Error != "" {
			continue
		}

		fmt.Fprintf(w, "\nRank %d: %s\n", report.Rank, report.URL)
		fmt.Fprintf(w, "Source: %s\n", report.SourcePath)
		fmt.Fprintln(w, "-"+strings.Repeat("-", 89))

		if len(report.ByProduct) == 0 {
			fmt.Fprintln(w, "  No code examples found")
			continue
		}

		// Sort products for consistent output
		products := make([]string, 0, len(report.ByProduct))
		for p := range report.ByProduct {
			products = append(products, p)
		}
		sort.Strings(products)

		fmt.Fprintf(w, "  %-20s %6s %6s %6s %6s %8s %6s\n",
			"Product", "Total", "Input", "Output", "Tested", "Testable", "Maybe")
		fmt.Fprintln(w, "  "+strings.Repeat("-", 68))

		for _, product := range products {
			stats := report.ByProduct[product]
			fmt.Fprintf(w, "  %-20s %6d %6d %6d %6d %8d %6d\n",
				product, stats.TotalCount, stats.InputCount, stats.OutputCount,
				stats.TestedCount, stats.TestableCount, stats.MaybeTestableCount)
		}

		fmt.Fprintf(w, "  %s\n", strings.Repeat("-", 68))
		fmt.Fprintf(w, "  %-20s %6d %6d %6d %6d %8d %6d\n",
			"TOTAL", report.TotalExamples, report.TotalInput, report.TotalOutput,
			report.TotalTested, report.TotalTestable, report.TotalMaybeTestable)
	}

	return nil
}

// OutputJSON outputs the reports in JSON format.
func OutputJSON(w io.Writer, reports []PageReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(reports)
}

// OutputCSV outputs the reports in CSV format.
// If showDetails is false, outputs one row per page (summary).
// If showDetails is true, outputs one row per product per page (only products with non-zero values).
func OutputCSV(w io.Writer, reports []PageReport, showDetails bool) error {
	if showDetails {
		return outputCSVDetails(w, reports)
	}
	return outputCSVSummary(w, reports)
}

// outputCSVSummary outputs one row per page with aggregate stats.
func outputCSVSummary(w io.Writer, reports []PageReport) error {
	// Header
	fmt.Fprintln(w, "Rank,URL,SourcePath,ContentDir,Total,Input,Output,Tested,Testable,Maybe,Error")

	for _, report := range reports {
		// Escape fields that might contain commas or quotes
		url := escapeCSV(report.URL)
		sourcePath := escapeCSV(report.SourcePath)
		contentDir := escapeCSV(report.ContentDir)
		errorMsg := escapeCSV(report.Error)

		fmt.Fprintf(w, "%d,%s,%s,%s,%d,%d,%d,%d,%d,%d,%s\n",
			report.Rank, url, sourcePath, contentDir,
			report.TotalExamples, report.TotalInput, report.TotalOutput,
			report.TotalTested, report.TotalTestable, report.TotalMaybeTestable,
			errorMsg)
	}

	return nil
}

// outputCSVDetails outputs one row per product per page.
// Only includes products where at least one column has a non-zero value.
func outputCSVDetails(w io.Writer, reports []PageReport) error {
	// Header
	fmt.Fprintln(w, "Rank,URL,SourcePath,ContentDir,Product,Total,Input,Output,Tested,Testable,Maybe,Error")

	for _, report := range reports {
		// Escape fields that might contain commas or quotes
		url := escapeCSV(report.URL)
		sourcePath := escapeCSV(report.SourcePath)
		contentDir := escapeCSV(report.ContentDir)
		errorMsg := escapeCSV(report.Error)

		if report.Error != "" {
			// For error rows, output a single row with the error
			fmt.Fprintf(w, "%d,%s,%s,%s,,%d,%d,%d,%d,%d,%d,%s\n",
				report.Rank, url, sourcePath, contentDir,
				report.TotalExamples, report.TotalInput, report.TotalOutput,
				report.TotalTested, report.TotalTestable, report.TotalMaybeTestable,
				errorMsg)
			continue
		}

		if len(report.ByProduct) == 0 {
			// No code examples - output a single row with zeros
			fmt.Fprintf(w, "%d,%s,%s,%s,,%d,%d,%d,%d,%d,%d,\n",
				report.Rank, url, sourcePath, contentDir,
				0, 0, 0, 0, 0, 0)
			continue
		}

		// Sort products for consistent output
		products := make([]string, 0, len(report.ByProduct))
		for p := range report.ByProduct {
			products = append(products, p)
		}
		sort.Strings(products)

		for _, product := range products {
			stats := report.ByProduct[product]

			// Skip products where all columns are zero
			if stats.TotalCount == 0 && stats.InputCount == 0 && stats.OutputCount == 0 &&
				stats.TestedCount == 0 && stats.TestableCount == 0 && stats.MaybeTestableCount == 0 {
				continue
			}

			productEscaped := escapeCSV(product)
			fmt.Fprintf(w, "%d,%s,%s,%s,%s,%d,%d,%d,%d,%d,%d,\n",
				report.Rank, url, sourcePath, contentDir, productEscaped,
				stats.TotalCount, stats.InputCount, stats.OutputCount,
				stats.TestedCount, stats.TestableCount, stats.MaybeTestableCount)
		}
	}

	return nil
}

// escapeCSV escapes a string for CSV output.
// If the string contains commas, quotes, or newlines, it wraps in quotes and escapes internal quotes.
func escapeCSV(s string) string {
	if s == "" {
		return ""
	}

	needsQuotes := false
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuotes = true
			break
		}
	}

	if !needsQuotes {
		return s
	}

	// Escape quotes by doubling them and wrap in quotes
	escaped := strings.ReplaceAll(s, `"`, `""`)
	return `"` + escaped + `"`
}
