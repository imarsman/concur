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

func NewTask(task string) *Task {
	t := Task{Task: task}

	return &t
}

// TaskList a list of tasks to run
type TaskList struct {
	Tasks  []Task
	Offset int
}

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

// SequenceIncr increment sequence without lock
func (tls *TaskListSet) SequenceIncr() {
	atomic.AddInt64(&tls.Sequence, 1)
}

// GetSequence get lock free sequence value
func (tls *TaskListSet) GetSequence() int64 {
	return atomic.LoadInt64(&tls.Sequence)
}

// Add add a task list ot the taskSet
func (tls *TaskListSet) Add(taskList TaskList) {
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

func (tls TaskListSet) NextAll() (tasks []Task, err error) {
	for i := range tls.TaskLists {
		var task Task
		task, err = tls.next(i)
		if err != nil {
			return
		}
		tasks = append(tasks, task)
		// fmt.Printf("tasks %+v\n", tasks)
	}

	return
}

// next treat task list as a circle that loops back to zero
func (tls *TaskListSet) next(list int) (task Task, err error) {
	var taskList *TaskList
	if list <= len(tls.TaskLists)-1 {
		taskList = tls.TaskLists[list]
	} else {
		err = fmt.Errorf("list %d out of bounds for %d lists", list, len(tls.TaskLists)-1)
		return
	}
	task = taskList.Tasks[taskList.Offset]
	newOffset := taskList.Offset
	// fmt.Println("new offset", newOffset, len(taskList.Tasks))
	if newOffset >= len(taskList.Tasks)-1 {
		taskList.Offset = 0
	} else {
		taskList.Offset++
	}
	// fmt.Println("task", task, "offset", taskList.Offset)

	return
}
