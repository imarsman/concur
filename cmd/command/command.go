package command

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/imarsman/goparallel/cmd/tasks"
	"golang.org/x/sync/semaphore"
)

var slots int
var sem *semaphore.Weighted

func init() {
	slots = 8
	sem = semaphore.NewWeighted(int64(slots))
}

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

func NewCommand(value string, taskListSet *tasks.TaskListSet) Command {
	c := Command{Command: value, Concurrency: slots, TaskListSet: taskListSet}
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

var reNumbered = regexp.MustCompile(`\{(?P<NUMBERED>\d+)\}`)
var reNumberedWithNoExtension = regexp.MustCompile(`\{(?P<NUMBERED>\d)+\.\}`)
var reNumberedBasename = regexp.MustCompile(`\{(?P<NUMBERED>\d+)\/\}`)
var reNumberedDirname = regexp.MustCompile(`\{(?P<NUMBERED>\d+)\/\/\}`)
var reNumberedBasenameNoExtension = regexp.MustCompile(`\{(?P<NUMBERED>\d+)\/\.\}`)

// Prepare replace placeholders with data from incoming
func (c *Command) Prepare() (err error) {
	sequence := c.TaskListSet.GetSequence()

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
	// sequence number for the command
	if strings.Contains(c.Command, "{#}") {
		c.Command = strings.ReplaceAll(c.Command, "{#}", fmt.Sprint(sequence+1))
	}
	// slot number for the command
	if strings.Contains(c.Command, "{%}") {
		var slotNumber = c.Concurrency
		if int64(c.Concurrency) <= sequence {
			slotNumber = int(sequence)%c.Concurrency + 1
		} else {
			slotNumber = int(sequence) + 1
		}
		c.Command = strings.ReplaceAll(c.Command, "{%}", fmt.Sprint(slotNumber))
	}

	numberedParams := getParams(reNumbered, c.Command)
	// fmt.Println(numberedParams)
	if numberedParams["NUMBERED"] != "" {
		numbered := numberedParams["NUMBERED"]
		var number int
		number, err = strconv.Atoi(numbered)
		if err != nil {
			return
		}
		if number-1 > len(tasks) {
			err = errors.New("out of range")
		}
		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d}`, number), task.Task)
	}

	// TODO: Implement properly
	numberedParamsNoExtensionParams := getParams(reNumberedWithNoExtension, c.Command)
	// fmt.Println(numberedParams)
	if numberedParams["NUMBERED"] != "" {
		numbered := numberedParamsNoExtensionParams["NUMBERED"]
		var number int
		number, err = strconv.Atoi(numbered)
		if err != nil {
			return
		}
		if number-1 > len(tasks) {
			err = errors.New("out of range")
		}
		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d.}`, number), task.Task)
	}

	// TODO: Implement properly
	reNumberedBasenameParams := getParams(reNumberedBasename, c.Command)
	// fmt.Println(numberedParams)
	if numberedParams["NUMBERED"] != "" {
		numbered := reNumberedBasenameParams["NUMBERED"]
		var number int
		number, err = strconv.Atoi(numbered)
		if err != nil {
			return
		}
		if number-1 > len(tasks) {
			err = errors.New("out of range")
		}
		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d/}`, number), task.Task)
	}

	// TODO: Implement properly
	reNumberedDirnameParams := getParams(reNumberedDirname, c.Command)
	// fmt.Println(numberedParams)
	if numberedParams["NUMBERED"] != "" {
		numbered := reNumberedDirnameParams["NUMBERED"]
		var number int
		number, err = strconv.Atoi(numbered)
		if err != nil {
			return
		}
		if number-1 > len(tasks) {
			err = errors.New("out of range")
		}
		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d//}`, number), task.Task)
	}

	// TODO: Implement properly
	reNumberedBasenameNoExtensionParams := getParams(reNumberedBasenameNoExtension, c.Command)
	// fmt.Println(numberedParams)
	if numberedParams["NUMBERED"] != "" {
		numbered := reNumberedBasenameNoExtensionParams["NUMBERED"]
		var number int
		number, err = strconv.Atoi(numbered)
		if err != nil {
			return
		}
		if number-1 > len(tasks) {
			err = errors.New("out of range")
		}
		task := tasks[number-1]
		c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d//}`, number), task.Task)
	}

	c.TaskListSet.SequenceIncr()

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
