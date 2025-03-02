package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

/*
- Trigger: A node that starts a workflow
- Condition: A node that checks a condition and decides which path to take; if-else condition or a decision tree (state machine / switch-case)
- Loop: A node that repeats a set of nodes for a certain number of times or until a condition is met. Typically used to process a list / array of items.
- Compute: A node that performs a computation (transform data, filter data, aggregate data, etc). This can be a simple JS script or an executable (firecracker, wasm, etc)
- Action: A node that performs an action(external / internal calls), outputs data, or triggers another workflow
*/

// Node represents a single unit of work in a workflow
type Node struct {
	ID           string    `json:"id"`
	State        NodeState `json:"state"`
	Dependencies []string  `json:"dependencies"`
}

// NodeState represents the state of a node in the workflow
type NodeState string

// Possible values for NodeState
const (
	Pending NodeState = "Pending"
	Running NodeState = "Running"
	Success NodeState = "Success"
	Failed  NodeState = "Failed"
)

// Workflow represents a collection of nodes connected as a DAG
// Nodes can be triggers, conditions, actions, etc
// Workflow is thread-safe [todo]
// Workflow execution is parallelized [todo]
type Workflow struct {
	Nodes map[string]*Node
	Mutex sync.Mutex // todo
}

func NewWorkflow() *Workflow {
	return &Workflow{Nodes: make(map[string]*Node)}
	// todo
}

// AddNode adds a new node to the workflow
func (w *Workflow) AddNode(node *Node) {
	w.Nodes[node.ID] = node
}

// Executes the workflow
func (w *Workflow) Run() {
	log.Println("Starting Workflow Execution")
	var wg sync.WaitGroup

	for _, node := range w.Nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			w.executeNode(n)
		}(node)
	}

	wg.Wait()
	log.Println("Workflow Execution Completed")
}

// Executes a single node in the workflow
func (w *Workflow) executeNode(node *Node) {
	log.Printf("Executing node: %s\n", node.ID)
	node.State = Running
	log.Printf("Node %s is in state: %s\n", node.ID, node.State)
	// todo
	node.State = Success
}

// LoadWorkflowFromJSON loads a workflow from a JSON file
func LoadWorkflowFromJSON(filename string) *Workflow {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var nodes []Node
	err = json.Unmarshal(file, &nodes)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	workflow := NewWorkflow()
	for _, node := range nodes {
		workflow.AddNode(&node)
	}
	return workflow
}

func main() {
	workflow := LoadWorkflowFromJSON("workflow.json")
	workflow.Run()
}
