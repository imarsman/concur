package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/imarsman/goparallel/cmd/awk"
	"github.com/imarsman/goparallel/cmd/command"
	"github.com/imarsman/goparallel/cmd/parse"
	"github.com/imarsman/goparallel/cmd/tasks"
)

var slots int

func init() {
	slots = 8
}

// args CLI args

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

// callArgs command line arguments
var callArgs struct {
	Command    string   `arg:"positional"`
	Arguments  []string `arg:"-a,--arguments,separate" help:"lists of arguments"`
	Awk        string   `arg:"-A,--awk" help:"process using supplied awk script"`
	DryRun     bool     `arg:"-d,--dry-run" help:"show command to run but don't run"`
	Slots      int64    `arg:"-s,--slots" default:"8" help:"number of parallel tasks"`
	Shuffle    bool     `arg:"-S,--shuffle" help:"shuffle tasks prior to running"`
	Ordered    bool     `arg:"-o,--ordered" help:"run tasks in their incoming order"`
	KeepOrder  bool     `arg:"-k,--keep-order" help:"don't keep output for calls separate"`
	PrintEmpty bool     `arg:"-E,--print-empty" help:"print empty lines"`
}

func main() {
	arg.MustParse(&callArgs)

	var awkCommand *awk.Command
	if callArgs.Awk != "" {
		var err error
		awkCommand, err = awk.NewCommand(callArgs.Awk)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if callArgs.Slots == 0 {
		callArgs.Slots = int64(runtime.NumCPU())
	}

	taskListSet := tasks.NewTaskListSet()

	// Make config to hold various parameters
	config := command.Config{
		Slots:       callArgs.Slots,
		DryRun:      callArgs.DryRun,
		Ordered:     callArgs.Ordered,
		KeepOrder:   callArgs.KeepOrder,
		Concurrency: callArgs.Slots,
		Awk:         awkCommand,
	}

	// if callArgs.Command == "" {
	// 	callArgs.Command = `echo`
	// }

	// Define command to run
	var c = command.NewCommand(
		callArgs.Command,
		&taskListSet,
		config,
	)

	c.SetConcurrency(callArgs.Slots)
	var wg = new(sync.WaitGroup)

	// Use stdin if it is available
	// It will be the first task list if it is available
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		stdinItems := []string{}
		var scanner = bufio.NewScanner(os.Stdin)

		// Tell scanner to scan by lines.
		scanner.Split(bufio.ScanLines)

		var item string
		for scanner.Scan() {
			item = scanner.Text()
			item = strings.TrimSpace(item)
			// fmt.Println(item)
			if len(callArgs.Arguments) == 0 {
				if len(strings.TrimSpace(item)) == 0 {
					if callArgs.PrintEmpty {
						c.Print(os.Stdout, "")
					}
					continue
				}
				var task = tasks.NewTask(item)
				var taskSet []tasks.Task
				taskSet = append(taskSet, *task)
				c2 := c.Copy()
				wg.Add(1)
				err := command.RunCommand(c2, taskSet, wg)
				if err != nil {
					fmt.Println("got error", err)
					os.Exit(1)
				}
				c.SequenceIncr()

			} else {
				stdinItems = append(stdinItems, item)
			}
		}
		if len(stdinItems) > 0 {
			taskList := tasks.NewTaskList()
			taskList.Add(stdinItems...)
			taskListSet.AddTaskList(taskList)
		} else {
			wg.Wait()
			return
		}
	}
	if len(callArgs.Arguments) == 0 {
		// Wait for all goroutines to complete
		wg.Wait()

		os.Exit(0)
	}

	if len(callArgs.Arguments) > 0 {
		// Add list verbatim
		if len(callArgs.Arguments) > 0 {
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
									files = append(files, f.Name())
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
	}

	for i := 0; i < taskListSet.Max(); i++ {
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

		wg.Add(1)

		err = command.RunCommand(c2, tasks, wg)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		c.SequenceIncr()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}
