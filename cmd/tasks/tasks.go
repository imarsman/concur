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
	StdInList *TaskList
	TaskLists []*TaskList
	ArgsLists []*TaskList
	Sequence  int64
	Offset    int64
}

// NewTaskListSet make a new task list set
func NewTaskListSet() TaskListSet {
	tls := TaskListSet{}
	stdInList := NewTaskList()
	tls.StdInList = &stdInList
	tls.TaskLists = make([]*TaskList, 0)
	tls.ArgsLists = make([]*TaskList, 0)
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

// GetSequence get lock free sequence value
func (tls *TaskListSet) GetSequence() int64 {
	return atomic.LoadInt64(&tls.Sequence)
}

// AddTaskList add a task list ot the taskSet
func (tls *TaskListSet) AddTaskList(taskList TaskList) {
	tls.TaskLists = append(tls.TaskLists, &taskList)
}

// AddArgList add a task list ot the taskSet
func (tls *TaskListSet) AddArgList(taskList TaskList) {
	tls.ArgsLists = append(tls.ArgsLists, &taskList)
}

// SetStdInList set the stdin list
func (tls *TaskListSet) SetStdInList(taskList *TaskList) {
	tls.StdInList = taskList
}

// Max get maximum task list size
func (tls *TaskListSet) Max() (max int) {
	for _, v := range tls.TaskLists {
		if len(v.Tasks) > max {
			max = len(v.Tasks)
		}
	}
	if len(tls.StdInList.Tasks) > max {
		max = len(tls.StdInList.Tasks)
	}
	for _, v := range tls.ArgsLists {
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
		task, atEnd, err = tls.next(i)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}
	for i := range tls.ArgsLists {
		var task Task
		task, atEnd, err = tls.next(i)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}
	if len(tls.StdInList.Tasks) > 0 {
		var tl = tls.StdInList
		var task Task
		task, atEnd, err = tl.next()
		if err != nil {
			return
		}
		tasks = append(tasks, task)
	}

	return
}

/*
Fails
$ goparallel 'echo {1} {2} {3}' -a '1 2 3' -o
1 2 3
2 3 1
3 1 2
*/

func (tls TaskListSet) BackUpAll() (err error) {
	for i := range tls.TaskLists {
		var taskList = tls.TaskLists[i]
		offset := taskList.Offset
		fmt.Println("offset", offset)
		if offset == len(taskList.Tasks)-1 {
			fmt.Println("here", taskList.Offset)
			taskList.Offset = 0
			fmt.Println("here", taskList.Offset)
		} else {
			if offset >= len(taskList.Tasks)-1 {
				fmt.Println("there", taskList.Offset)
				offset = len(taskList.Tasks) - 1
				fmt.Println("there", taskList.Offset)
			}
		}
		// fmt.Println("here", i)
	}
	return
}

// next treat task list as a circle that loops back to zero
func (tls *TaskListSet) next(list int) (task Task, atEnd bool, err error) {
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

// next treat task list as a circle that loops back to zero
func (tl *TaskList) next() (task Task, atEnd bool, err error) {
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
