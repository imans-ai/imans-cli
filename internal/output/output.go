package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Printer struct {
	Out    io.Writer
	ErrOut io.Writer
	JSON   bool
	Quiet  bool
}

type KeyValue struct {
	Key   string
	Value string
}

func (p *Printer) PrintJSON(v any) error {
	encoder := json.NewEncoder(p.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	tw := table.NewWriter()
	tw.SetOutputMirror(p.Out)
	headerRow := make(table.Row, 0, len(headers))
	for _, header := range headers {
		headerRow = append(headerRow, header)
	}
	tw.AppendHeader(headerRow)
	for _, row := range rows {
		outRow := make(table.Row, 0, len(row))
		for _, item := range row {
			outRow = append(outRow, item)
		}
		tw.AppendRow(outRow)
	}
	tw.Render()
	return nil
}

func (p *Printer) PrintKeyValues(items []KeyValue) error {
	max := 0
	for _, item := range items {
		if len(item.Key) > max {
			max = len(item.Key)
		}
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(p.Out, "%-*s  %s\n", max, item.Key, item.Value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Printer) Warnf(format string, args ...any) {
	if p.Quiet {
		return
	}
	_, _ = fmt.Fprintf(p.ErrOut, format+"\n", args...)
}

func JoinParts(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, ", ")
}
