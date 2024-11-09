package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/logger"
	"github.com/engpetarmarinov/gotama/internal/task"
	"github.com/engpetarmarinov/gotama/internal/timeutil"
	"github.com/redis/go-redis/v9"
)

type RDB struct {
	client redis.UniversalClient
	clock  timeutil.Clock
}

func NewRDB(client redis.UniversalClient, clock timeutil.Clock) *RDB {
	return &RDB{
		client: client,
		clock:  clock,
	}
}

func (r *RDB) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RDB) Close() error {
	return r.client.Close()
}

func (r *RDB) runScript(ctx context.Context, script *redis.Script, keys []string, args ...any) error {
	if err := script.Run(ctx, r.client, keys, args...).Err(); err != nil {
		return fmt.Errorf("redis eval error: %v", err)
	}
	return nil
}

func (r *RDB) runScriptWithErrorCode(ctx context.Context, script *redis.Script, keys []string, args ...any) (int64, error) {
	res, err := script.Run(ctx, r.client, keys, args...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis eval error: %v", err)
	}
	n, ok := res.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected return value from Lua script: %v", res)
	}
	return n, nil
}

// queueKeyPrefix returns a prefix for all keys in the given queue.
func queueKeyPrefix(qname string) string {
	return fmt.Sprintf("gotama:%s:", qname)
}

// taskKeyPrefix returns a prefix for task key.
func taskKeyPrefix(qname string) string {
	return fmt.Sprintf("%st:", queueKeyPrefix(qname))
}

// taskKey returns a redis key for the given task message.
func taskKey(qname, id string) string {
	return fmt.Sprintf("%s%s", taskKeyPrefix(qname), id)
}

// pendingKey returns a redis key for the given queue name.
func pendingKey(qname string) string {
	return fmt.Sprintf("%spending", queueKeyPrefix(qname))
}

// runningKey returns a redis key for the given queue name.
func runningKey(qname string) string {
	return fmt.Sprintf("%srunning", queueKeyPrefix(qname))
}

// failedKey returns a redis key for the given queue name.
func failedKey(qname string) string {
	return fmt.Sprintf("%sfailed", queueKeyPrefix(qname))
}

// scheduledKey returns a redis key for the scheduled tasks.
func scheduledKey(qname string) string {
	return fmt.Sprintf("%sscheduled", queueKeyPrefix(qname))
}

// retryKey returns a redis key for the retry tasks.
func retryKey(qname string) string {
	return fmt.Sprintf("%sretry", queueKeyPrefix(qname))
}

// getAllTasksCmd fetches all tasks with an offset and limit.
//
// Input:
// KEYS[1] -> gotama:<qname>:t:*
// --
// ARGV[1] -> offset
// ARGV[2] -> limit
//
// Output:
// Returns {total_keys, paginated_keys}
var getAllTasksCmd = redis.NewScript(`
local keys = redis.call("KEYS", KEYS[1])
local sorted_keys = {}
for i, key in ipairs(keys) do
	local created_at = redis.call("HGET", key, "created_at")
	local msg = redis.call("HGET", key, "msg")
	sorted_keys[i] = {tonumber(created_at) or 0, key , msg}
end
local function customSort(a, b)
	if a[1] == b[1] then
		return a[2] < b[2] -- Sort by key ASC
	else
		return a[1] > b[1] -- Sort by created_at DESC
	end
end
table.sort(sorted_keys, customSort)

local total_keys = #sorted_keys
local start_index = ARGV[1] + 1
local end_index = math.min(ARGV[1] + ARGV[2], total_keys)
local paginated_keys = {}
for i = start_index, end_index do
	paginated_keys[i - start_index + 1] = sorted_keys[i][3]
end
return {total_keys, paginated_keys}
`)

