package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/imarsman/concur/cmd/awk"
	"github.com/imarsman/concur/cmd/command"
	"github.com/imarsman/concur/cmd/parse"
	"github.com/imarsman/concur/cmd/tasks"
	"github.com/posener/complete/v2"
	"github.com/posener/complete/v2/predict"
)

var (
	// GitCommit build flag
	GitCommit   string
	GitLastTag  string
	GitExactTag string
	// // CompilationDate build flag
	// CompilationDate string
	// CommitDate the date of latest commit
	Date string
)

var slots int

func init() {
	slots = 8
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Args command line arguments
type Args struct {
	Command     string   `arg:"positional"`
	Arguments   []string `arg:"-a,--arguments,separate" help:"lists of arguments"`
	Awk         string   `arg:"-A,--awk" help:"process using awk script or a script filename."`
	DryRun      bool     `arg:"-d,--dry-run" help:"show command to run but don't run"`
	Slots       int64    `arg:"-s,--slots" default:"8" help:"number of parallel tasks"`
	Shuffle     bool     `arg:"-S,--shuffle" help:"shuffle tasks prior to running"`
	Ordered     bool     `arg:"-o,--ordered" help:"run tasks in their incoming order"`
	KeepOrder   bool     `arg:"-k,--keep-order" help:"don't keep output for calls separate"`
	PrintEmpty  bool     `arg:"-P,--print-empty" help:"print empty lines"`
	ExitOnError bool     `arg:"-E,--exit-on-error" help:"exit on first error"`
	SplitAtNull bool     `arg:"-0,--null" help:"split at null character"`
	IgnoreError bool     `arg:"-i,--ignore-error" help:"Ignore errors"`
	StdIn       bool     `arg:"-I,--stdin" help:"send input to stdin"`
}

// Version get version information
func (Args) Version() string {
	var buf = new(bytes.Buffer)

	msg := "concur"
	buf.WriteString(fmt.Sprintln(msg))
	buf.WriteString(fmt.Sprintln(strings.Repeat("-", len(msg))))

	if GitCommit != "" {
		buf.WriteString(fmt.Sprintf("Commit: %13s\n", GitCommit))
	}
	// if Date != "" {
	// 	buf.WriteString(fmt.Sprintf("Commit date: %26s\n", Date))
	// }
	if Date != "" {
		buf.WriteString(fmt.Sprintf("Compile Date: %12s\n", Date))
	}

	return buf.String()
}

func main() {
	// Create the complete command.
	// Here we define completion values for each flag.
	cmd := &complete.Command{
		Flags: map[string]complete.Predictor{
			"arguments":     predict.Nothing,
			"awk":           predict.Nothing,
			"dry-run":       predict.Nothing,
			"slots":         predict.Nothing,
			"shuffle":       predict.Nothing,
			"ordered":       predict.Nothing,
			"keep-order":    predict.Nothing,
			"print-empty":   predict.Nothing,
			"exit-on-error": predict.Nothing,
			"null":          predict.Nothing,
			"ignore-error":  predict.Nothing,
			"stdin":         predict.Nothing,
		},
	}

	cmd.Complete("concur")

	var callArgs = Args{}

	arg.MustParse(&callArgs)

	// Slots are handled by a semaphore. If slots are set to 1 ordered processing is forced.
	if callArgs.Ordered {
		callArgs.Slots = 1
	}
	// Avoid locking things up
	if callArgs.Slots == 0 {
		callArgs.Slots = 1
	}

	var awkCommand *awk.Command
	if callArgs.Awk != "" {
		awkScript := callArgs.Awk
		// If there is a space in the value it is probably not a file
		if !strings.Contains(callArgs.Awk, "{") {
			if _, err := os.Stat(callArgs.Awk); err == nil {
				b, err := ioutil.ReadFile(callArgs.Awk)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				awkScript = string(b)
			}
		}
		var err error
		awkCommand, err = awk.NewCommand(awkScript)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if callArgs.Slots == 0 {
		callArgs.Slots = int64(runtime.NumCPU())
	}

	// Make config to hold various parameters
	config := command.Config{
		Slots:       callArgs.Slots,
		DryRun:      callArgs.DryRun,
		KeepOrder:   callArgs.KeepOrder,
		Concurrency: callArgs.Slots,
		Awk:         awkCommand,
		PrintEmpty:  callArgs.PrintEmpty,
		ExitOnError: callArgs.ExitOnError,
		StdIn:       callArgs.StdIn,
	}

	taskListSet := tasks.NewTaskListSet()

	// Define command to run
	var c = command.NewCommand(
		callArgs.Command,
		&taskListSet,
		config,
	)

	c.SetConcurrency(callArgs.Slots)
	var wg = new(sync.WaitGroup)

	var foundArgumentList = false
	if len(callArgs.Arguments) > 0 {
		foundArgumentList = true
		for _, v := range callArgs.Arguments {
			taskList := tasks.NewTaskList()
			parts := strings.Split(v, " ")

			// TODO: handle reading file lines if that makes sense
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if parse.RERange.MatchString(part) {
					items, err := parse.Range(part)
					if err != nil {
						fmt.Println(err)
						return
					}
					taskList.Add(items...)
				} else {
					baseDir := filepath.Dir(part)
					matches, err := filepath.Glob(part)
					if err != nil {
						continue
					}
					if len(matches) == 0 {
						taskList.Add(strings.TrimSpace(part))
					} else {
						var files []string
						for _, f := range matches {
							f, _ := os.Stat(f)
							if !f.IsDir() {
								files = append(files, filepath.Join(baseDir, f.Name()))
							}
						}
						taskList.Add(files...)
					}
				}
				if callArgs.Shuffle {
					taskList.Shuffle()
				}
			}
			taskListSet.AddTaskList(taskList)
		}
	}

	stdin := false

	// splitAtNull split at null terminator
	var splitAtNull = func(input []byte, atEOF bool) (advance int, token []byte, err error) {
		searchBytes := []byte("\000")
		searchLen := len(searchBytes)
		dataLen := len(input)

		// Return nothing if at end of file and no data passed
		if atEOF && dataLen == 0 {
			return 0, nil, nil
		}

		// Find next separator and return token
		if i := bytes.Index(input, searchBytes); i >= 0 {
			return i + searchLen, input[0:i], nil
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return dataLen, input, nil
		}

		// Request more data.
		return 0, nil, nil
	}

	// Use stdin if it is available
	// It will be the first task list if it is available. If there are other task lists they can be used as additional
	// task items.
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		stdin = true
		var scanner = bufio.NewScanner(os.Stdin)

		// Tell scanner to scan by lines.
		if callArgs.SplitAtNull {
			scanner.Split(splitAtNull)
		} else {
			scanner.Split(bufio.ScanLines)
		}

		var item string
		for scanner.Scan() {
			item = scanner.Text()
			item = strings.TrimSpace(item)
			// If we have just stdin and no -a lists handle them as they come.
			if len(item) == 0 {
				// Print out empty lines if that has been flagged
				if callArgs.PrintEmpty {
					c.Print(os.Stdout, "")
				}
				continue
			}
			var task = tasks.NewTask(item)
			var taskSet []tasks.Task
			taskSet = append(taskSet, *task)

			if foundArgumentList {
				newTasks, err := taskListSet.NextAll()
				if err != nil {
				}
				taskSet = append(taskSet, newTasks...)
			}
			c2 := c.Copy()
			wg.Add(1)
			err := command.RunCommand(c2, taskSet, wg)
			if err != nil {
				fmt.Println("got error", err)
				os.Exit(1)
			}
			c.SequenceIncr()
		}
	}

	// If we are not getting stdin run through and process all non-stdin list items
	if !stdin {
		// Run through as many iterations as the longest list
		for i := 0; i < taskListSet.Max(); i++ {
			wg.Add(1)
			tasks, err := taskListSet.NextAll()

			empty := true
			for _, t := range tasks {
				if len(strings.TrimSpace(t.Task)) > 0 {
					empty = false
					continue
				}
			}

			c2 := c.Copy()
			if empty {
				if callArgs.PrintEmpty {
					c2.Print(os.Stdout, "")
				}
				continue
			}

			err = command.RunCommand(c2, tasks, wg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			c.SequenceIncr()
		}
	}

	wg.Wait()
}
