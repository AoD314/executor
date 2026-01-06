package domain

import (
	"os/exec"
	"strings"
	"time"
)

type Scheduler struct {
	tasks    *Tasks
	is_run   bool
	max_jobs int
}

func (s *Scheduler) Start() {
	s.is_run = true
}

func (s *Scheduler) Stop() {
	s.is_run = false
}

func (s *Scheduler) SetMaxJobs(m int) {
	if m > 0 && m <= 128 {
		s.max_jobs = m
	}
}

func NewScheduler(max_jobs int, t *Tasks) *Scheduler {

	var s *Scheduler

	s = &Scheduler{
		tasks:    t,
		is_run:   false,
		max_jobs: max_jobs,
	}
	go schedule(s)

	return s
}

func schedule(s *Scheduler) {
	for {

		time.Sleep(1 * time.Second)

		if s.is_run == false {
			continue
		}

		running_task := s.tasks.amount_running_task()

		if running_task >= s.max_jobs {
			continue
		}

		if running_task == 0 {
			tasks := s.tasks.get_task_by_status(1, STATUS_TASK_WAITING)
			if len(tasks) == 1 {
				go run_command(&tasks[0], s.tasks)
			}
		}
	}
}

func run_command(t *Task, tasks *Tasks) {
	t.Status = STATUS_TASK_RUNNING
	t.Time_start = time.Now()
	tasks.Update(t)

	arr := strings.Split(t.Command, " ")
	name := arr[0]
	args := arr[1:]

	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()

	if err == nil {
		l := len(out)
		if l <= int(t.OutputLimit) {
			l = 0
		} else {
			l -= int(t.OutputLimit)
		}
		t.Output = string(out[l:])
		t.Status = STATUS_TASK_DONE
	} else {
		l := len(err.Error())
		if l <= 2*int(t.OutputLimit) {
			l = 0
		} else {
			l -= 2 * int(t.OutputLimit)
		}
		t.Output = "ERROR: " + string(err.Error()[l:])
		t.Status = STATUS_TASK_FAILED
	}
	t.Time_finish = time.Now()
	tasks.Update(t)
}
