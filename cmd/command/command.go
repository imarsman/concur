package command

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/alessio/shellescape"
	"github.com/imarsman/concur/cmd/awk"
	"github.com/imarsman/concur/cmd/parse"
	"github.com/imarsman/concur/cmd/tasks"
	"golang.org/x/sync/semaphore"
)

var sem *semaphore.Weighted

func init() {
	sem = semaphore.NewWeighted(8)
}

var once sync.Once

// SetConcurrency allow concurrency to be set
func (c *Command) SetConcurrency(concurrency int64) {
	once.Do(func() {
		sem = semaphore.NewWeighted(concurrency)
	})
}

// Config config parameters
type Config struct {
	Awk         *awk.Command // awk script to use
	Slots       int64
	DryRun      bool
	KeepOrder   bool
	Concurrency int64
	PrintEmpty  bool
	ExitOnError bool
	StdIn       bool
}

// Command a command
type Command struct {
	Input    string
	Command  string
	Slots    int64
	Config   Config
	Sequence int64
	Empty    bool
}

// NewCommand create a new command struct instance
func NewCommand(value string, taskListSet *tasks.TaskListSet, config Config) Command {
	c := Command{
		Command:  strings.TrimSpace(value),
		Slots:    config.Slots,
		Config:   config,
		Sequence: 1,
	}

	return c
}

// GetSequence get lock free sequence value
func (c *Command) GetSequence() int64 {
	return atomic.LoadInt64(&c.Sequence)
}

// SequenceIncr increment sequence without lock
func (c *Command) SequenceIncr() {
	atomic.AddInt64(&c.Sequence, 1)
}

