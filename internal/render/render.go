package render

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/sunil-saini/astat/internal/output"
)

type TableData struct {
	Headers []string
	Rows    [][]string
	JSON    any
}

func Print(d TableData) error {
	switch viper.GetString("output") {
	case "json":
		return output.PrintJSON(d.JSON)

	case "table", "":
		t := output.NewTable(d.Headers)
		for _, row := range d.Rows {
			t.Append(row)
		}
		t.Render()
		return nil

	default:
		return fmt.Errorf("unknown output format")
	}
}
