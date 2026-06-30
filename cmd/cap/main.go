package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/a-h/cap/cmd/globals"
	capcontext "github.com/a-h/cap/context"
	"github.com/a-h/cap/model"
	"github.com/a-h/cap/query"
	"github.com/a-h/cap/review"
	"github.com/a-h/cap/store"
	"github.com/a-h/cap/validate"
	"github.com/alecthomas/kong"
)

type CLI struct {
	globals.Globals
	Version VersionCmd `cmd:"" help:"Show version information"`
	Init    InitCmd    `cmd:"" help:"Create the system directory layout and install templates"`
	New     NewCmd     `cmd:"" help:"Scaffold a new entity from its template"`
	List    ListCmd    `cmd:"" help:"List the entities of a kind"`

	Validate ValidateCmd `cmd:"" help:"Check the model for structural and reference problems"`
	Show     ShowCmd     `cmd:"" help:"Show an entity, what it links to, and what links to it"`
	Graph    GraphCmd    `cmd:"" help:"Print the top-down composition tree for an entity"`
	Context  ContextCmd  `cmd:"" help:"Print the context bundle for a capability"`
	Review   ReviewCmd   `cmd:"" help:"Print a review packet for an entity, or the whole model"`
}

var Version = "dev"

type VersionCmd struct{}

func (cmd *VersionCmd) Run(out io.Writer) error {
	fmt.Fprint(out, Version)
	return nil
}

type InitCmd struct {
	Root string `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
}

func (cmd *InitCmd) Run(out io.Writer) error {
	written, err := store.Init(cmd.Root)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Initialised %s\n", cmd.Root)
	for _, name := range written {
		fmt.Fprintf(out, "  installed template %s\n", name)
	}
	return nil
}

// kindAliases maps the accepted names and short forms a user may type to an entity
// kind. Both the singular and plural, and a short alias, resolve to the same kind.
var kindAliases = map[string]model.Kind{
	"context": model.KindContext, "contexts": model.KindContext, "ctx": model.KindContext,
	"concept": model.KindConcept, "concepts": model.KindConcept, "con": model.KindConcept,
	"capability": model.KindCapability, "capabilities": model.KindCapability, "cap": model.KindCapability,
	"invariant": model.KindInvariant, "invariants": model.KindInvariant, "inv": model.KindInvariant,
	"specification": model.KindSpecification, "specifications": model.KindSpecification, "spec": model.KindSpecification,
	"adr": model.KindADR, "adrs": model.KindADR,
	"scenario": model.KindScenario, "scenarios": model.KindScenario, "scn": model.KindScenario,
	"verification": model.KindVerification, "verifications": model.KindVerification, "ver": model.KindVerification,
	"task": model.KindTask, "tasks": model.KindTask,
}

// resolveKind maps a user-supplied kind name or alias to an entity kind.
func resolveKind(name string) (model.Kind, error) {
	if kind, ok := kindAliases[strings.ToLower(name)]; ok {
		return kind, nil
	}
	return "", fmt.Errorf("unknown kind %q", name)
}

type NewCmd struct {
	Kind string `arg:"" help:"Entity kind or alias (context/ctx, concept/con, capability/cap, invariant/inv, specification/spec, adr, scenario/scn, verification/ver, task)"`
	Name string `arg:"" help:"Name of the entity, used as the title and filename slug"`
	Root string `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
}

func (cmd *NewCmd) Run(out io.Writer) error {
	kind, err := resolveKind(cmd.Kind)
	if err != nil {
		return err
	}
	path, err := store.Scaffold(cmd.Root, kind, cmd.Name)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Created %s\n", path)
	return nil
}

type ListCmd struct {
	Kind   string `arg:"" help:"Entity kind or alias to list (for example capabilities, caps, scenarios)"`
	Root   string `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
	Format string `help:"Output format" default:"text" enum:"text,json"`
}

func (cmd *ListCmd) Run(out io.Writer) error {
	kind, err := resolveKind(cmd.Kind)
	if err != nil {
		return err
	}
	res, err := store.Load(cmd.Root)
	if err != nil {
		return err
	}
	entries := query.List(res.Model, kind)
	if cmd.Format == "json" {
		return encodeJSON(out, entries)
	}
	for _, e := range entries {
		if e.Inline {
			fmt.Fprintf(out, "%s  %s  (inline)\n", e.ID, e.Title)
			continue
		}
		fmt.Fprintf(out, "%s  %s\n", e.ID, e.Title)
	}
	return nil
}

type ShowCmd struct {
	ID     string `arg:"" help:"Entity identifier, for example cap-0003"`
	Root   string `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
	Format string `help:"Output format" default:"text" enum:"text,json"`
}

