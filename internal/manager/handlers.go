package manager

import (
	"encoding/json"
	"github.com/engpetarmarinov/gotama/internal/base"
	"io"
	"log/slog"
	"net/http"
)

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: get tasks
	resp := base.TaskResponse{}
	writeSuccessResponse(w, resp)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: get task
	resp := base.TaskResponse{}
	writeSuccessResponse(w, resp)
}

func postTaskHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn(err.Error())
		writeErrorResponse(w, http.StatusBadRequest, "can't read body")
		return
	}

	var reqReq base.TaskRequest
	err = json.Unmarshal(body, &reqReq)
	if err != nil {
		slog.Warn(err.Error())
		writeErrorResponse(w, http.StatusBadRequest, "can't unmarshal req")
		return
	}

	taskMsg, err := base.NewTaskMessageFromRequest(&reqReq)
	if err != nil {
		slog.Warn(err.Error())
		writeErrorResponse(w, http.StatusBadRequest, "can't get task msg")
		return
	}

	//TODO: store task msg

	resp, err := base.NewTaskResponseFromMessage(taskMsg)
	if err != nil {
		slog.Warn(err.Error())
		writeErrorResponse(w, http.StatusInternalServerError, "can't get task response")
		return
	}

	writeSuccessResponse(w, resp)
}

func putTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: update task
	resp := base.TaskResponse{}
	writeSuccessResponse(w, resp)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: delete task
	resp := base.TaskResponse{}
	writeSuccessResponse(w, resp)
}
