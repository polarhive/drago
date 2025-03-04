package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

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
	Dependencies []string
	State        NodeState
	Retries      int
	Type         string
}

type DAG struct {
	mu     sync.RWMutex
	nodes  map[string]*Node
	edges  map[string][]string
	logger *zap.Logger
}

type Workflow struct {
	dag      *DAG
	workers  int
	stepMode bool
	logger   *zap.Logger
}

func main() {
	workers := flag.Int("workers", 2, "Number of paralel workers")
	stepMode := flag.Bool("step", false, "Step mode")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dag := &DAG{
		nodes:  make(map[string]*Node),
		edges:  make(map[string][]string),
		logger: logger,
	}

	// ex: workflow setup
	createSampleDAG(dag)

	workflow := &Workflow{
		dag:      dag,
		workers:  *workers,
		stepMode: *stepMode,
		logger:   logger,
	}

	if *stepMode {
		workflow.RunStepMode()
	} else {
		workflow.RunParallel()
	}
}

func createSampleDAG(dag *DAG) {
	nodes := []*Node{
		{ID: "start", Type: "trigger"},
		{ID: "process", Type: "compute", Dependencies: []string{"start"}},
		{ID: "decision", Type: "decision", Dependencies: []string{"process"}},
		{ID: "api-call", Type: "api", Dependencies: []string{"decision"}},
		{ID: "end", Type: "action", Dependencies: []string{"api-call"}},
	}

	for _, node := range nodes {
		dag.AddNode(node)
	}
}

func (d *DAG) AddNode(node *Node) {
	d.mu.Lock()
	defer d.mu.Unlock()
	node.State = Pending
	d.nodes[node.ID] = node
	d.edges[node.ID] = node.Dependencies
}

func (w *Workflow) RunParallel() {
	executionOrder := w.topologicalSort()
	sem := make(chan struct{}, w.workers)
	var wg sync.WaitGroup

	for _, nodeID := range executionOrder {
		node := w.dag.nodes[nodeID]
		wg.Add(1)
		sem <- struct{}{}

		go func(n *Node) {
			defer wg.Done()
			defer func() { <-sem }()
			w.processNode(n)
		}(node)
	}

	wg.Wait()
	w.logger.Info("Workflow completed")
}

func (w *Workflow) RunStepMode() {
	executionOrder := w.topologicalSort()
	scanner := bufio.NewScanner(os.Stdin)

	for _, nodeID := range executionOrder {
		node := w.dag.nodes[nodeID]
		w.logger.Info("Next node",
			zap.String("node", node.ID),
			zap.Strings("dependencies", node.Dependencies),
		)

		fmt.Print("Press Enter to execute...")
		scanner.Scan()
		w.processNode(node)
	}
}

func (w *Workflow) processNode(node *Node) {
	w.updateNodeState(node, Running)

	// Simulate different processing based on node type
	var success bool
	switch node.Type {
	case "trigger":
		success = w.handleTrigger(node)
	case "compute":
		success = w.handleCompute(node)
	case "decision":
		success = w.handleDecision(node)
	case "api":
		success = w.handleAPI(node)
	default:
		success = true
	}

	if success {
		w.updateNodeState(node, Success)
	} else if node.Retries < 3 {
		w.updateNodeState(node, Retrying)
		node.Retries++
		w.processNode(node)
	} else {
		w.updateNodeState(node, Failed)
	}
}

func (w *Workflow) updateNodeState(node *Node, state NodeState) {
	w.dag.mu.Lock()
	defer w.dag.mu.Unlock()
	node.State = state
	w.logger.Info("Node state updated",
		zap.String("node", node.ID),
		zap.String("state", string(state)),
	)
}

func (w *Workflow) topologicalSort() []string {
	w.dag.mu.RLock()
	defer w.dag.mu.RUnlock()

	inDegree := make(map[string]int)
	queue := make([]string, 0)
	order := make([]string, 0)

	// Initialize in-degree
	for nodeID := range w.dag.nodes {
		inDegree[nodeID] = 0
	}

	// Find start nodes
	for nodeID, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, nodeID)
		}
	}

	// Topo sort
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, neighbor := range w.dag.edges[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	return order
}

// Simulated node handlers
func (w *Workflow) handleTrigger(node *Node) bool {
	time.Sleep(time.Millisecond * 200)
	return true // Always succeed
}

func (w *Workflow) handleCompute(node *Node) bool {
	time.Sleep(time.Millisecond * 500)
	return rand.Float32() < 0.8 // 80% success rate
}

func (w *Workflow) handleDecision(node *Node) bool {
	time.Sleep(time.Millisecond * 100)
	return true // Always succeed
}

func (w *Workflow) handleAPI(node *Node) bool {
	time.Sleep(time.Millisecond * 300)
	return rand.Float32() < 0.6 // 60% success rate
}
