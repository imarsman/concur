package tasks

import (
	"testing"

	"github.com/matryer/is"
)

func TestTask(t *testing.T) {
	is := is.New(t)
	task := NewTask("tasks.go")
	t.Logf("%+v", task)
	is.True(1 == 1)
}

func TestTaskSet(t *testing.T) {
	is := is.New(t)
	taskList := NewTaskList()
	taskList.Add("tasks.go", "tasks_test.go")
	taskListSet := NewTaskListSet()
	taskListSet.Add(taskList)

	t.Logf("Task list set %+v", taskListSet)

	for i := 0; i < 10; i++ {
		task, atEnd, err := taskListSet.next(0)
		is.NoErr(err)
		is.True(atEnd == false)

		t.Logf("task %+v", task)
	}
	is.True(1 == 1)
}

func TestTaskSetSequence(t *testing.T) {
	is := is.New(t)
	taskList := NewTaskList()
	taskList.Add("tasks.go", "tasks_test.go")
	taskListSet := NewTaskListSet()
	taskListSet.Add(taskList)

	t.Logf("Task list set %+v", taskListSet)

	for i := 0; i < 10; i++ {
		task, atEnd, err := taskListSet.next(0)
		is.NoErr(err)
		is.True(atEnd == false)

		t.Logf("task %+v Sequence %d", task, taskListSet.Sequence)
		taskListSet.SequenceIncr()

	}
	is.True(1 == 1)
}
