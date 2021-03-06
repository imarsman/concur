package command

import (
	"testing"

	"github.com/matryer/is"
)

func TestCommand(t *testing.T) {
	is := is.New(t)

	command := Command{}
	command.Command = "ls -ltr"
	err := command.Execute()
	is.NoErr(err)
}

// func TestPrepare(t *testing.T) {
// 	is := is.New(t)

// 	wd, err := os.Getwd()
// 	is.NoErr(err)
// 	t.Log(os.Getwd())

// 	// find . -type f -name "*.txt" | goparallel echo {}
// 	taskList := tasks.NewTaskList()
// 	taskList.Add(filepath.Join(wd, "./command_test.go"))
// 	taskList.Add(filepath.Join(wd, "./command.go"))
// 	taskList2 := tasks.NewTaskList()
// 	taskList2.Add("a", "b", "c", "d") // {0..50}

// 	taskListSet := tasks.NewTaskListSet()
// 	taskListSet.AddTaskList(taskList)
// 	taskListSet.AddTaskList(taskList2)

// 	comands := []string{
// 		"echo 'full path: {} {2}'",
// 		"echo 'input line no ext: {.} {2}'",
// 		"echo 'filename: {/} {2}'",
// 		"echo 'path {//} {2}'",
// 		"echo 'fn no path: {/.} {2}'",
// 		"echo 'sequence {#}'",
// 		"echo 'slot number {%}'",
// 	}

// 	for i := 0; i < taskListSet.Max(); i++ {
// 		for _, v := range comands {
// 			c := NewCommand(v, &taskListSet, Config{})
// 			tasks, err := c.TaskListSet.NextAll()
// 			is.NoErr(err)
// 			err = c.Prepare(tasks)
// 			if err != nil {

// 			}
// 			// is.True(atEnd == false)

// 			t.Log("start", v, "c command", c.Command)
// 			err = c.Execute()
// 			is.NoErr(err)
// 		}
// 	}

// 	c := NewCommand("sequence {#}", &taskListSet, Config{})
// 	c.SequenceReset()
// 	for i := 0; i < 20; i++ {
// 		c2 := c.Copy()
// 		tasks, err := c.TaskListSet.NextAll()
// 		is.NoErr(err)
// 		err = c.Prepare(tasks)
// 		is.NoErr(err)
// 		// is.True(atEnd == false)

// 		t.Log("start", "sequence {#}", "c command", c2.Command)
// 	}

// 	c.SequenceReset()
// 	for i := 0; i < 20; i++ {
// 		c := NewCommand("slot number {%}", &taskListSet, Config{})
// 		tasks, err := c.TaskListSet.NextAll()
// 		is.NoErr(err)
// 		err = c.Prepare(tasks)
// 		is.NoErr(err)
// 		// is.True(atEnd == false)

// 		t.Log("start", "slot number {%}", "c command", c.Command, false)
// 	}
// }
