package tasks

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// Task a task to run
type Task struct {
	Task string
}

// NewTask make a new task
func NewTask(task string) *Task {
	t := Task{Task: task}

	return &t
}

// TaskList a list of tasks to run
type TaskList struct {
	Tasks  []Task
	Offset int
}

// NewTaskList make a new task list
func NewTaskList() TaskList {
	tl := TaskList{}
	tl.Offset = 0

	return tl
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
func (tl TaskList) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tl.Tasks), func(i, j int) { tl.Tasks[i], tl.Tasks[j] = tl.Tasks[j], tl.Tasks[i] })
}

// TaskListSet a set of task lists
type TaskListSet struct {
	TaskLists []*TaskList
	Offset    int64
}

// NewTaskListSet make a new task list set
func NewTaskListSet() TaskListSet {
	tls := TaskListSet{}
	tls.TaskLists = make([]*TaskList, 0)
	// tls.Sequence = 1

	return tls
}

// OffsetReset reset sequence
func (tls *TaskListSet) OffsetReset() {
	for _, tl := range tls.TaskLists {
		tl.Offset = 0
	}

	atomic.StoreInt64(&tls.Offset, 0)
}

// AddTaskList add a task list ot the taskSet
func (tls *TaskListSet) AddTaskList(taskList TaskList) {
	tls.TaskLists = append(tls.TaskLists, &taskList)
}

// Max get maximum task list size
func (tls *TaskListSet) Max() (max int) {
	for _, v := range tls.TaskLists {
		if len(v.Tasks) > max {
			max = len(v.Tasks)
		}
	}

	return
}

// NextAll get next item slice for all tasks item lists
func (tls TaskListSet) NextAll() (tasks []Task, err error) {
	for i := range tls.TaskLists {
		var task Task
		task, err = tls.Next(i)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}

	return
}

// Next treat task list as a circle that loops back to zero
func (tls *TaskListSet) Next(list int) (task Task, err error) {
	var taskList *TaskList
	if list <= len(tls.TaskLists)-1 {
		taskList = tls.TaskLists[list]
	} else {
		err = fmt.Errorf("list %d out of bounds for %d lists", list, len(tls.TaskLists)-1)
		return
	}
	task = taskList.Tasks[taskList.Offset]
	newOffset := taskList.Offset
	if newOffset >= len(taskList.Tasks)-1 {
		taskList.Offset = 0
	} else {
		taskList.Offset++
	}

	return
}

// Next treat task list as a circle that loops back to zero
func (tl *TaskList) Next() (task Task, atEnd bool, err error) {
	task = tl.Tasks[tl.Offset]
	newOffset := tl.Offset
	if newOffset >= len(tl.Tasks)-1 {
		tl.Offset = 0
	} else {
		tl.Offset++
	}
	if tl.Offset == len(tl.Tasks) {
		atEnd = true
	}

	return
}
