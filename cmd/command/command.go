package command

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/alessio/shellescape"
	"github.com/imarsman/goparallel/cmd/awk"
	"github.com/imarsman/goparallel/cmd/parse"
	"github.com/imarsman/goparallel/cmd/tasks"
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
	Awk         *awk.Awk // awk script to use
	Slots       int64
	DryRun      bool
	Ordered     bool
	KeepOrder   bool
	Concurrency int64
}

func init() {
}

// Command a command
type Command struct {
	Command  string
	Slots    int64
	Config   Config
	Sequence int64
}

// NewCommand create a new command struct instance
func NewCommand(value string, taskListSet *tasks.TaskListSet, config Config) Command {
	c := Command{
		Command:  value,
		Slots:    config.Slots,
		Config:   config,
		Sequence: 1,
	}

	return c
}

// SequenceSet set sequence
// For use with things like {2} where getting extra items is needed.
func (c *Command) SequenceSet(sequence int64) (err error) {
	// if sequence > int64(len(tls.TaskLists)-1) {
	// 	err = fmt.Errorf("sequence to set %d outside of max %d", len(tls.TaskLists)-1, sequence)
	// }
	atomic.AddInt64(&c.Sequence, sequence)

	return
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

var count int = 0

// RunCommand run all items in task lists against RunCommand
func RunCommand(c Command, taskSet []tasks.Task, wg *sync.WaitGroup) (err error) {
	count++
	// fmt.Println("count", count)

	ctx := context.Background()

	err = c.Prepare(taskSet)
	if err != nil {
		wg.Done()

		return
	}

	err = sem.Acquire(ctx, 1)
	if err != nil {
		wg.Done()

		panic(err)
	}

	var run = func() {
		defer wg.Done()
		defer sem.Release(1)

		err = c.Execute()
		if err != nil {
			wg.Done()
			return
		}
	}

	// Run in order (slower) or in parallel
	if c.Config.Ordered {
		run()
	} else {
		go run()
	}

	return
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
	defer c.GetSlotNumber()

	sequence := c.GetSequence()

	// tasks, err := c.TaskListSet.NextAll()
	if err != nil {
		return
	}

	defaultTask := tasks[0]

	if strings.TrimSpace(c.Command) == "" {
		c.Command = "echo"
	}

	if !strings.Contains(c.Command, `{`) {
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
		c.Command = fmt.Sprintf("%s %s", c.Command, strings.TrimSpace(sb.String()))
	}

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
		replaceToken(parse.TokenInputLine, shellescape.Quote(defaultTask.Task))
	}

	// {.}
	// Input line without extension.
	if strings.Contains(c.Command, parse.TokenInputLineNoExtension) {
		dir := filepath.Dir(defaultTask.Task)
		base := filepath.Base(defaultTask.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))
		replacement := filepath.Join(dir, noExtension)

		replaceToken(parse.TokenInputLineNoExtension, shellescape.Quote(replacement))
	}

	// {/}
	// Basename of input line.
	if strings.Contains(c.Command, parse.TokenBaseName) {
		replaceToken(parse.TokenBaseName, filepath.Base(shellescape.Quote(defaultTask.Task)))
	}

	// {//}
	// Dirname of output line.
	if strings.Contains(c.Command, parse.TokenDirname) {
		replaceToken(parse.TokenDirname, filepath.Dir(shellescape.Quote(defaultTask.Task)))
	}

	// {/.}
	// Basename of input line without extension.
	if strings.Contains(c.Command, parse.TokenBaseNameNoExtension) {
		base := filepath.Base(defaultTask.Task)
		noExtension := strings.TrimSuffix(base, filepath.Ext(base))

		// replaceToken(parse.TokenDirname, filepath.Dir(shellescape.Quote(defaultTask.Task)))
		replaceToken(parse.TokenBaseNameNoExtension, shellescape.Quote(noExtension))
	}
	// {#}
	// Sequence number of the job to run.
	if strings.Contains(c.Command, parse.TokenSequence) {
		replaceToken(parse.TokenSequence, shellescape.Quote(fmt.Sprint(sequence)))
	}

	// {%}
	// Job slot number.
	if strings.Contains(c.Command, parse.TokenSlot) {
		replaceToken(parse.TokenSlot, shellescape.Quote(fmt.Sprint(c.GetSlotNumber())))
	}

	if len(tasks) > 1 {
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

				c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d}`, number), shellescape.Quote(replacement))
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

				c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d.}`, number), shellescape.Quote(replacement))
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

				c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d/}`, number), shellescape.Quote(replacement))
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

				c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d//}`, number), filepath.Dir(task.Task))
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

				c.Command = strings.ReplaceAll(c.Command, fmt.Sprintf(`{%d./}`, number), shellescape.Quote(replacement))
			}
		}
	}

	return
}

// Execute execute a shell command
// For now, returns the stdout and stderr.
// Sends stdout and stderr to system stdout and stderr.
// func (c *Command) Execute() (stdout, stdErr string, err error) {
func (c *Command) Execute() (err error) {
	var buffStdOut bytes.Buffer
	var buffStdErr bytes.Buffer

	cmd := exec.Command("bash", "-c", c.Command)

	cmd.Stdout = &buffStdOut
	cmd.Stderr = &buffStdErr

	// If we are on a dry run print out what would be run, otherwise run the command.
	if !c.Config.DryRun {
		err = cmd.Run()
		if err != nil {
			fmt.Println(cmd.String())
			fmt.Println("got error on run", cmd.String(), err)
		}

	} else {
		fmt.Println(cmd.String())
	}

	// Make buffers for command output
	outStr := buffStdOut.String()
	errStr := buffStdErr.String()

	// if outStr != "" {
	c.Print(os.Stdout, outStr)
	// }
	if errStr != "" {
		c.Print(os.Stderr, errStr)
	}

	return
}

var mu sync.Mutex

// Print send to output
func (c *Command) Print(file *os.File, str string) {
	if !c.Config.KeepOrder {
		mu.Lock()
		defer mu.Unlock()
	}

	fmt.Fprintln(file, strings.TrimSpace(str))
}
