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
		taskList := tasks.NewTaskList()
		for _, v := range callArgs.Files {
			matches, err := filepath.Glob(v)
			if err != nil {
				return
			}
			taskList.Add(matches...)
		}
		if callArgs.Shuffle {
			taskList.Shuffle()
		}
		taskListSet.Add(taskList)
	}

	if len(callArgs.Arguments) > 0 {
		for _, v := range callArgs.Arguments {
			taskList := tasks.NewTaskList()
			if parse.RERange.MatchString(v) {
				// fmt.Println(v)
				items, err := parse.Range(v)
				if err != nil {
					fmt.Println(err)
					return
				}
				// fmt.Println("items", items)
				taskList.Add(items...)
			} else {
				taskList.Add(v)
			}
			if callArgs.Shuffle {
				taskList.Shuffle()
			}
			// fmt.Println("adding", taskList)
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
