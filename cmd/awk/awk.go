package awk

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/benhoyt/goawk/interp"
	"github.com/benhoyt/goawk/parser"
)

// Command a container for awk script execution
type Command struct {
	Parser      *parser.Program
	Config      *interp.Config
	Interpreter *interp.Interpreter
}

// NewCommand make a new Awk struct for running awk scripts
func NewCommand(command string) (awk *Command, err error) {
	awk = &Command{}
	prog, err := parser.ParseProgram([]byte(command), nil)
	if err != nil {
		err = fmt.Errorf("got error %v", err)
		return
	}
	interpreter, err := interp.New(prog)
	if err != nil {
		return
	}
	awk.Interpreter = interpreter

	return
}

// Execute run a precompiled interpreter against a payload
func (cmd *Command) Execute(payload string) (output string, err error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	config := &interp.Config{
		Output: outBuf,
		Stdin:  strings.NewReader(payload),
		Error:  errBuf,
	}

	result, err := cmd.Interpreter.Execute(config)
	if err != nil {
		err = fmt.Errorf("got error %d - %v", result, err)
		return
	}

	return outBuf.String(), err
}
