package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

const Version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:           "gotama-cli <command> <subcommand> [flags]",
	Short:         "gotama-cli",
	Long:          `Command line tool to inspect tasks and queues managed by gotama`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Example:       `$ gotama-cli tasks list --limit=10 --offset=0`,
}

func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printTable(cols []string, printRows func(w io.Writer, tmpl string)) {
	format := strings.Repeat("%v\t", len(cols)) + "\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	var headers []interface{}
	var seps []interface{}
	for _, name := range cols {
		headers = append(headers, name)
		seps = append(seps, strings.Repeat("-", len(name)))
	}
	fmt.Fprintf(tw, format, headers...)
	fmt.Fprintf(tw, format, seps...)
	printRows(tw, format)
	tw.Flush()
}
