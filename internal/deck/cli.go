package deck

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"
)

func Main(args []string, stdout, stderr io.Writer) int {
	store, err := DefaultStore()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return Run(args, stdout, stderr, store, time.Now)
}

func Run(args []string, stdout, stderr io.Writer, store Store, now func() time.Time) int {
	if len(args) == 0 {
		PrintHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		PrintHelp(stdout)
		return 0
	case "path":
		fmt.Fprintln(stdout, store.Path())
		return 0
	case "inbox":
		return runInbox(args[1:], stdout, stderr, store)
	case "new":
		return runNew(args[1:], stdout, stderr, store, now)
	case "show":
		return runShow(args[1:], stdout, stderr, store)
	case "update":
		return runUpdate(args[1:], stdout, stderr, store, now)
	case "done":
		return runDone(args[1:], stdout, stderr, store, now)
	case "run":
		return runAddRun(args[1:], stdout, stderr, store, now)
	case "artifact":
		return runArtifact(args[1:], stdout, stderr, store, now)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		PrintHelp(stderr)
		return 1
	}
}

func runInbox(args []string, stdout, stderr io.Writer, store Store) int {
	fs := newFlagSet("inbox", stderr)
	includeDone := fs.Bool("all", false, "include done tasks")
	if !parseFlags(fs, args) {
		return 2
	}

	deck, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	PrintInbox(stdout, deck.Tasks, *includeDone)
	return 0
}

func runNew(args []string, stdout, stderr io.Writer, store Store, now func() time.Time) int {
	fs := newFlagSet("new", stderr)
	repo := fs.String("repo", "", "repository path")
	issue := fs.String("issue", "", "issue URL")
	pr := fs.String("pr", "", "pull request URL")
	nextAction := fs.String("next", "", "next action")
	var context repeatedFlag
	fs.Var(&context, "context", "context line")
	if !parseFlags(fs, args) {
		return 2
	}
	goal := strings.Join(fs.Args(), " ")

	deck, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	task, err := deck.AddTask(NewTaskInput{
		Goal:       goal,
		Repo:       *repo,
		Issue:      *issue,
		PR:         *pr,
		Context:    context,
		NextAction: *nextAction,
	}, now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := store.Save(deck); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "created %s\n", task.ID)
	return 0
}

func runShow(args []string, stdout, stderr io.Writer, store Store) int {
	fs := newFlagSet("show", stderr)
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := singleID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	deck, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	task, ok := deck.FindTask(id)
	if !ok {
		fmt.Fprintf(stderr, "task not found: %s\n", id)
		return 1
	}
	PrintTask(stdout, task)
	return 0
}

func runUpdate(args []string, stdout, stderr io.Writer, store Store, now func() time.Time) int {
	fs := newFlagSet("update", stderr)
	status := fs.String("status", "", "status")
	nextAction := fs.String("next", "", "next action")
	var context repeatedFlag
	fs.Var(&context, "context", "context line")
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := singleID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	if *status != "" {
		if _, err := ParseStatus(*status); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}

	deck, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var next *string
	if *nextAction != "" {
		next = nextAction
	}
	task, err := deck.UpdateTask(id, UpdateTaskInput{
		Status:     *status,
		NextAction: next,
		Context:    context,
	}, now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := store.Save(deck); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "updated %s  %s\n", task.ID, task.Status)
	return 0
}

func runDone(args []string, stdout, stderr io.Writer, store Store, now func() time.Time) int {
	fs := newFlagSet("done", stderr)
	nextAction := fs.String("next", "", "final note")
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := singleID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	deck, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var next *string
	if *nextAction != "" {
		next = nextAction
	}
	task, err := deck.UpdateTask(id, UpdateTaskInput{
		Status:     string(StatusDone),
		NextAction: next,
	}, now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := store.Save(deck); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "done %s\n", task.ID)
	return 0
}

func runAddRun(args []string, stdout, stderr io.Writer, store Store, now func() time.Time) int {
	fs := newFlagSet("run", stderr)
	agent := fs.String("agent", "agent", "agent name")
	summary := fs.String("summary", "", "run summary")
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := singleID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	deck, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	task, run, err := deck.AddRun(id, AddRunInput{
		Agent:   *agent,
		Summary: *summary,
	}, now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := store.Save(deck); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "added %s to %s  %s\n", run.ID, task.ID, task.Status)
	return 0
}

func runArtifact(args []string, stdout, stderr io.Writer, store Store, now func() time.Time) int {
	fs := newFlagSet("artifact", stderr)
	kind := fs.String("kind", "file", "artifact kind")
	note := fs.String("note", "", "artifact note")
	if !parseFlags(fs, args) {
		return 2
	}
	rest := fs.Args()
	if len(rest) != 2 {
		fmt.Fprintln(stderr, "usage: deck artifact <task-id> <path>")
		return 2
	}

	deck, err := store.Load()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	task, artifact, err := deck.AddArtifact(rest[0], AddArtifactInput{
		Kind: *kind,
		Path: rest[1],
		Note: *note,
	}, now())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := store.Save(deck); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "added %s artifact to %s\n", artifact.Kind, task.ID)
	return 0
}

type repeatedFlag []string

func (r *repeatedFlag) String() string {
	return strings.Join(*r, ", ")
}

func (r *repeatedFlag) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) bool {
	return fs.Parse(reorderFlagArgs(fs, args)) == nil
}

func singleID(args []string, stderr io.Writer) (string, bool) {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "expected exactly one task id")
		return "", false
	}
	return args[0], true
}

func reorderFlagArgs(fs *flag.FlagSet, args []string) []string {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}

		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if index := strings.Index(name, "="); index >= 0 {
			name = name[:index]
		}
		flag := fs.Lookup(name)
		if flag == nil || isBoolFlag(flag) || strings.Contains(arg, "=") {
			continue
		}
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positionals...)
}

func isBoolFlag(flag *flag.Flag) bool {
	value, ok := flag.Value.(interface{ IsBoolFlag() bool })
	return ok && value.IsBoolFlag()
}
