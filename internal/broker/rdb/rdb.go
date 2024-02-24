package rdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func NewRDB(client redis.UniversalClient) *RDB {
	return &RDB{
		client: client,
		clock:  timeutil.NewRealClock(),
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

// TODO: do I need active tasks?
// ActiveKey returns a redis key for the active tasks.
func ActiveKey(qname string) string {
	return fmt.Sprintf("%sactive", QueueKeyPrefix(qname))
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
// KEYS[1] -> gotama:{<qname>}:t:*
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
		var msg task.Message
		encodedStr, ok := encoded.(string)
		if !ok {
			return 0, tasks, errors.New(fmt.Sprintf("error trying to cast %v to []byte", encoded))
		}
		err := json.Unmarshal([]byte(encodedStr), &msg)
		if err != nil {
			slog.Error("Error unmarshalling msg", "err", err)
			return 0, tasks, err
		}
		tasks = append(tasks, &msg)
	}

	return totalTasks, tasks, nil
}

// GetTask fetches a task by its ID.
func (r *RDB) GetTask(ctx context.Context, taskID string) (*task.Message, error) {
	encoded := r.client.HGet(ctx, TaskKey(task.QueueDefault, taskID), "msg").Val()
	var msg task.Message
	err := json.Unmarshal([]byte(encoded), &msg)
	if err != nil {
		slog.Error("Error unmarshalling msg", "err", err)
		return nil, err
	}

	return &msg, nil
}

// enqueueTaskCmd enqueues a given task message.
//
// Input:
// KEYS[1] -> gotama:{<qname>}:t:<task_id>
// KEYS[2] -> gotama:{<qname>}:pending
// --
// ARGV[1] -> task message data
// ARGV[2] -> task ID
// ARGV[3] -> current unix time in nano sec
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
           "state", "pending",
           "pending_since", ARGV[3])
redis.call("LPUSH", KEYS[2], ARGV[2])
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
	}
	argv := []interface{}{
		encoded,
		msg.ID,
		r.clock.Now().UnixNano(),
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
