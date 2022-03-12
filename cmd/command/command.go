package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/imarsman/goparallel/cmd/tasks"
)

// Command a command
type Command struct {
	Value string
}

func NewCommand(value string) *Command {
	c := Command{Value: value}
	return &c
}

func isValid(fp string) bool {
	// Check if file already exists
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	return false
}

// Prepare replace placeholders with data from incoming
func (c *Command) Copy() (newCommand Command) {
	newCommand = *c

	return newCommand
}

// Prepare replace placeholders with data from incoming
func (c *Command) Prepare(task *tasks.Task) {
	if !isValid(task.Task) {
		return
	}
	// /path/to/file.ext interpolated
	if strings.Contains(c.Value, "{}") {
		fmt.Println("found")
		c.Value = strings.ReplaceAll(c.Value, "{}", task.Task)
	}
	// /path/to/file.ext -> /path/to/file
	if strings.Contains(c.Value, "{.}") {
		dir := filepath.Dir(task.Task)
		base := filepath.Base(task.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		c.Value = strings.ReplaceAll(c.Value, "{.}", filepath.Join(dir, noExtension))
	}
	// /path/to/file.ext -> /path/to
	if strings.Contains(c.Value, "{/}") {
		dir := filepath.Dir(task.Task)
		c.Value = strings.ReplaceAll(c.Value, "{.}", dir)
	}
}

// Execute execute a shell command
func (c *Command) Execute() (stdout, stdErr string, err error) {
	var buffStdOut bytes.Buffer
	var buffStdErr bytes.Buffer

	stdOutMW := io.MultiWriter(os.Stdout, &buffStdOut)
	stdErrMW := io.MultiWriter(os.Stderr, &buffStdErr)

	cmd := exec.Command("bash", "-c", c.Value)
	cmd.Stdout = stdOutMW
	cmd.Stderr = stdErrMW

	err = cmd.Run()
	if err != nil {
		return
	}

	stdout = buffStdOut.String()
	stdErr = buffStdErr.String()

	return
}
