package commands

import (
	"os"
	"text/tabwriter"
)

func GetWriter() *tabwriter.Writer {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 4, 4, ' ', 0)
	return w
}
