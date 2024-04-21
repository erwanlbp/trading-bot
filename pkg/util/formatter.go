package util

import (
	"strings"

	"github.com/olekukonko/tablewriter"
)

type ASCIITableOption func(*tablewriter.Table)

func ToASCIITable[T any](data []T, headers []string, footer []string, columns func(T) []string, opts ...ASCIITableOption) string {

	var builder strings.Builder

	table := tablewriter.NewWriter(&builder)
	table.SetHeader(headers)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("|")
	for _, item := range data {
		table.Append(columns(item))
	}
	table.SetFooter(footer)

	for _, opt := range opts {
		opt(table)
	}

	table.Render() // Send output

	return builder.String()
}
