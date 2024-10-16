// SPDX-License-Identifier: Apache-2.0

package printer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/mattn/go-isatty"
	"gopkg.in/yaml.v3"
)

var (
	errInputNotASliceOfStructs = errors.New("input is not a slice of structs")
	errInputNotASlice          = errors.New("input is not a slice")
	errElementNotASlice        = errors.New("element is not a slice")
)

// structToTable converts a slice of structs into a tabular representation
func structToTable(data interface{}) ([][]string, error) {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Slice {
		return nil, errors.Join(errInputNotASliceOfStructs, errInputNotASlice)
	}

	var headers []string
	elemType := val.Type().Elem()

	if elemType.Kind() != reflect.Struct {
		return nil, errors.Join(errInputNotASliceOfStructs, errElementNotASlice)
	}

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = field.Name
		} else {
			// Remove `omitEmpty` etc. options
			if idx := strings.Index(fieldName, ","); idx != -1 {
				fieldName = fieldName[:idx]
			}
		}
		headers = append(headers, fieldName)
	}

	result := [][]string{headers}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		values := make([]string, elem.NumField())
		for j := 0; j < elem.NumField(); j++ {
			fieldValue := elem.Field(j).Interface()
			fieldValueStr, err := yaml.Marshal(fieldValue)
			if err != nil {
				return nil, err
			}

			// Remove trailing newline from YAML output since we add it ourselves when printing
			values[j] = strings.TrimSuffix(string(fieldValueStr), "\n")
		}
		result = append(result, values)
	}

	return result, nil
}

var IsTTY = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

// Format defines the option output format of a resource.
type Format int

const (
	// Human prints it in human readable format. This can be either a table or
	// a single line, depending on the resource implementation.
	Human Format = iota
	JSON
)

// NewFormatValue is used to define a flag that can be used to define a custom
// flag via the flagset.Var() method.
func NewFormatValue(val Format, p *Format) *Format {
	*p = val
	return p
}

func (f *Format) String() string {
	switch *f {
	case Human:
		return "human"
	case JSON:
		return "json"
	}

	return "unknown format"
}

func (f *Format) Set(s string) error {
	var v Format
	switch s {
	case "human":
		v = Human
	case "json":
		v = JSON
	default:
		return fmt.Errorf("failed to parse Format: %q. Valid values: %+v",
			s, []string{"human", "json"})
	}

	*f = v
	return nil
}

func (f *Format) Type() string {
	return "string"
}

// Printer is used to print information to the defined output.
type Printer struct {
	humanOut    io.Writer
	resourceOut io.Writer

	format *Format
}

// NewPrinter returns a new Printer for the given output and format.
func NewPrinter(format *Format) *Printer {
	return &Printer{
		format: format,
	}
}

// Printf is a convenience method to Printf to the defined output.
func (p *Printer) Printf(format string, i ...interface{}) {
	_, _ = fmt.Fprintf(p.Out(), format, i...)
}

// Println is a convenience method to Println to the defined output.
func (p *Printer) Println(i ...interface{}) {
	_, _ = fmt.Fprintln(p.Out(), i...)
}

// Print is a convenience method to Print to the defined output.
func (p *Printer) Print(i ...interface{}) {
	_, _ = fmt.Fprint(p.Out(), i...)
}

// Out defines the output to write human-readable text. If format is not set to
// human, Out returns io.Discard, which means that any output will be
// discarded
func (p *Printer) Out() io.Writer {
	if p.humanOut != nil {
		return p.humanOut
	}

	if *p.format == Human {
		return color.Output
	}

	return io.Discard
}

// PrintProgress starts a spinner with the relevant message. The returned
// function needs to be called in a defer or when it's decided to stop the
// spinner
func (p *Printer) PrintProgress(message string) func() {
	if !IsTTY {
		_, _ = fmt.Fprintln(p.Out(), message)
		return func() {}
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(p.Out()))
	s.Suffix = fmt.Sprintf(" %s", message)

	_ = s.Color("bold", "magenta")
	s.Start()
	return func() {
		s.Stop()

		// NOTE(fatih) the spinner library doesn't clear the line properly,
		// hence remove it ourselves. This line should be removed once it's
		// fixed in upstream.  https://github.com/briandowns/spinner/pull/117
		_, _ = fmt.Fprint(p.Out(), "\r\033[2K")
	}
}

// Format returns the format that was set for this printer
func (p *Printer) Format() Format { return *p.format }

// SetHumanOutput sets the output for human readable messages.
func (p *Printer) SetHumanOutput(out io.Writer) {
	p.humanOut = out
}