func (cmd *ShowCmd) Run(out io.Writer) error {
	res, err := store.Load(cmd.Root)
	if err != nil {
		return err
	}
	id := model.ID(cmd.ID).Canonical()
	kind, ok := res.Model.Lookup(id)
	if !ok {
		return fmt.Errorf("entity %s not found", id)
	}
	view := buildShow(res.Model, id, kind)
	if cmd.Format == "json" {
		return encodeJSON(out, view)
	}
	fmt.Fprint(out, view.render())
	return nil
}

type GraphCmd struct {
	ID      string   `arg:"" optional:"" help:"Entity identifier to root the tree at, for example scn-0001; graphs the whole model from its bounded contexts when omitted"`
	Root    string   `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
	Format  string   `help:"Output format" default:"text" enum:"text,json,dot"`
	Depth   int      `help:"Limit how many links from each root to expand; 0 means unlimited"`
	Exclude []string `help:"Entity kinds to omit, for example verification,task"`
}

func (cmd *GraphCmd) Run(out io.Writer) error {
	res, err := store.Load(cmd.Root)
	if err != nil {
		return err
	}
	opts, err := cmd.options()
	if err != nil {
		return err
	}
	if cmd.ID == "" {
		if cmd.Format == "dot" {
			fmt.Fprint(out, query.RenderDOT(query.BuildGraph(res.Model, opts)))
			return nil
		}
		roots := query.BuildForest(res.Model, opts)
		if cmd.Format == "json" {
			return encodeJSON(out, roots)
		}
		fmt.Fprint(out, query.RenderForest(roots))
		return nil
	}
	id := model.ID(cmd.ID).Canonical()
	if _, ok := res.Model.Lookup(id); !ok {
		return fmt.Errorf("entity %s not found", id)
	}
	if cmd.Format == "dot" {
		fmt.Fprint(out, query.RenderDOT(query.BuildGraphFrom(res.Model, id, opts)))
		return nil
	}
	tree := query.BuildTree(res.Model, id, opts)
	if cmd.Format == "json" {
		return encodeJSON(out, tree)
	}
	fmt.Fprint(out, tree.Render())
	return nil
}

// options builds the query options from the depth and exclude flags, resolving each
// exclude value to an entity kind.
func (cmd *GraphCmd) options() (query.Options, error) {
	opts := query.Options{MaxDepth: cmd.Depth}
	for _, name := range cmd.Exclude {
		kind, err := resolveKind(name)
		if err != nil {
			return query.Options{}, err
		}
		if opts.Exclude == nil {
			opts.Exclude = map[model.Kind]bool{}
		}
		opts.Exclude[kind] = true
	}
	return opts, nil
}

type showView struct {
	ID       model.ID      `json:"id"`
	Kind     model.Kind    `json:"kind"`
	Title    string        `json:"title"`
	LinksTo  []query.Entry `json:"linksTo"`
	LinkedBy []query.Entry `json:"linkedBy"`
}

func buildShow(m *model.Model, id model.ID, kind model.Kind) showView {
	v := showView{ID: id, Kind: kind, Title: query.Title(m, id)}
	for _, child := range query.Children(m, id) {
		v.LinksTo = append(v.LinksTo, query.Entry{ID: child, Title: query.Title(m, child)})
	}
	for _, parent := range query.Parents(m, id) {
		v.LinkedBy = append(v.LinkedBy, query.Entry{ID: parent, Title: query.Title(m, parent)})
	}
	return v
}

func (v showView) render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s (%s)\n", v.ID, v.Title, v.Kind)
	if len(v.LinksTo) > 0 {
		b.WriteString("\nLinks to:\n")
		for _, e := range v.LinksTo {
			fmt.Fprintf(&b, "  - %s  %s\n", e.ID, e.Title)
		}
	}
	if len(v.LinkedBy) > 0 {
		b.WriteString("\nLinked by:\n")
		for _, e := range v.LinkedBy {
			fmt.Fprintf(&b, "  - %s  %s\n", e.ID, e.Title)
		}
	}
	return b.String()
}

type ValidateCmd struct {
	Root   string `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
	Strict bool   `help:"Treat warnings as errors when setting the exit status"`
	Format string `help:"Output format" default:"text" enum:"text,json"`
}

