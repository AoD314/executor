package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

const STATUS_TASK_WAITING = 1
const STATUS_TASK_RUNNING = 2
const STATUS_TASK_CANSELED = 3
const STATUS_TASK_FAILED = 4
const STATUS_TASK_DONE = 5

type Task struct {
	Id          uint32    `json:"id"`
	OutputLimit uint16    `json:"outlim"`
	Type_run    int16     `json:"type"`
	Status      int       `json:"status"`
	Time_start  time.Time `json:"tstart"`
	Time_finish time.Time `json:"tfinish"`
	Time_human  string    `json:"time"`
	Command     string    `json:"cmd"`
	Output      string    `json:"output"`
}

type Tasks struct {
	repo *SQLiteRepository
}

func NewTasks() *Tasks {
	return &Tasks{
		repo: NewSQLiteRepository(),
	}
}

func (t *Tasks) amount_running_task() int {

	q := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE status = %d;", STATUS_TASK_RUNNING)
	fmt.Println(q)

	rows, err := t.repo.db.Query(q)
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	if rows == nil {
		return 0
	}

	defer rows.Close()

	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			fmt.Errorf("Error: %s", err)
		}
		break
	}

	return count
}

func (t *Tasks) get_task_by_status(limit int, status int) []Task {
	var tasks []Task

	q := fmt.Sprintf("SELECT * FROM tasks WHERE status = %d LIMIT %d;", status, limit)
	fmt.Println(q)

	rows, err := t.repo.db.Query(q)
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}
	if rows == nil {
		return tasks
	}
	defer rows.Close()

	for rows.Next() {
		var task Task

		var tstart string
		var tfinish string

		if err := rows.Scan(
			&task.Id,
			&task.OutputLimit,
			&task.Type_run,
			&task.Status,
			&task.Output,
			&task.Command,
			&tstart,
			&tfinish); err != nil {
			fmt.Errorf("Error: %s", err)
		}

		task.Time_start, _ = time.Parse(time.RFC3339, tstart)
		task.Time_finish, _ = time.Parse(time.RFC3339, tfinish)

		diff := task.Time_finish.Sub(task.Time_start)

		hours := int(diff.Hours())
		minutes := int(diff.Minutes()) % 60
		seconds := diff.Seconds() / 60

		task.Time_human = fmt.Sprintf("%02d:%02d:%2.3f", hours, minutes, seconds)

		tasks = append(tasks, task)
	}
	return tasks
}

func (t *Tasks) Update(task *Task) {
	q := fmt.Sprintf("UPDATE tasks SET lim = '%d', type = '%d', status = '%d', output = '%s', cmd = '%s', time_start = '%s', time_finish = '%s' WHERE id = %d;", task.OutputLimit, task.Type_run, task.Status, task.Output, task.Command, task.Time_start.Format(time.RFC3339), task.Time_finish.Format(time.RFC3339), task.Id)
	fmt.Println(q)

	_, err := t.repo.db.Exec(q)
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}
}

func (t *Tasks) Delete(id int64) {
	q := fmt.Sprintf("DELETE FROM tasks WHERE id = %d;", id)
	fmt.Println(q)

	_, err := t.repo.db.Exec(q)
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}
}

func (t *Tasks) Add(task *Task) {
	q := fmt.Sprintf("INSERT INTO tasks (lim, type, status, output, cmd, time_start, time_finish) VALUES (%d, %d, %d, '%s', '%s', '%s', '%s');",
		task.OutputLimit, task.Type_run, task.Status, task.Output, task.Command, task.Time_start.Format(time.RFC3339), task.Time_finish.Format(time.RFC3339))
	fmt.Println(q)

	_, err := t.repo.db.Exec(q)
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}
}

func (t *Tasks) GetJson() string {

	var tasks []Task

	tasks = append(tasks, t.get_task_by_status(1000, STATUS_TASK_RUNNING)...)
	tasks = append(tasks, t.get_task_by_status(1000, STATUS_TASK_WAITING)...)
	tasks = append(tasks, t.get_task_by_status(1000, STATUS_TASK_FAILED)...)
	tasks = append(tasks, t.get_task_by_status(1000, STATUS_TASK_DONE)...)

	var data []byte

	if len(tasks) > 0 {
		var err error
		data, err = json.Marshal(tasks)
		if err != nil {
			fmt.Errorf("Error: %s", err)
		}
	} else {
		data = []byte("{}")
	}

	return string(data)
}

/*

func (t *Tasks) GetCommand(index uint64) string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, task := range t.tasks {
		if task.Id == index {
			return task.Command
		}
	}

	return ""
}

func (t *Tasks) GetIDs() []uint64 {
	var ids []uint64
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, task := range t.tasks {
		ids = append(ids, task.Id)
	}

	return ids
}

func (t *Tasks) GetStatus(index uint64) string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, task := range t.tasks {
		if task.Id == index {
			return task.Status
		}
	}

	return ""
}

func (t *Tasks) SetStatus(index uint64, status string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for i, _ := range t.tasks {
		task := &t.tasks[i]
		if task.Id == index {
			task.Status = status
			if status == "run" {
				task.Time_start = time.Now()
			}
			if status == "done" {
				task.Time_finish = time.Now()
			}

			secs := time.Since(task.Time_start).Seconds()
			h := int(secs / 3600)
			secs -= float64(h * 3600)
			m := int(secs / 60)
			secs -= float64(m * 60)
			task.Time_human = fmt.Sprintf("%02dh %02dm %02.3f", h, m, secs)

			break
		}
	}
}

func (t *Tasks) SetOutput(index uint64, output string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for i, _ := range t.tasks {
		task := &t.tasks[i]
		if task.Id == index {
			task.Output = output
			break
		}
	}
}

func (t *Tasks) Add(Command string, TypeRun int8) uint64 {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	globalID++
	idObject := globalID

	task := Task{
		Id:          idObject,
		Time_start:  time.Now(),
		Time_finish: time.Now(),
		Command:     Command,
		Status:      "wait",
		Type_run:    TypeRun,
	}

	t.tasks = append(t.tasks, task)

	return idObject
}

func (t *Tasks) GetJson() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	data, err := json.Marshal(t.tasks)
	if err != nil {
		fmt.Errorf("Error: %s", err)
	}

	return string(data)
}

func (t *Tasks) Remove(ID uint64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var index int
	index = -1

	for i, task := range t.tasks {
		if task.Id == ID {
			index = i
			break
		}
	}

	if index >= 0 {
		t.tasks = append(t.tasks[:index], t.tasks[index+1:]...)
		return nil
	} else {

		return fmt.Errorf("Can't found task")
	}
}

func (t *Tasks) Len() int {
	return len(t.tasks)
}
*/
