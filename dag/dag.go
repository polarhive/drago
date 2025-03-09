package dag

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/polarhive/drago/nodes"
	"go.uber.org/zap"
)

// DAG represents a directed acyclic graph (DAG) for workflow execution.
type DAG struct {
	Nodes  map[string]*nodes.Node // Stores all nodes by ID
	Edges  map[string][]string    // Adjacency list representation of dependencies
	Mu     sync.RWMutex           // Mutex for concurrent access
	Logger *zap.Logger            // Logger for debugging and errors
}

// NewDAG creates a new DAG instance.
func NewDAG(logger *zap.Logger) *DAG {
	return &DAG{
		Nodes:  make(map[string]*nodes.Node),
		Edges:  make(map[string][]string),
		Logger: logger,
	}
}

// AddNode adds a node to the DAG and registers its dependencies.
func (d *DAG) AddNode(node *nodes.Node) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	node.State = nodes.Pending
	d.Nodes[node.ID] = node

	for _, dep := range node.Dependencies {
		if _, exists := d.Edges[dep]; !exists {
			d.Edges[dep] = []string{}
		}
		d.Edges[dep] = append(d.Edges[dep], node.ID)
	}
}

// LoadFromJSON loads a DAG structure from a JSON file.
func (d *DAG) LoadFromJSON(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var nodeList []*nodes.Node
	if err := json.Unmarshal(file, &nodeList); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	d.Mu.Lock()
	defer d.Mu.Unlock()

	for _, node := range nodeList {
		d.Nodes[node.ID] = node
		for _, dep := range node.Dependencies {
			d.Edges[dep] = append(d.Edges[dep], node.ID)
		}
	}
	return nil
}
