package manager

import (
	"github.com/engpetarmarinov/gotama/internal/base"
	"net/http"
)

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: get tasks
	resp := []base.Task{
		{Status: base.TaskStatusPending},
	}
	writeSuccessResponse(w, resp)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: get task
	resp := base.Task{
		Status: base.TaskStatusPending,
	}
	writeSuccessResponse(w, resp)
}

func postTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: add task
	resp := base.Task{
		Status: base.TaskStatusPending,
	}
	writeSuccessResponse(w, resp)
}

func putTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: update task
	resp := base.Task{
		Status: base.TaskStatusPending,
	}
	writeSuccessResponse(w, resp)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: delete task
	resp := base.Task{
		Status: base.TaskStatusPending,
	}
	writeSuccessResponse(w, resp)
}
