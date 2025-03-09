package workflow

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/polarhive/drago/dag"
	"github.com/polarhive/drago/nodes"
	"github.com/polarhive/drago/storage"
	"go.uber.org/zap"
)

type Workflow struct {
	dag      *dag.DAG
	workers  int
	stepMode bool
	logger   *zap.Logger
	kv       *storage.KV
}

// Retrieve a value from key-value storage
func (w *Workflow) Get(key string) (any, bool) {
	return w.kv.Get(key)
}

// Store a value in key-value storage
func (w *Workflow) Set(key string, value any) {
	w.kv.Set(key, value)
}

// Create a new workflow instance
func NewWorkflow(d *dag.DAG, workers int, stepMode bool, logger *zap.Logger) *Workflow {
	return &Workflow{
		dag:      d,
		workers:  workers,
		stepMode: stepMode,
		logger:   logger,
		kv:       storage.NewKV(),
	}
}

// Execute the workflow
func (w *Workflow) Run() {
	executionOrder := w.topologicalSort()
	if w.stepMode {
		w.runStepMode(executionOrder)
	} else {
		w.runParallel(executionOrder)
	}
}

// Execute workflow nodes in parallel with worker limits
func (w *Workflow) runParallel(order []string) {
	sem := make(chan struct{}, w.workers)
	var wg sync.WaitGroup

	for _, nodeID := range order {
		node := w.dag.Nodes[nodeID]
		wg.Add(1)
		sem <- struct{}{}

		go func(n *nodes.Node) {
			defer wg.Done()
			defer func() { <-sem }()
			w.processNode(n)
		}(node)
	}

	wg.Wait()
	w.logger.Info("Workflow completed")
}

// Execute workflow nodes step by step (manual user confirmation)
func (w *Workflow) runStepMode(order []string) {
	scanner := bufio.NewScanner(os.Stdin)

	for _, nodeID := range order {
		node := w.dag.Nodes[nodeID]
		w.logger.Info("Next node",
			zap.String("node", node.ID),
			zap.Strings("dependencies", node.Dependencies),
		)

		fmt.Print("Press Enter to execute...")
		scanner.Scan()
		w.processNode(node)
	}
}

// Execute a node based on its type and input data
func (w *Workflow) processNode(node *nodes.Node) {
	w.updateNodeState(node, nodes.Running)

	inputs := make(map[string]any)
	for _, key := range node.InputKeys {
		if val, exists := w.Get(key); exists {
			inputs[key] = val
		}
	}

	params := nodes.NodeTypes[node.Type]
	success := w.executeNode(node, params, inputs)

	if success {
		w.updateNodeState(node, nodes.Success)
		if node.OutputKey != "" {
			w.Set(node.OutputKey, fmt.Sprintf("result-%s", node.ID))
		}
		return
	}

	// Retry logic
	if node.Retries < params.MaxRetries {
		node.Retries++
		time.Sleep(time.Duration(node.Retries) * params.RetryBackoff)
		w.processNode(node)
	} else {
		w.updateNodeState(node, nodes.Failed)
	}
}

// Simulate node execution based on success probability
func (w *Workflow) executeNode(node *nodes.Node, params nodes.ExecutionParameters, inputs map[string]interface{}) bool {
	w.logger.Debug("Executing node",
		zap.String("node", node.ID),
		zap.Any("inputs", inputs),
	)

	time.Sleep(params.BaseDelay)
	return rand.Float32() < params.SuccessRate
}

// Update the state of a node in the DAG
func (w *Workflow) updateNodeState(node *nodes.Node, state nodes.NodeState) {
	w.dag.Mu.Lock()
	defer w.dag.Mu.Unlock()
	node.State = state
	w.logger.Info("Node state updated",
		zap.String("node", node.ID),
		zap.String("state", string(state)),
	)
}

// Perform a topological sort to determine execution order
func (w *Workflow) topologicalSort() []string {
	w.dag.Mu.RLock()
	defer w.dag.Mu.RUnlock()

	inDegree := make(map[string]int)
	queue := make([]string, 0)
	order := make([]string, 0)

	// Calculate in-degrees for all nodes
	for _, node := range w.dag.Nodes {
		inDegree[node.ID] = len(node.Dependencies)
	}

	// Identify nodes with no dependencies (ready to execute)
	for nodeID, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, nodeID)
		}
	}

	// Process nodes in topological order
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		order = append(order, nodeID)

		for _, dependent := range w.dag.Edges[nodeID] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	return order
}