func (cmd *ValidateCmd) Run(out io.Writer) error {
	res, err := store.Load(cmd.Root)
	if err != nil {
		return err
	}
	problems := append(res.Problems, validate.Check(res.Model, res.Files)...)
	sort.SliceStable(problems, func(i, j int) bool {
		if problems[i].File != problems[j].File {
			return problems[i].File < problems[j].File
		}
		return problems[i].Line < problems[j].Line
	})

	if err := printProblems(out, cmd.Format, problems); err != nil {
		return err
	}

	var errors, warnings int
	for _, p := range problems {
		switch p.Severity {
		case store.SeverityError:
			errors++
		case store.SeverityWarning:
			warnings++
		}
	}
	if errors > 0 || (cmd.Strict && warnings > 0) {
		return fmt.Errorf("validation failed: %d error(s), %d warning(s)", errors, warnings)
	}
	return nil
}

type ContextCmd struct {
	ID     string `arg:"" help:"Capability identifier, for example cap-0003"`
	Root   string `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
	Format string `help:"Output format" default:"text" enum:"text,json"`
}

func (cmd *ContextCmd) Run(out io.Writer) error {
	res, err := store.Load(cmd.Root)
	if err != nil {
		return err
	}
	id := model.ID(cmd.ID).Canonical()
	bundle, ok := capcontext.For(res.Model, id)
	if !ok {
		return fmt.Errorf("capability %s not found", id)
	}
	if cmd.Format == "json" {
		return encodeJSON(out, bundle)
	}
	fmt.Fprint(out, bundle.String())
	return nil
}

type ReviewCmd struct {
	ID     string `arg:"" optional:"" help:"Entity identifier to review; reviews the whole model when omitted"`
	Root   string `help:"Path to the system model root directory" default:"cap" env:"CAP_ROOT"`
	Format string `help:"Output format" default:"text" enum:"text,json"`
}

func (cmd *ReviewCmd) Run(out io.Writer) error {
	res, err := store.Load(cmd.Root)
	if err != nil {
		return err
	}
	if cmd.ID != "" {
		id := model.ID(cmd.ID).Canonical()
		packet, ok := review.Assemble(res, id)
		if !ok {
			return fmt.Errorf("entity %s not found", id)
		}
		return printReview(out, cmd.Format, []review.Packet{packet})
	}
	return printReview(out, cmd.Format, review.AssembleAll(res))
}

func printReview(out io.Writer, format string, packets []review.Packet) error {
	if format == "json" {
		return encodeJSON(out, packets)
	}
	for i, p := range packets {
		if i > 0 {
			fmt.Fprintln(out, "\n---")
		}
		fmt.Fprint(out, p.Render())
	}
	return nil
}

func printProblems(out io.Writer, format string, problems []store.Problem) error {
	if format == "json" {
		return encodeJSON(out, problems)
	}
	for _, p := range problems {
		fmt.Fprintln(out, p.String())
	}
	return nil
}

// encodeJSON writes value to out as indented JSON.
func encodeJSON(out io.Writer, value any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

// run parses args, dispatches to the selected command writing its output to out,
// and returns the command's error. It is the testable entry point that main wraps,
// so a test drives the CLI in-process and asserts on the captured output.
func run(args []string, out io.Writer) error {
	cli := CLI{Globals: globals.Globals{}}
	parser, err := kong.New(&cli,
		kong.Name("cap"),
		kong.Description("Capability-centred system traceability"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Writers(out, out),
		kong.BindTo(out, (*io.Writer)(nil)),
	)
	if err != nil {
		return err
	}
	ctx, err := parser.Parse(args)
	if err != nil {
		var parseErr *kong.ParseError
		if len(args) == 0 && errors.As(err, &parseErr) {
			return parseErr.Context.PrintUsage(false)
		}
		return err
	}
	return ctx.Run(&cli.Globals)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "cap: "+err.Error())
		os.Exit(1)
	}
}
