package util

import (
	"strings"

	"github.com/olekukonko/tablewriter"
)

func ToASCIITable[T any](data []T, headers []string, columns func(T) []string) string {

	var builder strings.Builder

	table := tablewriter.NewWriter(&builder)
	table.SetHeader(headers)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, item := range data {
		table.Append(columns(item))
	}
	table.Render() // Send output

	return builder.String()
}