// GetAllTasks fetches tasks with a given offset.
func (r *RDB) GetAllTasks(ctx context.Context, offset int, limit int) (int64, []*task.Message, error) {
	keys := []string{
		taskKey(task.QueueDefault, "*"),
	}
	argv := []any{
		offset,
		limit,
	}
	logger.Info("Fetching all tasks", "offset", argv[0], "limit", argv[1])
	var tasks []*task.Message
	res, err := getAllTasksCmd.Run(ctx, r.client, keys, argv...).Result()
	if err != nil {
		return 0, tasks, fmt.Errorf("redis eval error: %v", err)
	}
	parsedRes, ok := res.([]any)
	if !ok {
		return 0, tasks, fmt.Errorf("unexpected return value from Lua script: %v", res)
	}

	totalTasks, ok := parsedRes[0].(int64)
	if !ok {
		return 0, tasks, fmt.Errorf("unexpected return total_keys from Lua script: %v", parsedRes)
	}

	encodedMsgs, ok := parsedRes[1].([]any)
	if !ok {
		return 0, tasks, fmt.Errorf("unexpected return paginated_keys from Lua script: %v", parsedRes)
	}

	for _, encoded := range encodedMsgs {
		encodedStr, ok := encoded.(string)
		if !ok {
			return 0, tasks, fmt.Errorf("error trying to cast %v to string", encoded)
		}
		msg, err := task.DecodeMessage(encodedStr)
		if err != nil {
			logger.Error("Error decoding msg", "error", err)
			return 0, tasks, err
		}
		tasks = append(tasks, msg)
	}

	return totalTasks, tasks, nil
}

// GetTask fetches a task by its ID.
func (r *RDB) GetTask(ctx context.Context, taskID string) (*task.Message, error) {
	encoded, err := r.client.HGet(ctx, taskKey(task.QueueDefault, taskID), "msg").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	msg, err := task.DecodeMessage(encoded)
	if err != nil {
		logger.Error("Error decoding msg", "error", err)
		return nil, err
	}

	return msg, nil
}

// enqueueTaskCmd enqueues a given task message.
//
// Input:
// KEYS[1] -> gotama:<qname>:t:<task_id>
// KEYS[2] -> gotama:<qname>:pending
// KEYS[3] -> gotama:<qname>:scheduled
// --
// ARGV[1] -> task message data
// ARGV[2] -> task ID
// ARGV[3] -> current unix time in milli sec
// ARGV[4] -> period in milli sec
// ARGV[5] -> type, RECURRING or ONCE
//
// Output:
// Returns 1 if successfully enqueued
// Returns 0 if task ID already exists
var enqueueTaskCmd = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
    return 0
end
redis.call("HSET", KEYS[1],
           "msg", ARGV[1],
           "status", "pending",
           "pending_since", ARGV[3],
           "created_at", ARGV[3],
           "period", ARGV[4])
redis.call("LPUSH", KEYS[2], ARGV[2])
if ARGV[5] == "RECURRING" then
    redis.call("LPUSH", KEYS[3], ARGV[2])
end
return 1
`)

const KeyQueues = "queues" // SET

// EnqueueTask adds the given task to the pending list of the queue.
func (r *RDB) EnqueueTask(ctx context.Context, msg *task.Message) error {
	encoded, err := task.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("cannot encode message: %v", err)
	}
	if err := r.client.SAdd(ctx, KeyQueues, msg.Queue).Err(); err != nil {
		return err
	}
	keys := []string{
		taskKey(msg.Queue, msg.ID),
		pendingKey(msg.Queue),
		scheduledKey(msg.Queue),
	}
	argv := []any{
		encoded,
		msg.ID,
		r.clock.Now().UnixMilli(),
		msg.Period.Milliseconds(),
		msg.Type.String(),
	}
	logger.Info("Adding task", "id", keys[0], "queue", keys[1])
	n, err := r.runScriptWithErrorCode(ctx, enqueueTaskCmd, keys, argv...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("task id already exists")
	}
	return nil
}

// Input:
// KEYS[1] -> gotama:<qname>:pending
// KEYS[2] -> gotama:<qname>:running
// --
// ARGV[1] -> task key prefix
//
// Output:
// Returns nil if no processable task is found in the given queue.
// Returns an encoded TaskMessage.
var dequeueTaskCmd = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
    local id = redis.call("RPOPLPUSH", KEYS[1], KEYS[2])
    if id then
        local key = ARGV[1] .. id
        redis.call("HSET", key, "status", "running")
        return redis.call("HGET", key, "msg")
    end
end
return nil`)

