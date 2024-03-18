package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Task represents a meaninglessness of our life
// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Task represents a meaninglessness of our life
type Task struct {
	id            int
	createdTime   time.Time // время создания
	completedTime time.Time // время выполнения
	taskResult    string

	err error
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	taskCh := make(chan Task, 10)

	go CreateTasks(ctx, taskCh)

	taskManager := NewTaskManager(taskCh, cancel, 3*time.Second)

	taskManager.HandleTasks()
	taskManager.Print()
}

type TaskManager struct {
	handleTime time.Duration
	taskCh     <-chan Task

	result     map[int]Task
	errors     []error
	cancelFunc context.CancelFunc
}

func NewTaskManager(taskCh <-chan Task, cancelFunc context.CancelFunc, handleTime time.Duration) *TaskManager {
	return &TaskManager{
		taskCh:     taskCh,
		result:     make(map[int]Task),
		errors:     []error{},
		handleTime: handleTime,
		cancelFunc: cancelFunc,
	}
}

func (m *TaskManager) HandleTasks() {
	ticker := time.NewTicker(m.handleTime)

loop:
	for {
		select {
		case task := <-m.taskCh:
			// получение тасков
			doneTask, err := taskSorter(taskWorker(task))
			if err != nil {
				m.errors = append(m.errors, err)
				break
			}
			m.result[doneTask.id] = doneTask
			break
		case <-ticker.C:
			m.cancelFunc()
			break loop
		}
	}
}

// taskSorter над названием можно подумать)
func taskSorter(task Task) (Task, error) {
	// тут можно заменить на task.err != nil
	if strings.HasSuffix(task.taskResult, "successed") {
		return task, nil
	}

	return Task{}, fmt.Errorf("task id: %d time: %s, error: %w", task.id, task.createdTime.Format(time.RFC3339), task.err)
}

// taskWorker над названием можно подумать)
func taskWorker(task Task) Task {
	if task.err == nil {
		task.taskResult = "task has been successed"
	} else {
		task.taskResult = "something went wrong"
	}

	task.completedTime = time.Now()
	time.Sleep(time.Millisecond * 150)
	return task
}

func (m *TaskManager) Print() {
	println("Errors:")
	for _, err := range m.errors {
		println(err.Error())
	}
	println("Done tasks:")
	for task := range m.result {
		println(task)
	}
}

func CreateTasks(ctx context.Context, tCh chan<- Task) {
	for {
		select {
		case <-ctx.Done():
			close(tCh)
			return
		default:
			task := Task{
				createdTime: time.Now(),
				id:          int(time.Now().Unix()),
			}

			if time.Now().UnixMilli()%2 > 0 { // вот такое условие появления ошибочных тасков
				task.err = fmt.Errorf("Some error occured")
			}

			tCh <- task // передаем таск на выполнение
		}
	}
}
