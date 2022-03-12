package tasks

import (
	"fmt"
	"math/rand"
	"time"
)

// Task a task to run
type Task struct {
	Task string
}

func NewTask(task string) *Task {
	t := Task{Task: task}

	return &t
}

// TaskList a list of tasks to run
type TaskList struct {
	Tasks  []Task
	Offset int
}

func NewTaskList() *TaskList {
	tl := TaskList{}
	return &tl
}

// Add add tasks to a task list
func (tl *TaskList) Add(tasks ...string) {
	for _, v := range tasks {
		task := Task{}
		task.Task = v
		tl.Tasks = append(tl.Tasks, task)
	}
}

// Shuffle shuffle the task lines for a task list
func (tl *TaskList) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tl.Tasks), func(i, j int) { tl.Tasks[i], tl.Tasks[j] = tl.Tasks[j], tl.Tasks[i] })
}

// TaskListSet a set of task lists
type TaskListSet struct {
	TaskLists []*TaskList
}

func NewTaskListSet() *TaskListSet {
	tls := TaskListSet{}
	tls.TaskLists = make([]*TaskList, 0)

	return &tls
}

// Add add a task list ot the taskSet
func (ts *TaskListSet) Add(taskList *TaskList) {
	ts.TaskLists = append(ts.TaskLists, taskList)
}

// Next treat task list as a circle that loops back to zero
func (ts *TaskListSet) Next(list int) (task Task, err error) {
	var taskList *TaskList
	if list <= len(ts.TaskLists)-1 {
		taskList = ts.TaskLists[list]
	} else {
		err = fmt.Errorf("list %d out of bounds for %d lists", list, len(ts.TaskLists)-1)
		return
	}
	// if taskList.Offset <= len(taskList.Tasks)-1 {
	task = taskList.Tasks[taskList.Offset]
	newOffset := taskList.Offset + 1
	if newOffset > len(taskList.Tasks)-1 {
		taskList.Offset = 0
	} else {
		taskList.Offset = newOffset
	}

	return
}