func (r *RDB) DequeueTask(ctx context.Context, qname string) (*task.Message, error) {
	keys := []string{
		pendingKey(qname),
		runningKey(qname),
	}
	argv := []any{
		taskKeyPrefix(qname),
	}
	encoded, err := dequeueTaskCmd.Run(ctx, r.client, keys, argv...).Result()
	if errors.Is(err, redis.Nil) {
		return nil, base.ErrorNoTasksInQueue
	} else if err != nil {
		return nil, fmt.Errorf("redis eval error: %v", err)
	}

	encodedStr, ok := encoded.(string)
	if !ok {
		return nil, fmt.Errorf("error trying to cast %v to string", encoded)
	}

	msg, err := task.DecodeMessage(encodedStr)
	if err != nil {
		logger.Error("Error decoding msg", "error", err)
		return nil, err
	}

	return msg, nil
}

// updateTaskCmd enqueues a given task message.
//
// Input:
// KEYS[1] -> gotama:<qname>:t:<task_id>
// KEYS[2] -> gotama:<qname>:scheduled
// --
// ARGV[1] -> task message data
// ARGV[2] -> period in milli s
// ARGV[3] -> task type - ONCE or RECURRING
// ARGV[4] -> task id
//
// Output:
// Returns 1 if successfully enqueued
// Returns 0 if task ID does not exist
var updateTaskCmd = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
    return 0
end
redis.call("HSET", KEYS[1],
           "msg", ARGV[1],
           "period", ARGV[2])
redis.call("LREM", KEYS[2], 0, ARGV[4])
if ARGV[3] == "RECURRING" then
    redis.call("LPUSH", KEYS[2], ARGV[4])
end
return 1
`)

// UpdateTask adds the given task to the pending list of the queue.
func (r *RDB) UpdateTask(ctx context.Context, msg *task.Message) error {
	encoded, err := task.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("cannot encode message: %v", err)
	}
	keys := []string{
		taskKey(msg.Queue, msg.ID),
		scheduledKey(msg.Queue),
	}
	argv := []any{
		encoded,
		msg.Period.Milliseconds(),
		msg.Type.String(),
		msg.ID,
	}
	logger.Info("Updating task", "id", keys[0])
	n, err := r.runScriptWithErrorCode(ctx, updateTaskCmd, keys, argv...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("task id does not exist")
	}
	return nil
}

// KEYS[1] -> gotama:<qname>:t:<task_id>
// KEYS[2] -> gotama:<qname>:pending
// KEYS[3] -> gotama:<qname>:scheduled
// KEYS[4] -> gotama:<qname>:retry
// -------
// ARGV[1] -> task ID
var removeCmd = redis.NewScript(`
redis.call("LREM", KEYS[2], 0, ARGV[1])
redis.call("LREM", KEYS[3], 0, ARGV[1])
redis.call("LREM", KEYS[4], 0, ARGV[1])
if redis.call("DEL", KEYS[1]) == 0 then
    return redis.error_reply("NOT FOUND")
end
return redis.status_reply("OK")
`)

// RemoveTask deletes the task from all queues and the task itself
func (r *RDB) RemoveTask(ctx context.Context, taskID string) error {
	keys := []string{
		taskKey(task.QueueDefault, taskID),
		pendingKey(task.QueueDefault),
		scheduledKey(task.QueueDefault),
		retryKey(task.QueueDefault),
	}

	argv := []any{
		taskID,
	}

	return r.runScript(ctx, removeCmd, keys, argv...)
}

// KEYS[1] -> gotama:<qname>:running
// KEYS[2] -> gotama:<qname>:retry
// KEYS[3] -> gotama:<qname>:t:<task_id>
// -------
// ARGV[1] -> task ID
var scheduleTaskRetryCmd = redis.NewScript(`
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
    return redis.error_reply("NOT FOUND")
