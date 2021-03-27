package cmd

import (
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// TableWriter initializes and configures a tablewriter.Table object. This table
// should only be used by the FormatTable function.
func TableWriter(writer io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)

	// Configure headers
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	// Configure separators
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)

	// Configure padding and whitespace
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	return table
}

// FormatTable returns a properly formatted string given some input. The first row
// of the data matrix must be the column headers.
func FormatTable(data [][]string) string {
	tableString := &strings.Builder{}
	table := TableWriter(tableString)

	table.SetHeader(data[0])
	table.AppendBulk(data[1:len(data)])
	table.Render()

	return tableString.String()
}
