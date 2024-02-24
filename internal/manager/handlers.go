package manager

import (
	"context"
	"encoding/json"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/task"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func getTasksHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		limitStr := params.Get("limit")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 100
		}
		offsetStr := params.Get("offset")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		totalTaskMsgs, taskMsgs, err := broker.GetAllTasks(context.Background(), offset, limit)
		if err != nil {
			slog.Error(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "error getting all tasks")
			return
		}

		var tasks []*task.Response
		for _, taskMsg := range taskMsgs {
			taskResp, err := task.NewResponseFromMessage(taskMsg)
			if err != nil {
				slog.Error(err.Error())
				writeErrorResponse(w, http.StatusInternalServerError, "error getting task response")
				return
			}

			tasks = append(tasks, taskResp)
		}

		resp := struct {
			Total int64            `json:"total"`
			Tasks []*task.Response `json:"tasks"`
		}{
			Total: totalTaskMsgs,
			Tasks: tasks,
		}
		writeSuccessResponse(w, resp)
	}
}

func getTaskHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := strings.ToLower(strings.TrimSpace(r.PathValue("id")))
		if taskID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "no task id provided")
			return
		}
		taskMsg, err := broker.GetTask(context.Background(), taskID)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "error getting msg")
			return
		}

		resp, err := task.NewResponseFromMessage(taskMsg)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "error getting task response")
			return
		}

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

		var taskReq task.Request
		err = json.Unmarshal(body, &taskReq)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusBadRequest, "error unmarshalling req")
			return
		}

		taskMsg, err := task.NewMessageFromRequest(&taskReq)
		if err != nil {
			slog.Warn(err.Error())
			writeErrorResponse(w, http.StatusBadRequest, "error getting task msg")
			return
		}

		err = broker.EnqueueTask(context.Background(), taskMsg)
		if err != nil {
			slog.Error(err.Error())
			writeErrorResponse(w, http.StatusInternalServerError, "error enqueueing task")
			return
		}

		resp, err := task.NewResponseFromMessage(taskMsg)
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
		resp := task.Response{}
		writeSuccessResponse(w, resp)
	}
}

func deleteTaskHandler(broker broker.Broker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO: delete task
		resp := task.Response{}
		writeSuccessResponse(w, resp)
	}
}
