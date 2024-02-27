package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/cli"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks <command> [flags]",
	Short: "Manage tasks",
	Example: `
		$ gotama-cli tasks list --limit=10 --offset=0`,
}

var tasksListCmd = &cobra.Command{
	Use:     "list --limit=<limit> --offset=<offset> [flags]",
	Aliases: []string{"ls"},
	Short:   "List tasks",
	Long: `
	List tasks.

	The --limit and --offset flags are optional.`,
	Example: `
$ gotama-cli tasks list
$ gotama-cli tasks list --limit=10 --offset=0`,
	Run: tasksList,
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(tasksListCmd)
	//TODO: implement the rest of the API
	tasksListCmd.Flags().Int("limit", 100, "page size")
	tasksListCmd.Flags().Int("offset", 0, "offset size")
}

func tasksList(cmd *cobra.Command, args []string) {
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tasks, err := cli.GetTasks(offset, limit)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	printTable(
		[]string{
			"ID",
			"Status",
			"Name",
			"Type",
			"Period",
			"Payload",
			"Error",
			"CreatedAt",
			"CompletedAt",
			"FailedAt",
		},
		func(w io.Writer, tmpl string) {
			for _, t := range tasks {
				payload, _ := json.Marshal(t.Payload)
				fmt.Fprintf(w, tmpl,
					t.ID,
					t.Status,
					t.Name,
					t.Type,
					t.Period,
					string(payload),
					base.NewSafeString(t.Error).String(),
					t.CreatedAt,
					base.NewSafeString(t.CompletedAt).String(),
					base.NewSafeString(t.FailedAt).String(),
				)
			}
		},
	)
}
