package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/imarsman/goparallel/cmd/parse"
	"github.com/imarsman/goparallel/cmd/tasks"
	"golang.org/x/sync/semaphore"
)

// var slots int
// var sem *semaphore.Weighted
// var ctx context.Context

func init() {
	// slots = 8
	// ctx = context.TODO()
	// sem = semaphore.NewWeighted(int64(slots))
}

// Command a command
type Command struct {
	Command     string
	Concurrency int
	TaskListSet *tasks.TaskListSet
	DryRun      bool
}

// RunCommand run all items in task lists against RunCommand
func RunCommand(c Command) {
	ctx := context.TODO()
	sem := semaphore.NewWeighted(int64(c.Concurrency))

	var wg sync.WaitGroup
	for i := 0; i < c.TaskListSet.Max(); i++ {
		// Add delta of 1 to waitgroup
		wg.Add(1)

		// Copying
		// -------
		// TaskListSet is a pointer so remains.
		// Command is a string and gets copied to the new one from the incoming
		// Other attributes are not changed.
		c2 := c.Copy()

		atEnd, err := c2.Prepare()
		if err != nil {
			fmt.Println(err)
		}
		// We have reached the end of the list of input items
		// This can happen when more than one input item is used in a command.
		// e.g. echo {1} {2}
		// This is not implemented yet.
		if atEnd && i < c.TaskListSet.Max() {
			break
		}

		err = sem.Acquire(ctx, 1)
		if err != nil {
			panic(err)
		}
		go func() {
			defer sem.Release(1)
			defer wg.Done()
			_, _, err = c2.Execute()
			if err != nil {
				wg.Done()
				panic(err)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

// NewCommand create a new command struct instance
func NewCommand(value string, taskListSet *tasks.TaskListSet, slots int, dryRun bool) Command {
	c := Command{Command: value, Concurrency: slots, TaskListSet: taskListSet, DryRun: dryRun}

	return c
}

// is a file valid - not using currently as it will cause un-needed failures
func isValid(fp string) bool {
	// Check if file already exists
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	return false
}

// Copy get copy of a command to avoid overwriting the source copy of the
// Command attribute
func (c *Command) Copy() (newCommand Command) {
	newCommand = *c

	return newCommand
}

/**
 * Parses url with the given regular expression and returns the
 * group values defined in the expression.
 *
 */
func getParams(regEx *regexp.Regexp, input string) (paramsMap map[string]string) {
	match := regEx.FindStringSubmatch(input)

	paramsMap = make(map[string]string)
	for i, name := range regEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

// Prepare replace placeholders with data from incoming
func (c *Command) Prepare() (atEnd bool, err error) {
	sequence := c.TaskListSet.GetSequence()

	tasks, atEnd, err := c.TaskListSet.NextAll()
	if err != nil {
		fmt.Println("error")
		return
	}

	// This will cause un-needed errors
	// Properness of outgoing data is the responsibility of the call creator
	// if !isValid(tasks[0].Task) {
	// 	err = errors.New("invalid file")
	// 	return
	// }

	defaultTask := tasks[0]

	// {}
	// Input line
	if strings.Contains(c.Command, "{}") {
		c.Command = strings.ReplaceAll(c.Command, "{}", defaultTask.Task)
	}

	// {.}
	// Input line without extension.
	if strings.Contains(c.Command, "{.}") {
		dir := filepath.Dir(defaultTask.Task)
		base := filepath.Base(defaultTask.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		c.Command = strings.ReplaceAll(c.Command, "{.}", filepath.Join(dir, noExtension))
	}

	// {/}
	// Basename of input line.
	if strings.Contains(c.Command, "{/}") {
		c.Command = strings.ReplaceAll(c.Command, "{/}", filepath.Base(defaultTask.Task))
	}

	// {//}
	// Dirname of output line.
	if strings.Contains(c.Command, "{//}") {
		c.Command = strings.ReplaceAll(c.Command, "{//}", filepath.Dir(defaultTask.Task))
	}

	// {/.}
	// Basename of input line without extension.
	if strings.Contains(c.Command, "{/.}") {
		base := filepath.Base(defaultTask.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		c.Command = strings.ReplaceAll(c.Command, "{/.}", noExtension)
	}
	// {#}
	// Sequence number of the job to run.
	if strings.Contains(c.Command, "{#}") {
		c.Command = strings.ReplaceAll(c.Command, "{#}", fmt.Sprint(sequence))
	}

	// {%}
	// Job slot number.
	if strings.Contains(c.Command, "{%}") {
		var slotNumber = c.Concurrency
		if int64(c.Concurrency) <= sequence {
			slotNumber = int(sequence)%c.Concurrency + 1
		} else {
			slotNumber = int(sequence) + 1
		}
		c.Command = strings.ReplaceAll(c.Command, "{%}", fmt.Sprint(slotNumber))
	}

	// {n}
	// Argument from input source n or the n'th argument.
	// Note - nth argument handling not implemented.
	found, number, err := parse.NumberFromToken(parse.RENumbered, c.Command)
	if err != nil {
		return
	}
	if found {
		if number-1 > len(tasks) {
			err = errors.New("out of range")
		}
		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d}`, number), task.Task)
	}

	// {n.}
	// Argument from input source n or the n'th argument without extension.
	// Note - nth argument handling not implemented.
	found, number, err = parse.NumberFromToken(parse.RENumberedWithNoExtension, c.Command)
	if err != nil {
		return
	}
	if found {
		if number-1 > len(tasks) {
			err = errors.New("out of range")
			return
		}
		task := tasks[number-1]

		dir := filepath.Dir(task.Task)
		base := filepath.Base(task.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		replacement := filepath.Join(dir, noExtension)

		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d.}`, number), replacement)
	}

	// {n/}
	// Basename of argument from input source n or the n'th argument.
	// Note - nth argument handling not implemented.
	found, number, err = parse.NumberFromToken(parse.RENumberedBasename, c.Command)
	if err != nil {
		return
	}
	if found {
		if number-1 > len(tasks) {
			err = errors.New("out of range")
			return
		}

		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d/}`, number), filepath.Base(task.Task))
	}

	// {n//}
	// Dirname of argument from input source n or the n'th argument.
	// Note - nth argument handling not implemented.
	found, number, err = parse.NumberFromToken(parse.RENumberedDirname, c.Command)
	if err != nil {
		return
	}
	if found {
		if number-1 > len(tasks) {
			err = errors.New("out of range")
			return
		}

		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d//}`, number), filepath.Dir(task.Task))
	}

	// {n/.}
	// Basename of argument from input source n or the n'th argument without extension.
	// Note - nth argument handling not implemented.
	found, number, err = parse.NumberFromToken(parse.RENumberedBasenameNoExtension, c.Command)
	if err != nil {
		return
	}
	if found {
		if number-1 > len(tasks) {
			err = errors.New("out of range")
		}

		task := tasks[number-1]
		base := filepath.Base(task.Task)
		replacement := strings.TrimSuffix(base, filepath.Ext(base))

		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d//}`, number), replacement)
	}

	c.TaskListSet.SequenceIncr()

	return
}

// Execute execute a shell command
// For now, returns the stdout and stderr.
// Sends stdout and stderr to system stdout and stderr.
func (c *Command) Execute() (stdout, stdErr string, err error) {
	var buffStdOut bytes.Buffer
	var buffStdErr bytes.Buffer

	stdOutMW := io.MultiWriter(os.Stdout, &buffStdOut)
	stdErrMW := io.MultiWriter(os.Stderr, &buffStdErr)

	cmd := exec.Command("bash", "-c", c.Command)
	cmd.Stdout = stdOutMW
	cmd.Stderr = stdErrMW

	if !c.DryRun {
		err = cmd.Run()
		if err != nil {
			return
		}

		stdout = buffStdOut.String()
		stdErr = buffStdErr.String()
	} else {
		fmt.Println(cmd.String())
	}

	return
}
