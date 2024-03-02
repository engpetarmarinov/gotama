package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/cli"
	"github.com/engpetarmarinov/gotama/internal/task"
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
	Use:     "list [id] [flags]",
	Aliases: []string{"ls"},
	Short:   "List tasks",
	Long: `
	List tasks.

	The --limit and --offset flags are optional.`,
	Example: `
$ gotama-cli tasks list
$ gotama-cli tasks list --limit=10 --offset=0
$ gotama-cli tasks list aac6ed79-4fc6-4b14-8614-889a8236ba54`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
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
			listTasks(limit, offset)
		} else {
			id := args[0]
			listTask(id)
		}
	},
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(tasksListCmd)
	tasksListCmd.Flags().Int("limit", 100, "page size")
	tasksListCmd.Flags().Int("offset", 0, "offset size")
	//TODO: implement the rest of the API
}

func listTasks(limit int, offset int) {
	tasks, err := cli.GetTasks(offset, limit)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	printTasksTable(tasks)
}

func listTask(id string) {
	tasks, err := cli.GetTask(id)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	printTasksTable(tasks)
}

func printTasksTable(tasks []task.Response) {
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