// SetResourceOutput sets the output for printing resources via PrintResource.
func (p *Printer) SetResourceOutput(out io.Writer) {
	p.resourceOut = out
}

// PrintResource prints the given resource in the format it was specified.
func (p *Printer) PrintResource(v interface{}) error {
	if p.format == nil {
		return errors.New("printer.Format is not set")
	}

	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	switch *p.format {
	case Human:
		var b string
		result, err := structToTable(v)
		if err == nil {
			t := table.NewWriter()
			if !color.NoColor {
				t.SetStyle(table.StyleColoredBlackOnMagentaWhite)
			}

			for i, line := range result {
				row := make(table.Row, len(line))
				for i, l := range line {
					row[i] = l
				}

				if i == 0 {
					t.AppendHeader(row)
				} else {
					t.AppendRow(row)
				}
			}

			b = t.Render()
		} else if errors.Is(err, errInputNotASliceOfStructs) {
			s, err := yaml.Marshal(v)
			if err != nil {
				return err
			}

			// Remove trailing newline from YAML output since we add it ourselves when printing
			b = strings.TrimSuffix(string(s), "\n")
		} else {
			return err
		}

		_, _ = fmt.Fprintln(out, b)
		return nil
	case JSON:
		return p.PrintJSON(v)
	}

	return fmt.Errorf("unknown printer.Format: %T", *p.format)
}

func (p *Printer) ConfirmCommand(confirmationName, commandShortName, confirmFailedName string) error {
	if p.Format() != Human {
		return fmt.Errorf("cannot %s with the output format %q (run with -force to override)", commandShortName, p.Format())
	}

	if !IsTTY {
		return fmt.Errorf("cannot confirm %s %q (run with -force to override)", confirmFailedName, confirmationName)
	}

	confirmationMessage := fmt.Sprintf("%s %s %s", Bold("Please type"), BoldBlue(confirmationName), Bold("to confirm:"))

	prompt := &survey.Input{
		Message: confirmationMessage,
	}

	var userInput string
	err := survey.AskOne(prompt, &userInput)
	if err != nil {
		if err == terminal.InterruptErr {
			os.Exit(0)
		} else {
			return err
		}
	}

	// If the confirmations don't match up, let's return an error.
	if userInput != confirmationName {
		return fmt.Errorf("incorrect value entered, skipping %s", commandShortName)
	}

	return nil
}

func (p *Printer) PrintJSON(v interface{}) error {
	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(out, string(buf))
	return nil

}

func (p *Printer) PrettyPrintJSON(b []byte) error {
	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	var buf bytes.Buffer
	err := json.Indent(&buf, b, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(out, buf.String())
	return nil
}

func GetMilliseconds(timestamp time.Time) int64 {
	if timestamp.IsZero() {
		return 0
	}

	numSeconds := timestamp.UTC().UnixNano() /
		(int64(time.Millisecond) / int64(time.Nanosecond))

	return numSeconds
}

func GetMillisecondsIfExists(timestamp *time.Time) *int64 {
	if timestamp == nil {
		return nil
	}

	numSeconds := GetMilliseconds(*timestamp)

	return &numSeconds
}

func Emoji(emoji string) string {
	if IsTTY {
		return emoji
	}
	return ""
}

// BoldBlue returns a string formatted with blue and bold.
func BoldBlue(msg interface{}) string {
	// the 'color' package already handles IsTTY gracefully
	return color.New(color.FgBlue).Add(color.Bold).Sprint(msg)
}

// BoldRed returns a string formatted with red and bold.
func BoldRed(msg interface{}) string {
	return color.New(color.FgRed).Add(color.Bold).Sprint(msg)
}

// BoldGreen returns a string formatted with green and bold.
func BoldGreen(msg interface{}) string {
	return color.New(color.FgGreen).Add(color.Bold).Sprint(msg)
}

// BoldBlack returns a string formatted with Black and bold.
func BoldBlack(msg interface{}) string {
	return color.New(color.FgBlack).Add(color.Bold).Sprint(msg)
}

// BoldYellow returns a string formatted with yellow and bold.
func BoldYellow(msg interface{}) string {
	return color.New(color.FgYellow).Add(color.Bold).Sprint(msg)
}

// Red returns a string formatted with red and bold.
func Red(msg interface{}) string {
	return color.New(color.FgRed).Sprint(msg)
}

// Bold returns a string formatted with bold.
func Bold(msg interface{}) string {
	// the 'color' package already handles IsTTY gracefully
	return color.New(color.Bold).Sprint(msg)
}
