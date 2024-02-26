package rdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/engpetarmarinov/gotama/internal/base"
	"github.com/engpetarmarinov/gotama/internal/broker"
	"github.com/engpetarmarinov/gotama/internal/task"
	"github.com/engpetarmarinov/gotama/internal/timeutil"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type RDB struct {
	broker.Broker
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

func (r *RDB) runScript(ctx context.Context, script *redis.Script, keys []string, args ...interface{}) error {
	if err := script.Run(ctx, r.client, keys, args...).Err(); err != nil {
		return errors.New(fmt.Sprintf("redis eval error: %v", err))
	}
	return nil
}

func (r *RDB) runScriptWithErrorCode(ctx context.Context, script *redis.Script, keys []string, args ...interface{}) (int64, error) {
	res, err := script.Run(ctx, r.client, keys, args...).Result()
	if err != nil {
		return 0, errors.New(fmt.Sprintf("redis eval error: %v", err))
	}
	n, ok := res.(int64)
	if !ok {
		return 0, errors.New(fmt.Sprintf("unexpected return value from Lua script: %v", res))
	}
	return n, nil
}

// QueueKeyPrefix returns a prefix for all keys in the given queue.
func QueueKeyPrefix(qname string) string {
	return fmt.Sprintf("gotama:%s:", qname)
}

// TaskKeyPrefix returns a prefix for task key.
func TaskKeyPrefix(qname string) string {
	return fmt.Sprintf("%st:", QueueKeyPrefix(qname))
}

// TaskKey returns a redis key for the given task message.
func TaskKey(qname, id string) string {
	return fmt.Sprintf("%s%s", TaskKeyPrefix(qname), id)
}

// PendingKey returns a redis key for the given queue name.
func PendingKey(qname string) string {
	return fmt.Sprintf("%spending", QueueKeyPrefix(qname))
}

// RunningKey returns a redis key for the given queue name.
func RunningKey(qname string) string {
	return fmt.Sprintf("%srunning", QueueKeyPrefix(qname))
}

// FailedKey returns a redis key for the given queue name.
func FailedKey(qname string) string {
	return fmt.Sprintf("%sfailed", QueueKeyPrefix(qname))
}

// ScheduledKey returns a redis key for the scheduled tasks.
func ScheduledKey(qname string) string {
	return fmt.Sprintf("%sscheduled", QueueKeyPrefix(qname))
}

// RetryKey returns a redis key for the retry tasks.
func RetryKey(qname string) string {
	return fmt.Sprintf("%sretry", QueueKeyPrefix(qname))
}

const (
	KeyWorkers    = "workers"    // ZSET
	KeySchedulers = "schedulers" // ZSET
	KeyQueues     = "queues"     // SET
)

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
    local keys = redis.call('KEYS', KEYS[1])
    local sorted_keys = {}
    for i, key in ipairs(keys) do
        local pending_since = redis.call('HGET', key, 'pending_since')
        local msg = redis.call('HGET', key, 'msg')
        sorted_keys[i] = {tonumber(pending_since) or 0, msg}
    end
    table.sort(sorted_keys, function(a, b) return a[1] > b[1] end)
    
    local total_keys = #sorted_keys
    local start_index = ARGV[1] + 1
    local end_index = math.min(ARGV[1] + ARGV[2], total_keys)
    local paginated_keys = {}
    for i = start_index, end_index do
        paginated_keys[i - start_index + 1] = sorted_keys[i][2]
    end
    return {total_keys, paginated_keys}
`)

// GetAllTasks fetches tasks with a given offset.
func (r *RDB) GetAllTasks(ctx context.Context, offset int, limit int) (int64, []*task.Message, error) {
	keys := []string{
		TaskKey(task.QueueDefault, "*"),
	}
	argv := []interface{}{
		offset,
		limit,
	}
	slog.Info("Fetching all tasks", "offset", argv[0], "limit", argv[1])
	var tasks []*task.Message
	res, err := getAllTasksCmd.Run(ctx, r.client, keys, argv...).Result()
	if err != nil {
		return 0, tasks, errors.New(fmt.Sprintf("redis eval error: %v", err))
	}
	parsedRes, ok := res.([]interface{})
	if !ok {
		return 0, tasks, errors.New(fmt.Sprintf("unexpected return value from Lua script: %v", res))
	}

	totalTasks, ok := parsedRes[0].(int64)
	if !ok {
		return 0, tasks, errors.New(fmt.Sprintf("unexpected return total_keys from Lua script: %v", parsedRes))
	}

	encodedMsgs, ok := parsedRes[1].([]interface{})
	if !ok {
		return 0, tasks, errors.New(fmt.Sprintf("unexpected return paginated_keys from Lua script: %v", parsedRes))
	}

	for _, encoded := range encodedMsgs {
		encodedStr, ok := encoded.(string)
		if !ok {
			return 0, tasks, errors.New(fmt.Sprintf("error trying to cast %v to string", encoded))
		}
		msg, err := task.DecodeMessage(encodedStr)
		if err != nil {
			slog.Error("Error decoding msg", "err", err)
			return 0, tasks, err
		}
		tasks = append(tasks, msg)
	}

	return totalTasks, tasks, nil
}

// GetTask fetches a task by its ID.
func (r *RDB) GetTask(ctx context.Context, taskID string) (*task.Message, error) {
	encoded, err := r.client.HGet(ctx, TaskKey(task.QueueDefault, taskID), "msg").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("not found")
		}
		return nil, err
	}
	msg, err := task.DecodeMessage(encoded)
	if err != nil {
		slog.Error("Error decoding msg", "err", err)
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
           "period", ARGV[4])
redis.call("LPUSH", KEYS[2], ARGV[2])
if ARGV[5] == "RECURRING" then
    redis.call("LPUSH", KEYS[3], ARGV[2])
end
return 1
`)

