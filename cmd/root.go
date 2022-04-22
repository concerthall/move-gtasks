/*
Copyright © 2022 The Concert Hall Developers

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

var (
	// set in init()
	credFile  string
	tokenFile string
)

var rootCmd = &cobra.Command{
	Use:     "move-gtasks",
	Short:   "Move incomplete Google tasks --from a given day --to a given day.",
	Long:    longHelp(),
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		toDateString, _ := cmd.Flags().GetString("to")
		fromDateString, _ := cmd.Flags().GetString("from")
		timeTargets, err := parseTimeTargets(toDateString, fromDateString)
		if err != nil {
			log.Fatal(err)
		}

		clearRequest, _ := cmd.Flags().GetBool("clear-token")
		if clearRequest {
			if err := os.Remove(tokenFile); err != nil {
				log.Fatal(err)
			}
		}

		ctx := context.Background()
		b, err := ioutil.ReadFile(credFile)
		if err != nil {
			log.Fatalf("Unable to read client secret file at path %s: %v", credFile, err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, tasks.TasksScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)

		srv, err := tasks.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Unable to retrieve tasks Client %v", err)
		}

		r, err := srv.Tasklists.List().MaxResults(10).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve task lists. %v", err)
		}

		if len(r.Items) == 0 {
			fmt.Print("No task lists found.")
			return
		}

		// look for the task list called "My Tasks". I believe this
		// is a common default.
		var targetTaskList *tasks.TaskList
		for i, tasklist := range r.Items {
			if tasklist.Title == "My Tasks" {
				targetTaskList = r.Items[i]
				break
			}

			log.Fatal("There was no task list called \"My Tasks\"")
		}

		// Get all tasks in this list.
		tsks, err := srv.Tasks.List(targetTaskList.Id).Do()
		if err != nil {
			log.Fatal("Unable to get tasks from list", targetTaskList.Title)
		}

		// Determine which tasks need updating.
		tasksToUpdate := make([]*tasks.Task, 0)
		for _, task := range tsks.Items {
			if task.Status == "completed" || task.Due == "" {
				// ignore completed tasks or tasks without due dates
				continue
			}

			duedate, _ := time.Parse(time.RFC3339, task.Due)
			if err != nil {
				log.Fatal("couldn't parse the due date", err, "task:", task.Title)
			}

			// only move things from the day the user provided via --from.
			if duedate.YearDay() == timeTargets.From.YearDay() {
				tasksToUpdate = append(tasksToUpdate, task) // BOOKMARK pick up from here
			}
		}

		for _, t := range tasksToUpdate {
			t.Due = timeTargets.To.Format(time.RFC3339)
			_, err = srv.Tasks.Update(targetTaskList.Id, t.Id, t).Do()
			if err != nil {
				fmt.Printf("Failed to move task: %s (%s) to %s\n with error: %s", t.Title, t.Id, timeTargets.To.Format("2006-01-02"), err)
			} else {
				fmt.Printf("Moved Task: %s (%s) to %s\n", t.Title, t.Id, timeTargets.To.Format("2006-01-02"))
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type timeTargets struct {
	To, From time.Time
}

// parseTimeTargets takes common-language references to dates and
// converts them to the right time construct. If a YYYY-MM-DD string
// is provided instead, this will parse that.
func parseTimeTargets(userTo, userFrom string) (*timeTargets, error) {
	var to, from time.Time

	switch userTo {
	case "tomorrow":
		to = time.Now().Add(24 * time.Hour)
	case "yesterday":
		to = time.Now().Add(-24 * time.Hour)
	case "today":
		to = time.Now()
	default:
		var err error
		to, err = time.Parse("2006-01-02", userTo)
		if err != nil {
			return nil, fmt.Errorf("unable to parse the value of \"to\"")
		}
	}

	switch userFrom {
	case "tomorrow":
		from = time.Now().Add(24 * time.Hour)
	case "yesterday":
		from = time.Now().Add(-24 * time.Hour)
	case "today":
		from = time.Now()
	default:
		var err error
		from, err = time.Parse("2006-01-02", userFrom)
		if err != nil {
			return nil, fmt.Errorf("unable to parse the value of \"from\"")
		}
	}

	return &timeTargets{
		To:   to,
		From: from,
	}, nil
}

func init() {
	rootCmd.Flags().StringP("to", "t", "tomorrow", "Date that should receive tasks. Must be formatted as YYYY-MM-DD, or be the one of [yesterday, today, tomorrow].")
	rootCmd.Flags().StringP("from", "f", "today", "Date from which tasks should be pulled. Must be formatted as YYYY-MM-DD, or be one of [yesterday, today, tomorrow]")
	rootCmd.Flags().BoolP("clear-token", "c", false, "Clears your existing token from the filesystem before running the tool. Do this if you want to re-run OAuth workflows.")

	determineAppConfigDir()
}

func determineAppConfigDir() {
	// set up the location where the credentials should be stored.
	configDir, _ := os.UserConfigDir() // TODO: handle this error
	appConfigDir := path.Join(configDir, "move-gtasks")
	err := os.MkdirAll(appConfigDir, 0700)
	if err != nil {
		log.Fatal(err)
	}
	credFile = path.Join(appConfigDir, "credentials.json")
	tokenFile = path.Join(appConfigDir, "token.json")
}

func longHelp() string {
	determineAppConfigDir() // do this as a precaution
	longHelpText := `Move due-dated, incomplete Google Tasks to another day.
Adding a due date to a Google Tasks allows it to show up in your
Google Calendar, but moving Google Tasks to different days
is tedious. You are unable to move groups of tasks by multi-
selecting them and dragging them to a new day.

This tool will move all incomplete Google Tasks in a
given day to a new day to help manage your daily work.

You will need to enable a Google Cloud Platform project
with the Google Tasks API in order for this to work

See https://developers.google.com/tasks/quickstart/go for
more information on how to set up a project and download
OAuth credentials.

Things to note:

- It is not possible to move a task to a specific time.
- It is not possible to differentiate between recurring tasks
  and normal, dated tasks.

Example, if you wanted to move tasks from today to next year
(wow!), you might run it like this.

  move-gtasks --from 2022-04-05 --to 2023-04-05

If you don't provide --to and --from, we assume you want to move
today's tasks to tomorrow.
`

	credFileHelpText := fmt.Sprintf(`Store your Client OAuth creds at %s`, credFile)

	return fmt.Sprintf("%s\n\n%s", longHelpText, credFileHelpText)

}
