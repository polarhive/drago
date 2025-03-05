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
	"go.uber.org/zap"
)

type Workflow struct {
	dag      *dag.DAG
	workers  int
	stepMode bool
	logger   *zap.Logger
}

func NewWorkflow(d *dag.DAG, workers int, stepMode bool, logger *zap.Logger) *Workflow {
	return &Workflow{
		dag:      d,
		workers:  workers,
		stepMode: stepMode,
		logger:   logger,
	}
}

func (w *Workflow) Run() {
	if w.stepMode {
		w.RunStepMode()
	} else {
		w.RunParallel()
	}
}

func (w *Workflow) RunParallel() {
	executionOrder := w.topologicalSort()
	sem := make(chan struct{}, w.workers)
	var wg sync.WaitGroup

	for _, nodeID := range executionOrder {
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

func (w *Workflow) RunStepMode() {
	executionOrder := w.topologicalSort()
	scanner := bufio.NewScanner(os.Stdin)

	for _, nodeID := range executionOrder {
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

func (w *Workflow) processNode(node *nodes.Node) {
	w.updateNodeState(node, nodes.Running)
	params := nodes.NodeTypes[node.Type]

	success := w.executeNode(params)

	switch {
	case success:
		w.updateNodeState(node, nodes.Success)
	case node.Retries < params.MaxRetries:
		w.updateNodeState(node, nodes.Retrying)
		node.Retries++
		time.Sleep(time.Duration(node.Retries) * params.RetryBackoff)
		w.processNode(node)
	default:
		w.updateNodeState(node, nodes.Failed)
	}
}

func (w *Workflow) executeNode(params nodes.ExecutionParameters) bool {
	time.Sleep(params.BaseDelay)
	return rand.Float32() < params.SuccessRate
}

func (w *Workflow) updateNodeState(node *nodes.Node, state nodes.NodeState) {
	w.dag.Mu.Lock()
	defer w.dag.Mu.Unlock()
	node.State = state
	w.logger.Info("Node state updated",
		zap.String("node", node.ID),
		zap.String("state", string(state)),
	)
}

func (w *Workflow) topologicalSort() []string {
	w.dag.Mu.RLock()
	defer w.dag.Mu.RUnlock()

	inDegree := make(map[string]int)
	queue := make([]string, 0)
	order := make([]string, 0)

	for _, node := range w.dag.Nodes {
		inDegree[node.ID] = len(node.Dependencies)
	}

	for nodeID, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, nodeID)
		}
	}

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
