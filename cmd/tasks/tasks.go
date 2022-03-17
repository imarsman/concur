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
	Sequence  int64
	Offset    int64
}

// NewTaskListSet make a new task list set
func NewTaskListSet() TaskListSet {
	tls := TaskListSet{}
	tls.TaskLists = make([]*TaskList, 0)
	tls.Sequence = 1

	return tls
}

// SequenceSet set sequence
// For use with things like {2} where getting extra items is needed.
func (tls *TaskListSet) SequenceSet(sequence int64) (err error) {
	if sequence > int64(len(tls.TaskLists)-1) {
		err = fmt.Errorf("sequence to set %d outside of max %d", len(tls.TaskLists)-1, sequence)
	}
	atomic.AddInt64(&tls.Sequence, sequence)

	return
}

// SequenceIncr increment sequence without lock
func (tls *TaskListSet) SequenceIncr() {
	atomic.AddInt64(&tls.Sequence, 1)
}

// SequenceReset reset sequence
func (tls *TaskListSet) SequenceReset() {
	atomic.StoreInt64(&tls.Sequence, 0)
}

// OffsetReset reset sequence
func (tls *TaskListSet) OffsetReset() {
	fmt.Println("offset reset")
	for _, tl := range tls.TaskLists {
		tl.Offset = 0
		fmt.Println("offset", tl.Offset)
	}
	atomic.StoreInt64(&tls.Offset, 0)
}

// GetSequence get lock free sequence value
func (tls *TaskListSet) GetSequence() int64 {
	return atomic.LoadInt64(&tls.Sequence)
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
func (tls TaskListSet) NextAll() (tasks []Task, atEnd bool, err error) {
	for i := range tls.TaskLists {
		var task Task
		task, atEnd, err = tls.Next(i)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}

	return
}

// Next treat task list as a circle that loops back to zero
func (tls *TaskListSet) Next(list int) (task Task, atEnd bool, err error) {
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
	if taskList.Offset == tls.Max() {
		atEnd = true
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
