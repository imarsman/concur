package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/imarsman/goparallel/cmd/command"
	"github.com/imarsman/goparallel/cmd/parse"
	"github.com/imarsman/goparallel/cmd/tasks"
)

var slots int

func init() {
	slots = 8
}

// args CLI args

// callArgs command line arguments
var callArgs struct {
	Command   []string `arg:"positional"`
	Arguments []string `arg:"-a,--arguments,separate"`
	Files     []string `arg:"-f,--files"`
	DryRun    bool     `arg:"-d,--dry-run"`
	Slots     int      `arg:"-s,--slots"`
	Shuffle   bool     `arg:"-S,--shuffle"`
}

func main() {
	arg.MustParse(&callArgs)

	dryRun := false

	if callArgs.DryRun {
		dryRun = true
	}

	if callArgs.Slots == 0 {
		callArgs.Slots = runtime.NumCPU()
	}

	// Use stdin if it is available
	stat, _ := os.Stdin.Stat()

	taskListSet := tasks.NewTaskListSet()

	stdinItems := []string{}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		var scanner = bufio.NewScanner(os.Stdin)
		// Tell scanner to scan by lines.
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			item := scanner.Text()
			item = strings.TrimSpace(item)
			stdinItems = append(stdinItems, item)
		}
		if len(stdinItems) > 0 {
			taskList := tasks.NewTaskList()
			taskList.Add(stdinItems...)
			taskListSet.Add(taskList)
		}
	}

	if len(callArgs.Files) > 0 {
		for _, v := range callArgs.Files {
			taskList := tasks.NewTaskList()
			parts := strings.Split(v, " ")
			for _, part := range parts {
				part = strings.TrimSpace(part)

				matches, err := filepath.Glob(part)
				if err != nil {
					return
				}
				if callArgs.Shuffle {
					taskList.Shuffle()
				}
				taskList.Add(matches...)
			}
			taskListSet.Add(taskList)
		}
	}

	if len(callArgs.Arguments) > 0 {
		for _, v := range callArgs.Arguments {
			parts := strings.Split(v, " ")
			taskList := tasks.NewTaskList()
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
					taskList.Add(part)
				}
				if callArgs.Shuffle {
					taskList.Shuffle()
				}
			}
			taskListSet.Add(taskList)
		}
	}

	// Define command to run
	var c = command.NewCommand(
		strings.Join(callArgs.Command, " "),
		&taskListSet,
		callArgs.Slots,
		dryRun,
	)
	// Run command for all items
	command.RunCommand(c)
}
