package manager

import (
	"github.com/engpetarmarinov/gotama/internal/config"
	mw "github.com/engpetarmarinov/gotama/internal/middleware"
	"net/http"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) RegisterRoutes(config config.API, broker Broker) http.Handler {
	// swagger:route GET /api/v1/tasks tasks listTasks
	//
	// List tasks.
	//
	// Retrieves a list of all submitted tasks with pagination.
	//
	//     Produces:
	//     - application/json
	//
	//     Parameters:
	//     - +name: limit
	//       in: query
	//       description: Maximum number of tasks to return
	//       required: false
	//       type: integer
	//       format: int32
	//     - +name: offset
	//       in: query
	//       description: Offset to start returning tasks
	//       required: false
	//       type: integer
	//       format: int32
	//
	//     Responses:
	//       200: Response
	r.mux.HandleFunc(
		"GET /api/v1/tasks",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(getTasksHandler(broker))))))

	// swagger:route GET /api/v1/tasks/{taskId} tasks getTask
	//
	// Get a task.
	//
	// Retrieves the details of an existing task by its ID.
	//
	//     Produces:
	//     - application/json
	//
	//     Parameters:
	//     - +name: taskId
	//       in: path
	//       description: ID of the task to retrieve
	//       required: true
	//       type: string
	//
	//     Responses:
	//       200: Response
	//       404: Response
	r.mux.HandleFunc(
		"GET /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(getTaskHandler(broker))))))

	// swagger:route POST /api/v1/tasks tasks addTask
	//
	// Add a new task.
	//
	// This will create a new task that can be executed immediately or periodically.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Parameters:
	//     - name: task
	//       in: body
	//       description: Task object
	//       required: true
	//       schema:
	//         "$ref": "#/definitions/taskRequest"
	//
	//     Responses:
	//       201: Response
	r.mux.HandleFunc(
		"POST /api/v1/tasks",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(postTaskHandler(config, broker))))))

	// swagger:route PUT /api/v1/tasks/{taskId} tasks updateTask
	//
	// Update a task.
	//
	// Updates the details of an existing task by its ID.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Parameters:
	//     - name: taskId
	//       in: path
	//       description: ID of the task to update
	//       required: true
	//       type: string
	//     - name: task
	//       in: body
	//       description: Updated task object
	//       required: true
	//       schema:
	//         "$ref": "#/definitions/taskRequest"
	//
	//     Responses:
	//       200: Response
	//       404: Response
	r.mux.HandleFunc(
		"PUT /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(putTaskHandler(config, broker))))))

	// swagger:route DELETE /api/v1/tasks/{taskId} tasks deleteTask
	//
	// Delete a task.
	//
	// Deletes an existing task by its ID.
	//
	//     Produces:
	//     - application/json
	//
	//     Parameters:
	//     - +name: taskId
	//       in: path
	//       description: ID of the task to delete
	//       required: true
	//       type: string
	//
	//     Responses:
	//       200: Response
	//       404: Response
	r.mux.HandleFunc(
		"DELETE /api/v1/tasks/{id}",
		mw.WithLogging(mw.WithCommonHeaders(mw.WithAuth(mw.WithRBAC(deleteTaskHandler(broker))))))

	return r.mux
}
