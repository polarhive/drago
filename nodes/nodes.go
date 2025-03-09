package nodes

import "time"

// NodeState represents the current execution state of a node.
type NodeState string

const (
	Pending  NodeState = "pending"
	Running  NodeState = "running"
	Success  NodeState = "success"
	Failed   NodeState = "failed"
	Retrying NodeState = "retrying"
)

// Node represents a single task in the workflow.
type Node struct {
	ID           string   `json:"id"`           // Unique node identifier
	Type         string   `json:"type"`         // Type of node (e.g., "task", "decision")
	Dependencies []string `json:"dependencies"` // List of parent nodes
	InputKeys    []string `json:"input_keys"`   // Required input keys from KV store
	OutputKey    string   `json:"output_key"`   // Output key for KV store
	Retries      int      `json:"retries"`      // Current retry count
	State        NodeState `json:"state"`       // Execution state
}

// ExecutionParameters defines execution properties for different node types.
type ExecutionParameters struct {
	MaxRetries   int           // Maximum number of retries allowed
	RetryBackoff time.Duration // Backoff duration before retrying
	BaseDelay    time.Duration // Base execution delay
	SuccessRate  float32       // Probability of successful execution
}

// Predefined execution parameters for node types
var NodeTypes = map[string]ExecutionParameters{
	"task": {
		MaxRetries:   3,
		RetryBackoff: 2 * time.Second,
		BaseDelay:    1 * time.Second,
		SuccessRate:  0.9,
	},
	"decision": {
		MaxRetries:   1,
		RetryBackoff: 1 * time.Second,
		BaseDelay:    500 * time.Millisecond,
		SuccessRate:  1.0, // Always succeeds
	},
}
