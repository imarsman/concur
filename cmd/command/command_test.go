package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/imarsman/goparallel/cmd/tasks"
	"github.com/matryer/is"
)

func TestCommand(t *testing.T) {
	is := is.New(t)

	command := Command{}
	command.Command = "ls -ltr"
	err := command.Execute()
	is.NoErr(err)
}

func TestPrepare(t *testing.T) {
	is := is.New(t)

	wd, err := os.Getwd()
	is.NoErr(err)
	t.Log(os.Getwd())

	// find . -type f -name "*.txt" | goparallel echo {}
	taskList := tasks.NewTaskList()
	taskList.Add(filepath.Join(wd, "./command_test.go"))
	taskList.Add(filepath.Join(wd, "./command.go"))
	taskList2 := tasks.NewTaskList()
	taskList2.Add("a", "b", "c", "d") // {0..50}

	taskListSet := tasks.NewTaskListSet()
	taskListSet.AddTaskList(taskList)
	taskListSet.AddTaskList(taskList2)

	comands := []string{
		"echo 'full path: {} {2}'",
		"echo 'input line no ext: {.} {2}'",
		"echo 'filename: {/} {2}'",
		"echo 'path {//} {2}'",
		"echo 'fn no path: {/.} {2}'",
		"echo 'sequence {#}'",
		"echo 'slot number {%}'",
	}

	for i := 0; i < taskListSet.Max(); i++ {
		for _, v := range comands {
			c := NewCommand(v, &taskListSet, Config{})
			atEnd, err := c.Prepare()
			if err != nil {

			}
			is.True(atEnd == false)

			t.Log("start", v, "c command", c.Command)
			err = c.Execute()
			is.NoErr(err)
		}
	}

	taskListSet.SequenceReset()
	for i := 0; i < 20; i++ {
		c := NewCommand("sequence {#}", &taskListSet, Config{})
		atEnd, err := c.Prepare()
		is.NoErr(err)
		is.True(atEnd == false)

		t.Log("start", "sequence {#}", "c command", c.Command)
	}

	taskListSet.SequenceReset()
	for i := 0; i < 20; i++ {
		c := NewCommand("slot number {%}", &taskListSet, Config{})
		atEnd, err := c.Prepare()
		is.NoErr(err)
		is.True(atEnd == false)

		t.Log("start", "slot number {%}", "c command", c.Command, false)
	}
}
