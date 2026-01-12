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

func removeEmptyAndWhitespaceStrings(arr []string) []string {
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if strings.TrimSpace(v) != "" {
			result = append(result, strings.TrimSpace(v))
		}
	}
	return result
}

func mergeStringParams(arr []string) []string {
	result := make([]string, 0, len(arr))

	isMerging := false
	arg := make([]string, 0, len(arr))
	for _, v := range arr {
		l := len(v)
		if l > 1 {
			if v[0] == 34 && v[l-1] != 34 {
				arg = append(arg, v)
				isMerging = true
				continue
			}
			if v[l-1] == 34 {
				arg = append(arg, v)
				isMerging = false
			}
		}

		if isMerging {
			arg = append(arg, v)
		} else {
			if len(arg) == 0 {
				arg = append(arg, v)
			}
			val := strings.Join(arg, " ")
			if val[0] == 34 {
				val = val[1 : len(val)-1]
			}
			result = append(result, val)
			arg = arg[:0]
		}
	}
	return result
}

func run_command(t *Task, tasks *Tasks) {
	t.Status = STATUS_TASK_RUNNING
	t.Time_start = time.Now()
	tasks.Update(t)

	cmds := removeEmptyAndWhitespaceStrings(strings.Split(t.Command, "&&"))

	for _, one_cmd := range cmds {

		one_cmd = strings.TrimSpace(one_cmd)
		arr := strings.Split(one_cmd, " ")

		name := arr[0]
		args := arr[1:]

		args = mergeStringParams(args)

		cmd := exec.Command(name, args...)
		out, err := cmd.CombinedOutput()

		if err == nil {
			l := len(out)
			if l <= int(t.OutputLimit) {
				l = 0
			} else {
				l -= int(t.OutputLimit)
			}
			t.Output = strings.ReplaceAll(string(out[l:]), "'", "\"")
			t.Status = STATUS_TASK_DONE
		} else {
			l := len(err.Error())
			if l <= 2*int(t.OutputLimit) {
				l = 0
			} else {
				l -= 2 * int(t.OutputLimit)
			}
			t.Output = "ERROR: " + strings.ReplaceAll(string(err.Error()[l:]), "'", "\"")
			t.Status = STATUS_TASK_FAILED
			break
		}
	}

	t.Time_finish = time.Now()
	tasks.Update(t)
}
