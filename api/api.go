package api

import (
	"encoding/json"
	"executor/domain"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var server *Server
var isWorking bool

type HandlerFunction func(w http.ResponseWriter, r *http.Request)

func handler(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	paths = slices.DeleteFunc(paths, func(s string) bool {
		return s == ""
	})

	if len(paths) == 0 {
		website(w, r)
		return
	}
	if len(paths) >= 1 {
		switch paths[0] {
		case "task":
			if len(paths) == 2 {
				_, err := strconv.ParseUint(paths[1], 10, 64)
				if err != nil {
					fmt.Printf("Unable parse to int value: %s\n", paths[1])
				}
				if r.Method == http.MethodDelete {
					// delete one task by ID
					//actualServer.Tasks.Remove(id)
				}
			}
			if r.Method == http.MethodGet {
				// return all tasks
				response := server.tasks.GetJson()
				// fmt.Printf("response: %s\n", response)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
				return
			}
			if r.Method == http.MethodPost {
				// create a new task

				var data map[string]interface{}

				if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				Command := data["cmd"].(string)
				Limit, _ := strconv.ParseUint(data["limit"].(string), 10, 16)
				Type, _ := strconv.ParseInt(data["run"].(string), 10, 16)

				task := &domain.Task{
					Id:          0,
					OutputLimit: uint16(Limit),
					Type_run:    int16(Type),
					Status:      domain.STATUS_TASK_WAITING,
					Time_start:  time.Now(),
					Time_finish: time.Now(),
					Time_human:  "00:00:00",
					Command:     Command,
					Output:      " ",
				}

				server.tasks.Add(task)
				w.Write([]byte("{}"))
			}
		case "system":
			if r.Method == http.MethodPost {
				if isWorking {
					server.scheduler.Stop()
				} else {
					server.scheduler.Start()
				}
				isWorking = !isWorking
				if isWorking {
					w.Write([]byte("{\"status\": \"working\"}"))
				} else {
					w.Write([]byte("{\"status\": \"stopped\"}"))
				}
			}
			if r.Method == http.MethodGet {
				if isWorking {
					w.Write([]byte("{\"status\": \"working\"}"))
				} else {
					w.Write([]byte("{\"status\": \"stopped\"}"))
				}
			}
		}
	}
}

func website(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("website/main.html")
	if err != nil {
		log.Fatal(">> ERROR:")
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(content))
	//fmt.Fprint(w, actualServer.Mainpage)
}

type Server struct {
	address   string
	tasks     *domain.Tasks
	scheduler *domain.Scheduler
}

func NewServer(path string) *Server {
	var tasks *domain.Tasks
	isWorking = false

	tasks = domain.NewTasks()
	server = &Server{
		address:   path,
		tasks:     tasks,
		scheduler: domain.NewScheduler(1, tasks),
	}
	return server
}

func (s *Server) Run() error {
	http.HandleFunc("/", handler)
	return http.ListenAndServe(s.address, nil)
}
