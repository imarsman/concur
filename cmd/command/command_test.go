package command

import (
	"os"
	"path/filepath"
	"runtime"
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
	threads := runtime.NumCPU()

	wd, err := os.Getwd()
	is.NoErr(err)
	t.Log(os.Getwd())

	taskList := tasks.NewTaskList()
	taskList.Add(filepath.Join(wd, "./command_test.go"))
	taskList.Add(filepath.Join(wd, "./command.go"))
	taskList2 := tasks.NewTaskList()
	taskList2.Add(filepath.Join(wd, "a", "b", "c", "d"))

	taskListSet := tasks.NewTaskListSet()
	taskListSet.Add(taskList)
	taskListSet.Add(taskList2)

	comands := []string{"full path: {}", "input line no ext: {.}", "filename: {/}", "path {//}", "fn no path: {/.}"}

	for i := 0; i < taskListSet.Max(); i++ {
		for _, v := range comands {
			c := NewCommand(v, &taskListSet, threads)
			err := c.Prepare(taskListSet.Sequence)
			is.NoErr(err)

			t.Log("start", v, "c command", c.Command)
		}
	}
}