// EnqueueTask adds the given task to the pending list of the queue.
func (r *RDB) EnqueueTask(ctx context.Context, msg *task.Message) error {
	encoded, err := task.EncodeMessage(msg)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot encode message: %v", err))
	}
	if err := r.client.SAdd(ctx, KeyQueues, msg.Queue).Err(); err != nil {
		return err
	}
	keys := []string{
		TaskKey(msg.Queue, msg.ID),
		PendingKey(msg.Queue),
		ScheduledKey(msg.Queue),
	}
	argv := []interface{}{
		encoded,
		msg.ID,
		r.clock.Now().UnixMilli(),
		msg.Period.Milliseconds(),
		msg.Type.String(),
	}
	slog.Info("Adding task", "id", keys[0], "queue", keys[1])
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
		PendingKey(qname),
		RunningKey(qname),
	}
	argv := []interface{}{
		TaskKeyPrefix(qname),
	}
	encoded, err := dequeueTaskCmd.Run(ctx, r.client, keys, argv...).Result()
	if errors.Is(err, redis.Nil) {
		return nil, base.ErrorNoTasksInQueue
	} else if err != nil {
		return nil, errors.New(fmt.Sprintf("redis eval error: %v", err))
	}

	encodedStr, ok := encoded.(string)
	if !ok {
		return nil, errors.New(fmt.Sprintf("error trying to cast %v to string", encoded))
	}

	msg, err := task.DecodeMessage(encodedStr)
	if err != nil {
		slog.Error("Error decoding msg", "err", err)
		return nil, err
	}

	return msg, nil
}

// updateTaskCmd enqueues a given task message.
//
// Input:
// KEYS[1] -> gotama:<qname>:t:<task_id>
// --
// ARGV[1] -> task message data
// ARGV[2] -> period in milli s
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
return 1
`)

// UpdateTask adds the given task to the pending list of the queue.
func (r *RDB) UpdateTask(ctx context.Context, msg *task.Message) error {
	encoded, err := task.EncodeMessage(msg)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot encode message: %v", err))
	}
	keys := []string{
		TaskKey(msg.Queue, msg.ID),
	}
	argv := []interface{}{
		encoded,
		msg.Period.Milliseconds(),
	}
	slog.Info("Updating task", "id", keys[0])
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
		TaskKey(task.QueueDefault, taskID),
		PendingKey(task.QueueDefault),
		ScheduledKey(task.QueueDefault),
		RetryKey(task.QueueDefault),
	}

	argv := []interface{}{
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
		RunningKey(msg.Queue),
		RetryKey(msg.Queue),
		TaskKey(msg.Queue, msg.ID),
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
		RunningKey(msg.Queue),
		FailedKey(msg.Queue),
		TaskKey(msg.Queue, msg.ID),
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
		RunningKey(msg.Queue),
		TaskKey(msg.Queue, msg.ID),
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
local scheduled_task_ids = redis.call('LRANGE', KEYS[1], 0, -1)

for _, task_id in ipairs(scheduled_task_ids) do
    local task_key = KEYS[3] .. task_id
    local pending_since = tonumber(redis.call('HGET', task_key, 'pending_since'))
    local status = redis.call('HGET', task_key, 'status')
    local period = tonumber(redis.call('HGET', task_key, 'period'))
    local current_time = tonumber(ARGV[1])

    if status ~= 'pending' and current_time > pending_since + period then
        -- Priorities with RPUSH
        redis.call('RPUSH', KEYS[2], task_id)
        redis.call('HSET', task_key, 'pending_since', ARGV[1])
        redis.call('HSET', task_key, 'status', 'pending')
    end
end

local retry_task_ids = redis.call('LRANGE', KEYS[4], 0, -1)

for _, task_id in ipairs(retry_task_ids) do
    local task_key = KEYS[3] .. task_id
    local status = redis.call('HGET', task_key, 'status')

    if status ~= 'failed' and status ~= 'pending' then
        -- Priorities with RPUSH
        redis.call('RPUSH', KEYS[2], task_id)
        redis.call('HSET', task_key, 'pending_since', ARGV[1])
        redis.call('HSET', task_key, 'status', 'pending')
    end
end

return redis.status_reply("OK")`)

// EnqueueScheduledTasks checks for scheduled tasks and pass them to the pending queue
func (r *RDB) EnqueueScheduledTasks(ctx context.Context) error {
	keys := []string{
		ScheduledKey(task.QueueDefault),
		PendingKey(task.QueueDefault),
		TaskKey(task.QueueDefault, ""),
		RetryKey(task.QueueDefault),
	}

	argv := []interface{}{
		r.clock.Now().UnixMilli(),
	}
	return r.runScript(ctx, enqueueScheduledTasksCmd, keys, argv)
}
