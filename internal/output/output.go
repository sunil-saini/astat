package output

import "github.com/spf13/viper"

type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
)

func FormatFromConfig() Format {
	switch viper.GetString("output") {
	case "json":
		return JSON
	default:
		return Table
	}
}