end
redis.call("LREM", KEYS[2], 0, ARGV[1])
redis.call("LPUSH", KEYS[2], ARGV[1])
redis.call("HSET", KEYS[3], "status", "retry")
return redis.status_reply("OK")`)

// RequeueTaskRetry moves the task from running queue to the retry queue.
func (r *RDB) RequeueTaskRetry(ctx context.Context, msg *task.Message) error {
	keys := []string{
		runningKey(msg.Queue),
		retryKey(msg.Queue),
		taskKey(msg.Queue, msg.ID),
	}
	return r.runScript(ctx, scheduleTaskRetryCmd, keys, msg.ID)
}

// KEYS[1] -> gotama:<qname>:running
// KEYS[2] -> gotama:<qname>:failed
// KEYS[3] -> gotama:<qname>:t:<task_id>
// -------
// ARGV[1] -> task ID
var requeueTaskFailedCmd = redis.NewScript(`
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
    return redis.error_reply("NOT FOUND")
end
redis.call("LPUSH", KEYS[2], ARGV[1])
redis.call("HSET", KEYS[3], "status", "failed")
return redis.status_reply("OK")`)

// RequeueTaskFailed moves the task from running queue to the failed queue.
func (r *RDB) RequeueTaskFailed(ctx context.Context, msg *task.Message) error {
	keys := []string{
		runningKey(msg.Queue),
		failedKey(msg.Queue),
		taskKey(msg.Queue, msg.ID),
	}
	return r.runScript(ctx, requeueTaskFailedCmd, keys, msg.ID)
}

// KEYS[1] -> gotama:<qname>:running
// KEYS[2] -> gotama:<qname>:t:<task_id>
// -------
// ARGV[1] -> task ID
var markTaskAsCompleteCmd = redis.NewScript(`
if redis.call("LREM", KEYS[1], 0, ARGV[1]) == 0 then
    return redis.error_reply("NOT FOUND")
end
redis.call("HSET", KEYS[2], "status", "succeeded")
return redis.status_reply("OK")`)

// MarkTaskAsComplete moves the task from running queue to the failed queue.
func (r *RDB) MarkTaskAsComplete(ctx context.Context, msg *task.Message) error {
	keys := []string{
		runningKey(msg.Queue),
		taskKey(msg.Queue, msg.ID),
	}
	return r.runScript(ctx, markTaskAsCompleteCmd, keys, msg.ID)
}

// KEYS[1] -> gotama:<qname>:scheduled
// KEYS[2] -> gotama:<qname>:pending
// KEYS[3] -> gotama:<qname>:t:
// KEYS[4] -> gotama:<qname>:retry
// -------
// ARGV[1] -> current time in unix milli sec
var enqueueScheduledTasksCmd = redis.NewScript(`
local retry_task_ids = redis.call("LRANGE", KEYS[4], 0, -1)

for _, task_id in ipairs(retry_task_ids) do
    local task_key = KEYS[3] .. task_id
    local status = redis.call("HGET", task_key, "status")

    if status == "retry" then
        -- Priorities with RPUSH
        redis.call("RPUSH", KEYS[2], task_id)
        redis.call("HSET", task_key, "pending_since", ARGV[1])
        redis.call("HSET", task_key, "status", "pending")
    end
end

local scheduled_task_ids = redis.call("LRANGE", KEYS[1], 0, -1)

for _, task_id in ipairs(scheduled_task_ids) do
    local task_key = KEYS[3] .. task_id
    local pending_since = tonumber(redis.call("HGET", task_key, "pending_since"))
    local status = redis.call("HGET", task_key, "status")
    local period = tonumber(redis.call("HGET", task_key, "period"))
    local current_time = tonumber(ARGV[1])

    if status ~= "failed" and status ~= "retry" and status ~= "running" and status ~= "pending" and current_time > pending_since + period then
        redis.call("LPUSH", KEYS[2], task_id)
        redis.call("HSET", task_key, "pending_since", ARGV[1])
        redis.call("HSET", task_key, "status", "pending")
    end
end

return redis.status_reply("OK")`)

// EnqueueScheduledTasks checks for scheduled tasks and pass them to the pending queue
func (r *RDB) EnqueueScheduledTasks(ctx context.Context) error {
	keys := []string{
		scheduledKey(task.QueueDefault),
		pendingKey(task.QueueDefault),
		taskKey(task.QueueDefault, ""),
		retryKey(task.QueueDefault),
	}

	argv := []any{
		r.clock.Now().UnixMilli(),
	}
	return r.runScript(ctx, enqueueScheduledTasksCmd, keys, argv)
}
