package commands

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/pkg/mflag"
)

// Config represents the informations on the usages
type Config struct {
	UsageLine   string
	Description string
	Help        string
	Examples    string
}

// Streams allows to overload the output and input
type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

// Env containts the global options
type Env struct {
	Streams
	Debug bool
}

// Command is the interface that is used to handle the commands
type Command interface {
	GetBase() *Base
	GetName() string
	Parse(args []string) ([]string, error)
	Run(args []string) error
}

// Base must be embedded in the commands
type Base struct {
	Env
	Config
	Flags  mflag.FlagSet
	flHelp *bool
}

// Init initialises the Base structure
func (b *Base) Init(c Config) {
	b.Config = c
	b.Streams.Stdout = os.Stdout
	b.Streams.Stdin = os.Stdin
	b.Streams.Stderr = os.Stderr
	b.Flags.SetOutput(ioutil.Discard)
	b.flHelp = b.Flags.Bool([]string{"h", "-help"}, false, "Print usage")
}

// Parse parses the argurments
func (b *Base) Parse(args []string) (newArgs []string, err error) {
	if err = b.Flags.Parse(args); err != nil {
		err = fmt.Errorf("usage: c14 %v", b.UsageLine)
		return
	}
	if *b.flHelp {
		b.PrintUsage()
		os.Exit(1)
		return
	}
	newArgs = b.Flags.Args()
	return
}

// GetBase returns a pointer on Base
func (b *Base) GetBase() *Base {
	return b
}

// PrintUsage print on Stdout the usage message
func (b *Base) PrintUsage() {
	var usageTemplate = `Usage: c14 {{.UsageLine}}

{{.Help}}

{{.Options}}
{{.ExamplesHelp}}
`

	t := template.New("full")
	template.Must(t.Parse(usageTemplate))
	_ = t.Execute(os.Stdout, b)
}

// Options returns the options available, it used by PrintUsage
func (b *Base) Options() string {
	var options string

	visitor := func(flag *mflag.Flag) {
		var optionUsage string

		name := strings.Join(flag.Names, ", -")
		if flag.DefValue == "" {
			optionUsage = fmt.Sprintf("%s=\"\"", name)
		} else {
			optionUsage = fmt.Sprintf("%s=%s", name, flag.DefValue)
		}
		options += fmt.Sprintf("  -%-20s %s\n", optionUsage, flag.Usage)
	}
	b.Flags.VisitAll(visitor)
	if len(options) == 0 {
		return ""
	}
	return fmt.Sprintf("Options:\n%s", options)
}

// ExamplesHelp returns the examples, it used by PrintUsage
func (b *Base) ExamplesHelp() string {
	if b.Examples == "" {
		return ""
	}
	return fmt.Sprintf("Examples:\n%s", strings.Trim(b.Examples, "\n"))
}