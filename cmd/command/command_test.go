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
	so, se, err := command.Execute()
	is.NoErr(err)

	t.Log("stdout", so, "stderr", se)
}

func TestPrepare(t *testing.T) {
	is := is.New(t)

	wd, err := os.Getwd()
	is.NoErr(err)
	t.Log(os.Getwd())

	taskList := tasks.NewTaskList()
	taskList.Add(filepath.Join(wd, "./command_test.go"))
	taskList.Add(filepath.Join(wd, "./command.go"))
	taskList2 := tasks.NewTaskList()
	taskList2.Add("a", "b", "c", "d")

	taskListSet := tasks.NewTaskListSet()
	taskListSet.Add(taskList)
	taskListSet.Add(taskList2)

	comands := []string{
		"full path: {} {2}",
		"input line no ext: {.} {2}",
		"filename: {/} {2}",
		"path {//} {2}",
		"fn no path: {/.} {2}",
		"sequence {#}",
		"slot number {%}",
	}
	// "numbered item {2}"}

	for i := 0; i < taskListSet.Max(); i++ {
		for _, v := range comands {
			c := NewCommand(v, &taskListSet)
			err := c.Prepare()
			is.NoErr(err)

			t.Log("start", v, "c command", c.Command)
		}
	}

	taskListSet.SequenceReset()
	for i := 0; i < 20; i++ {
		c := NewCommand("sequence {#}", &taskListSet)
		err := c.Prepare()
		is.NoErr(err)

		t.Log("start", "sequence {#}", "c command", c.Command)
	}

	taskListSet.SequenceReset()
	for i := 0; i < 20; i++ {
		c := NewCommand("slot number {%}", &taskListSet)
		err := c.Prepare()
		is.NoErr(err)

		t.Log("start", "slot number {%}", "c command", c.Command)
	}
}
