package manager

import (
	"encoding/json"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"io"
	"log/slog"
	"net/http"
)

func getTasksHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: get tasks
		resp := base.TaskResponse{}
		writeSuccessResponse(w, resp)
	}
}

func getTaskHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: get task
		resp := base.TaskResponse{}
		writeSuccessResponse(w, resp)
	}
}

func postTaskHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusBadRequest, "error reading body")
			return
		}

		var reqReq base.TaskRequest
		err = json.Unmarshal(body, &reqReq)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusBadRequest, "error unmarshalling req")
			return
		}

		taskMsg, err := base.NewTaskMessageFromRequest(&reqReq)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusBadRequest, "error getting task msg")
			return
		}

		//TODO: store task msg
		err = broker.Ping()
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "error pinging broker")
			return
		}

		resp, err := base.NewTaskResponseFromMessage(taskMsg)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "error getting task response")
			return
		}

		writeSuccessResponse(w, resp)
	}
}

func putTaskHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: update task
		resp := base.TaskResponse{}
		writeSuccessResponse(w, resp)
	}
}

func deleteTaskHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: delete task
		resp := base.TaskResponse{}
		writeSuccessResponse(w, resp)
	}
}
