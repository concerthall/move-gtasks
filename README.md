# move-gtasks

<img src="logo.svg" width="20%" />

Move due-dated, incomplete Google Tasks to another day.

Adding a due date to a Google Tasks allows it to show up in your
Google Calendar, but moving Google Tasks to different days
is tedious. You are unable to move groups of tasks by multi-selecting
them and dragging them to a new day.

You will need to enable a Google Cloud Platform project
with the Google Tasks API in order for this to work.

See https://developers.google.com/tasks/quickstart/go for
more information on how to set up a project and download
OAuth credentials.

```
Usage:
  move-gtasks [flags]

Flags:
  -c, --clear-token   Clears your existing token from the filesystem before running the tool. Do this if you want to re-run OAuth workflows.
  -f, --from string   Date from which tasks should be pulled. Must be formatted as YYYY-MM-DD, or be one of [yesterday, today, tomorrow] (default "today")
  -h, --help          help for move-gtasks
  -t, --to string     Date that should receive tasks. Must be formatted as YYYY-MM-DD, or be the one of [yesterday, today, tomorrow]. (default "tomorrow")
  -v, --version       version for move-gtasks
```

---

**logo.svg image credit: [undraw.co](https://undraw.co/)**