// SequenceReset reset sequence
func (c *Command) SequenceReset() {
	atomic.StoreInt64(&c.Sequence, 0)
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

// GetSlotNumber get slot number based on sequence and concurrency
func (c Command) GetSlotNumber() int64 {
	var slotNumber = c.Slots
	var sequence = c.Sequence
	if int64(c.Slots) <= sequence {
		slotNumber = sequence % c.Slots
	} else {
		slotNumber = sequence
	}
	if slotNumber == 0 {
		slotNumber = c.Slots
	}

	return slotNumber
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
func (c *Command) Prepare(tasks []tasks.Task) (err error) {
	var taskStrings []string
	for _, t := range tasks {
		taskStrings = append(taskStrings, t.Task)
	}
	c.Input = strings.Join(taskStrings, " ")

	defer c.GetSlotNumber()
	c.Empty = false

	sequence := c.GetSequence()

	if err != nil {
		return
	}

	defaultTask := tasks[0]
	// If there is something besides a token but no tokens we don't want to interpolate tokens later
	var commandStringEmpty = true

	// If empty, flag that
	if c.Command == "" {
		c.Empty = true
	} else {
		// allow no command to be run but line reformatted if the only space delimite things in the line are {} type
		// tokens
		c.Empty = true
		parts := strings.Split(c.Command, " ")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if !parse.REAllTokens.MatchString(part) {
				commandStringEmpty = false
				c.Empty = false
				break
			}
		}
	}

	// look for tokens except for {#} and {%}
	var foundToken = parse.REToken.MatchString(c.Command)

	// If no tokens, supply them
	// With an empty command the result will be the placement of the incoming value
	if !foundToken && !c.Config.StdIn {
		var sb strings.Builder
		if len(tasks) == 1 {
			_, err = sb.WriteString("{}")
			if err != nil {
				return
			}
		} else {
			for i := range tasks {
				_, err = sb.WriteString(fmt.Sprintf("{%d} ", i+1))
				if err != nil {
					return
				}
			}
		}

		tokenStr := strings.TrimSpace(sb.String())
		// If the incoming command started with no non token code
		if commandStringEmpty {
			if c.Command == "" {
				c.Command = fmt.Sprintf("%s", tokenStr)
			} else {
				c.Command = fmt.Sprintf("%s %s", c.Command, tokenStr)
			}
		} else {
			// Put in our placeholders
			if c.Command != "" {
				c.Command = fmt.Sprintf("%s %s", c.Command, tokenStr)
			}
		}
	}

	// replaceToken replace a token with a replacement string for the command
	var replaceToken = func(pattern string, replace string) {
		if len(tasks) == 1 {
			for strings.Contains(c.Command, pattern) {
				c.Command = strings.Replace(c.Command, pattern, replace, 1)
			}
		} else {
			for strings.Contains(c.Command, pattern) {
				c.Command = strings.ReplaceAll(c.Command, pattern, replace)
			}
		}
	}

	// {}
	// Input line
	if strings.Contains(c.Command, parse.TokenInputLine) {
		replacement := defaultTask.Task

		if c.Empty {
			replaceToken(parse.TokenInputLine, replacement)
		} else {
			replaceToken(parse.TokenInputLine, shellescape.Quote(replacement))
		}
	}

	// {.}
	// Input line without extension.
	if strings.Contains(c.Command, parse.TokenInputLineNoExtension) {
		dir := filepath.Dir(defaultTask.Task)
		base := filepath.Base(defaultTask.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		replacement := filepath.Join(dir, noExtension)

		if c.Empty {
			replaceToken(parse.TokenInputLine, replacement)
		} else {
			replaceToken(parse.TokenInputLine, shellescape.Quote(replacement))
		}
	}

	// {/}
	// Basename of input line.
	if strings.Contains(c.Command, parse.TokenBaseName) {
		replacement := defaultTask.Task

		if c.Empty {
			replaceToken(parse.TokenBaseName, replacement)
		} else {
			replaceToken(parse.TokenBaseName, shellescape.Quote(replacement))
		}
	}

	// {//}
	// Dirname of output line.
	if strings.Contains(c.Command, parse.TokenDirname) {
		replacement := defaultTask.Task

		if c.Empty {
			replaceToken(parse.TokenDirname, replacement)
		} else {
			replaceToken(parse.TokenDirname, shellescape.Quote(replacement))
		}

		replaceToken(parse.TokenDirname, filepath.Dir(shellescape.Quote(defaultTask.Task)))
	}

	// {/.}
	// Basename of input line without extension.
	if strings.Contains(c.Command, parse.TokenBaseNameNoExtension) {
		replacement := defaultTask.Task

		base := filepath.Base(replacement)
		replacement = strings.TrimSuffix(base, filepath.Ext(base))

		if c.Empty {
			replaceToken(parse.TokenBaseNameNoExtension, replacement)
		} else {
			replaceToken(parse.TokenBaseNameNoExtension, shellescape.Quote(replacement))
		}
	}
	// {#}
	// Sequence number of the job to run.
	if strings.Contains(c.Command, parse.TokenSequence) {
		replacement := fmt.Sprint(sequence)
		if c.Empty {
			replaceToken(parse.TokenSequence, replacement)
		} else {
			replaceToken(parse.TokenSequence, shellescape.Quote(replacement))
		}
	}

	// {%}
	// Job slot number.
	if strings.Contains(c.Command, parse.TokenSlot) {
		replacement := fmt.Sprint(c.GetSlotNumber())
		if c.Empty {
			replaceToken(parse.TokenSlot, replacement)
		} else {
			replaceToken(parse.TokenSlot, shellescape.Quote(replacement))
		}
	}

	// Remove this after ensuring it is OK
	if len(tasks) >= 1 {
		var found bool
		var number int

		// {n}
		// Argument from input source n or the n'th argument.
		// Note - nth argument handling not implemented.
		found, number, err = parse.NumberFromToken(parse.RENumbered, c.Command)
		if err != nil {
			return
		}
		if found {
			for {
				found, number, err = parse.NumberFromToken(parse.RENumbered, c.Command)
				if err != nil {
					return
				}
				if !found {
					break
				}

				if len(tasks) < number {
					err = fmt.Errorf(
						"task item {%d} for task list count %d out of range",
						number,
						len(tasks),
					)
					return
				}
				if number > len(tasks) {
					break
				}
				task := tasks[number-1]

				// Avoid endless loop
				if parse.RENumbered.MatchString(task.Task) {
					err = fmt.Errorf(
						"item %s matches regular expression %s",
						task.Task,
						parse.RENumbered.String(),
					)
					return
				}

				replacement := task.Task

				if c.Empty {
					c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d}`, number), replacement)
				} else {
					c.Command = strings.ReplaceAll(
						c.Command, fmt.Sprintf(`{%d}`, number), shellescape.Quote(replacement),
					)
				}
			}
		}

		// {n.}
		// Argument from input source n or the n'th argument without extension.
		// Note - nth argument handling not implemented.
		found, number, err = parse.NumberFromToken(parse.RENumberedWithNoExtension, c.Command)
		if err != nil {
			return
		}
		if found {
			for {
				found, number, err = parse.NumberFromToken(parse.RENumberedWithNoExtension, c.Command)
				if err != nil {
					return
				}
				if !found {
					break
				}

				if len(tasks) < number {
					err = fmt.Errorf(
						"task item {%d.} for task list count %d out of range",
						number,
						len(tasks),
					)
					return
				}

				if number > len(tasks) {
					break
				}
				task := tasks[number-1]

				dir := filepath.Dir(task.Task)
				base := filepath.Base(task.Task)
				noExtension := strings.TrimSuffix(base, filepath.Ext(base))
				replacement := filepath.Join(dir, noExtension)

				// Avoid endless loop
				if parse.RENumberedWithNoExtension.MatchString(task.Task) {
					err = fmt.Errorf("item %s matches regular expression %s", replacement, parse.RENumberedWithNoExtension.String())
					return
				}

				if c.Empty {
					c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d.}`, number), replacement)
				} else {
					c.Command = strings.ReplaceAll(
						c.Command, fmt.Sprintf(`{%d.}`, number), shellescape.Quote(replacement),
					)
				}
			}
		}

		// {n/}
		// Basename of argument from input source n or the n'th argument.
		// Note - nth argument handling not implemented.
		found, number, err = parse.NumberFromToken(parse.RENumberedBasename, c.Command)
		if err != nil {
			return
		}
		if found {
			for {
				found, number, err = parse.NumberFromToken(parse.RENumberedBasename, c.Command)
				if err != nil {
					return
				}
				if !found {
					break
				}

				if len(tasks) < number {
					err = fmt.Errorf(
						"task item {%d/} for task list count %d out of range",
						number,
						len(tasks),
					)
					return
				}
				if number > len(tasks) {
					break
				}
				task := tasks[number-1]

				replacement := filepath.Base(task.Task)

				// Avoid endless loop
				if parse.RENumberedBasename.MatchString(task.Task) {
					err = fmt.Errorf(
						"item %s matches regular expression %s",
						replacement,
						parse.RENumberedBasename.String(),
					)
					return
				}
				if c.Empty {
					c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d/}`, number), replacement)
				} else {
					c.Command = strings.ReplaceAll(
						c.Command, fmt.Sprintf(`{%d/}`, number), shellescape.Quote(replacement),
					)
				}
			}
		}

		// {n//}
		// Dirname of argument from input source n or the n'th argument.
		// Note - nth argument handling not implemented.
		found, number, err = parse.NumberFromToken(parse.RENumberedDirname, c.Command)
		if err != nil {
			return
		}
		if found {
			for {
				found, number, err = parse.NumberFromToken(parse.RENumberedDirname, c.Command)
				if err != nil {
					return
				}
				if !found {
					break
				}

				if len(tasks) < number {
					err = fmt.Errorf(
						"task item {%d//} for task list count %d out of range",
						number,
						len(tasks),
					)
					return
				}
				if number > len(tasks) {
					break
				}
				task := tasks[number-1]

				replacent := filepath.Dir(task.Task)

				// Avoid endless loop
				if parse.RENumberedDirname.MatchString(task.Task) {
					err = fmt.Errorf(
						"item %s matches regular expression %s",
						task.Task, replacent,
					)
					return
				}
				replacement := filepath.Dir(task.Task)

				if c.Empty {
					c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d//}`, number), replacement)
				} else {
					c.Command = strings.ReplaceAll(
						c.Command, fmt.Sprintf(`{%d//}`, number), shellescape.Quote(replacement),
					)
				}
			}
		}

		// {n/.}
		// Basename of argument from input source n or the n'th argument without extension.
		// Note - nth argument handling not implemented.
		found, number, err = parse.NumberFromToken(parse.RENumberedBasenameNoExtension, c.Command)
		if err != nil {
			return
		}
		if found {
			for {
				found, number, err = parse.NumberFromToken(parse.RENumberedBasenameNoExtension, c.Command)
				if err != nil {
					return
				}
				if !found {
					break
				}

				if len(tasks) < number {
					err = fmt.Errorf(
						"task item {%d/.} for task list count %d out of range",
						number,
						len(tasks),
					)
					return
				}
				if number > len(tasks) {
					break
				}
				task := tasks[number-1]

				base := filepath.Base(task.Task)
				replacement := strings.TrimSuffix(base, filepath.Ext(base))

				// Avoid endless loop
				if parse.RENumberedBasenameNoExtension.MatchString(task.Task) {
					err = fmt.Errorf("item %s matches regular expression %s", replacement, parse.RENumberedBasenameNoExtension.String())
					return
				}

				if c.Empty {
					c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d./}`, number), replacement)
				} else {
					c.Command = strings.ReplaceAll(
						c.Command, fmt.Sprintf(`{%d./}`, number), shellescape.Quote(replacement),
					)
				}
			}
		}
	}

	return
}

// RunCommand run all items in task lists against RunCommand
func RunCommand(c Command, taskSet []tasks.Task, wg *sync.WaitGroup) (err error) {

	ctx := context.Background()

	err = c.Prepare(taskSet)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Acquire weight of one from semaphore
	err = sem.Acquire(ctx, 1)
	if err != nil {
		panic(err)
	}

	// Run and release waitgroup for overall processing and semaphore for concurrency
	var run = func() {
		defer wg.Done()
		defer sem.Release(1)
		err = c.Execute()
		if err != nil {
			return
		}
	}

	// Run as go process. Will run ordered if semaphore count set to 1
	go run()

	return
}

// Execute execute a shell command
// For now, returns the stdout and stderr.
// Sends stdout and stderr to system stdout and stderr.
// func (c *Command) Execute() (stdout, stdErr string, err error) {
func (c *Command) Execute() (err error) {
	outStr := c.Command
	errStr := ""

	// If the command started out as "" don't try to run command, otherwise run
	if !c.Empty {
		var buffStdOut bytes.Buffer
		var buffStdErr bytes.Buffer

		cmd := exec.Command("bash", "-c", c.Command)

		// If stdin was specified, send the input to the command's stdin
		if c.Config.StdIn {
			// https://stackoverflow.com/questions/23166468/how-can-i-get-stdin-to-exec-cmd-in-golang
			stdin, stdinErr := cmd.StdinPipe()
			if stdinErr != nil {
				err = stdinErr
				return
			}

			go func() {
				defer stdin.Close()
				io.WriteString(stdin, c.Input)
			}()
		}

		cmd.Stdout = &buffStdOut
		cmd.Stderr = &buffStdErr
		// If we are on a dry run print out what would be run, otherwise run the command.
		if !c.Config.DryRun {
			err = cmd.Run()
			if err != nil {
				if c.Config.ExitOnError {
					c.Print(os.Stderr, fmt.Sprintf("%v", err))
					os.Exit(1)
				}
			}
		} else {
			// with dry-run print out command and return
			fmt.Println(strings.TrimSpace(cmd.String()))
			return
		}
		// Don't print anything now. Wait until awk script stage.
		outStr = buffStdOut.String()
		errStr = buffStdErr.String()
	}

	// Run awk against what has been produced so far
	// Print out result
	if c.Config.Awk != nil {
		outStr, err = c.Config.Awk.Execute(outStr)
		if err != nil {
			errStr := fmt.Sprintf("%v", err)
			c.Print(os.Stderr, errStr)
			if c.Config.ExitOnError {
				os.Exit(1)
			}
		}
		if outStr == "" && c.Config.PrintEmpty {
			if c.Config.PrintEmpty {
				c.Print(os.Stdout, outStr)
			} else {
				return
			}
		}
		if outStr != "" {
			c.Print(os.Stdout, outStr)
		}
	} else {
		// No awk script so print output from command run
		if len(outStr) > 0 {
			c.Print(os.Stdout, outStr)
		} else if len(outStr) == 0 {
			if c.Config.PrintEmpty {
				c.Print(os.Stdout, outStr)
			}
		}
		if errStr != "" {
			c.Print(os.Stderr, errStr)
		}
	}

	return
}

var printWG = new(sync.WaitGroup)

// Print send to output
func (c *Command) Print(file *os.File, str string) {
	// Optional waitgroup for printing output in order
	printWG.Wait()
	if c.Config.KeepOrder {
		printWG.Add(1)
		defer printWG.Done()
	}
	fmt.Fprintln(file, strings.TrimSpace(str))
}
