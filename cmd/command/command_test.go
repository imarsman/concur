package command

import (
	"testing"

	"github.com/imarsman/goparallel/cmd/tasks"
	"github.com/matryer/is"
)

func TestCommand(t *testing.T) {
	is := is.New(t)

	command := Command{}
	command.Value = "ls -ltr"
	so, se, err := command.Execute()
	is.NoErr(err)

	t.Log("stdout", so, "stderr", se)
}

func TestPrepare(t *testing.T) {
	cOrig := NewCommand("ls {}")

	c := cOrig.Copy()
	c2 := cOrig.Copy()

	taskList := tasks.NewTaskList()
	taskList.Add("command_test.go")
	taskList.Add("command.go")

	task1 := tasks.NewTask("command_test.go")
	task2 := tasks.NewTask("command.go")

	c.Prepare(task1)
	c2.Prepare(task2)

	t.Logf("cOrig %+v", cOrig)
	t.Logf("c %+v", c)
	t.Logf("c2 %+v", c2)
}
