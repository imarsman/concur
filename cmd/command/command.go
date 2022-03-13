package command

import (
	"bytes"
	"errors"
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
	Command     string
	Concurrency int
	TaskListSet *tasks.TaskListSet
}

func RunCommand(c Command) {
	for i := 0; i < c.TaskListSet.Max(); i++ {
		// c2 := c.Copy()

	}
}

func NewCommand(value string, taskListSet *tasks.TaskListSet, concurrency int) Command {
	c := Command{Command: value, Concurrency: concurrency, TaskListSet: taskListSet}
	return c
}

func isValid(fp string) bool {
	// Check if file already exists
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	return false
}

// Copy get copy of a command to avoid overwriting the source copy
func (c *Command) Copy() (newCommand Command) {
	newCommand = *c

	return newCommand
}

// Prepare replace placeholders with data from incoming
func (c *Command) Prepare(sequence int64) (err error) {
	tasks, err := c.TaskListSet.NextAll()
	if err != nil {
		fmt.Println("error")
		return
	}

	if !isValid(tasks[0].Task) {
		err = errors.New("invalid file")
		return
	}

	defaultTask := tasks[0]
	// Input line
	// /path/to/file.ext interpolated
	if strings.Contains(c.Command, "{}") {
		c.Command = strings.ReplaceAll(c.Command, "{}", defaultTask.Task)
	}
	// Input line without extension
	// /path/to/file.ext -> /path/to/file
	if strings.Contains(c.Command, "{.}") {
		dir := filepath.Dir(defaultTask.Task)
		base := filepath.Base(defaultTask.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		c.Command = strings.ReplaceAll(c.Command, "{.}", filepath.Join(dir, noExtension))
	}
	// Basename of input line
	// /path/to/file.ext -> file.ext
	if strings.Contains(c.Command, "{/}") {
		c.Command = strings.ReplaceAll(c.Command, "{/}", filepath.Base(defaultTask.Task))
	}
	if strings.Contains(c.Command, "{//}") {
		c.Command = strings.ReplaceAll(c.Command, "{//}", filepath.Dir(defaultTask.Task))
	}
	// /path/to/file.ext -> /path/to/file
	if strings.Contains(c.Command, "{/.}") {
		base := filepath.Base(defaultTask.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		c.Command = strings.ReplaceAll(c.Command, "{/.}", noExtension)
	}
	//
	if strings.Contains(c.Command, "{#}") {
		c.Command = strings.ReplaceAll(c.Command, "{#}", fmt.Sprint(sequence))
	}
	//
	if strings.Contains(c.Command, "{%}") {
		var slotNumber = c.Concurrency
		if sequence > int64(c.Concurrency) {
			slotNumber = int(sequence) % c.Concurrency
		}
		c.Command = strings.ReplaceAll(c.Command, "{%}", fmt.Sprint(slotNumber))
	}

	return
}

// Execute execute a shell command
func (c *Command) Execute() (stdout, stdErr string, err error) {
	var buffStdOut bytes.Buffer
	var buffStdErr bytes.Buffer

	stdOutMW := io.MultiWriter(os.Stdout, &buffStdOut)
	stdErrMW := io.MultiWriter(os.Stderr, &buffStdErr)

	cmd := exec.Command("bash", "-c", c.Command)
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
