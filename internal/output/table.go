package output

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

func NewTable(headers []string) *tablewriter.Table {
	t := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader(headers),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{
				Left:   tw.On,
				Right:  tw.On,
				Top:    tw.On,
				Bottom: tw.On,
			},
		}),
	)
	return t
}
