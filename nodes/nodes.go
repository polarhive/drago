package nodes

import "time"

type NodeState string

const (
	Pending  NodeState = "pending"
	Running  NodeState = "running"
	Success  NodeState = "success"
	Failed   NodeState = "failed"
	Retrying NodeState = "retrying"
)

type Node struct {
	ID           string
	Type         string
	Dependencies []string
	State        NodeState
	Retries      int
}

type ExecutionParameters struct {
	BaseDelay    time.Duration
	SuccessRate  float32
	MaxRetries   int
	RetryBackoff time.Duration
}

var NodeTypes = map[string]ExecutionParameters{
	"trigger":  {200 * time.Millisecond, 1.0, 0, 0},
	"compute":  {500 * time.Millisecond, 0.8, 3, 100 * time.Millisecond},
	"decision": {100 * time.Millisecond, 1.0, 1, 50 * time.Millisecond},
	"api":      {300 * time.Millisecond, 0.6, 3, 200 * time.Millisecond},
	"action":   {100 * time.Millisecond, 0.9, 2, 100 * time.Millisecond},
}