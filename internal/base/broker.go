package base

type Broker interface {
	Ping() error
	Close() error
	//Enqueue(ctx context.Context, msg *TaskMessage) error
	//Dequeue(qnames ...string) (*TaskMessage, time.Time, error)
	//Remove(ctx context.Context, msg *TaskMessage) error
	//Complete(ctx context.Context, msg *TaskMessage) error
	//Schedule(ctx context.Context, msg *TaskMessage, processAt time.Time) error
	//WriteResult(qname, id string, data []byte) (n int, err error)
}
