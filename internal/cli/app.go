package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/sorafujitani/agent-deck/internal/deck"
)

type DeckService interface {
	Inbox(includeDone bool) ([]deck.Task, error)
	CreateTask(input deck.NewTaskInput) (deck.Task, error)
	GetTask(id string) (deck.Task, error)
	UpdateTask(id string, input deck.UpdateTaskInput) (deck.Task, error)
	MarkDone(id string, nextAction *string) (deck.Task, error)
	AddRun(id string, input deck.AddRunInput) (deck.Task, deck.RunRecord, error)
	AddArtifact(id string, input deck.AddArtifactInput) (deck.Task, deck.Artifact, error)
}

type App struct {
	service   DeckService
	storePath string
}

func NewApp(service DeckService, storePath string) *App {
	return &App{
		service:   service,
		storePath: storePath,
	}
}

func (a *App) Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		PrintHelp(stdout)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		PrintHelp(stdout)
		return 0
	case "path":
		fmt.Fprintln(stdout, a.storePath)
		return 0
	case "inbox":
		return a.runInbox(args[1:], stdout, stderr)
	case "new":
		return a.runNew(args[1:], stdout, stderr)
	case "show":
		return a.runShow(args[1:], stdout, stderr)
	case "update":
		return a.runUpdate(args[1:], stdout, stderr)
	case "done":
		return a.runDone(args[1:], stdout, stderr)
	case "run":
		return a.runAddRun(args[1:], stdout, stderr)
	case "artifact":
		return a.runArtifact(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		PrintHelp(stderr)
		return 1
	}
}

func (a *App) runInbox(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("inbox", stderr)
	includeDone := fs.Bool("all", false, "include done tasks")
	asJSON := fs.Bool("json", false, "print JSON")
	if !parseFlags(fs, args) {
		return 2
	}

	tasks, err := a.service.Inbox(*includeDone)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *asJSON {
		return printJSON(stdout, stderr, tasks)
	}
	PrintInbox(stdout, tasks)
	return 0
}

func (a *App) runNew(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("new", stderr)
	repo := fs.String("repo", "", "repository path")
	issue := fs.String("issue", "", "issue URL")
	pr := fs.String("pr", "", "pull request URL")
	nextAction := fs.String("next", "", "next action")
	asJSON := fs.Bool("json", false, "print JSON")
	var context repeatedFlag
	fs.Var(&context, "context", "context line")
	if !parseFlags(fs, args) {
		return 2
	}

	task, err := a.service.CreateTask(deck.NewTaskInput{
		Goal:       strings.Join(fs.Args(), " "),
		Repo:       *repo,
		Issue:      *issue,
		PR:         *pr,
		Context:    context,
		NextAction: *nextAction,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *asJSON {
		return printJSON(stdout, stderr, task)
	}
	fmt.Fprintf(stdout, "created %s\n", task.ID)
	return 0
}

func (a *App) runShow(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("show", stderr)
	asJSON := fs.Bool("json", false, "print JSON")
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := optionalID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	task, err := a.service.GetTask(id)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *asJSON {
		return printJSON(stdout, stderr, task)
	}
	PrintTask(stdout, task)
	return 0
}

func (a *App) runUpdate(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("update", stderr)
	status := fs.String("status", "", "status")
	nextAction := fs.String("next", "", "next action")
	asJSON := fs.Bool("json", false, "print JSON")
	var context repeatedFlag
	fs.Var(&context, "context", "context line")
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := optionalID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	if *status != "" {
		if _, err := deck.ParseStatus(*status); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}

	var next *string
	if *nextAction != "" {
		next = nextAction
	}
	task, err := a.service.UpdateTask(id, deck.UpdateTaskInput{
		Status:     *status,
		NextAction: next,
		Context:    context,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *asJSON {
		return printJSON(stdout, stderr, task)
	}
	fmt.Fprintf(stdout, "updated %s  %s\n", task.ID, task.Status)
	return 0
}

func (a *App) runDone(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("done", stderr)
	nextAction := fs.String("next", "", "final note")
	asJSON := fs.Bool("json", false, "print JSON")
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := optionalID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	var next *string
	if *nextAction != "" {
		next = nextAction
	}
	task, err := a.service.MarkDone(id, next)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *asJSON {
		return printJSON(stdout, stderr, task)
	}
	fmt.Fprintf(stdout, "done %s\n", task.ID)
	return 0
}

func (a *App) runAddRun(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("run", stderr)
	agent := fs.String("agent", "agent", "agent name")
	summary := fs.String("summary", "", "run summary")
	asJSON := fs.Bool("json", false, "print JSON")
	if !parseFlags(fs, args) {
		return 2
	}
	id, ok := optionalID(fs.Args(), stderr)
	if !ok {
		return 2
	}

	task, run, err := a.service.AddRun(id, deck.AddRunInput{
		Agent:   *agent,
		Summary: *summary,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *asJSON {
		return printJSON(stdout, stderr, addRunResult{Task: task, Run: run})
	}
	fmt.Fprintf(stdout, "added %s to %s  %s\n", run.ID, task.ID, task.Status)
	return 0
}

func (a *App) runArtifact(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("artifact", stderr)
	kind := fs.String("kind", "file", "artifact kind")
	note := fs.String("note", "", "artifact note")
	asJSON := fs.Bool("json", false, "print JSON")
	if !parseFlags(fs, args) {
		return 2
	}
	rest := fs.Args()
	id, path, ok := artifactArgs(rest, stderr)
	if !ok {
		return 2
	}

	task, artifact, err := a.service.AddArtifact(id, deck.AddArtifactInput{
		Kind: *kind,
		Path: path,
		Note: *note,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *asJSON {
		return printJSON(stdout, stderr, addArtifactResult{Task: task, Artifact: artifact})
	}
	fmt.Fprintf(stdout, "added %s artifact to %s\n", artifact.Kind, task.ID)
	return 0
}

type addRunResult struct {
	Task deck.Task      `json:"task"`
	Run  deck.RunRecord `json:"run"`
}

type addArtifactResult struct {
	Task     deck.Task     `json:"task"`
	Artifact deck.Artifact `json:"artifact"`
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

func optionalID(args []string, stderr io.Writer) (string, bool) {
	if len(args) > 1 {
		fmt.Fprintln(stderr, "expected zero or one task id")
		return "", false
	}
	if len(args) == 0 {
		return "latest", true
	}
	return args[0], true
}

func artifactArgs(args []string, stderr io.Writer) (string, string, bool) {
	switch len(args) {
	case 1:
		return "latest", args[0], true
	case 2:
		return args[0], args[1], true
	default:
		fmt.Fprintln(stderr, "usage: deck artifact [<task-id>|latest] <path>")
		return "", "", false
	}
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
