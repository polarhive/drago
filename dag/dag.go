package dag

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/polarhive/drago/nodes"
	"go.uber.org/zap"
)


type DAG struct {
	Mu     sync.RWMutex
	Nodes  map[string]*nodes.Node
	Edges  map[string][]string
	logger *zap.Logger
}

func NewDAG(logger *zap.Logger) *DAG {
	return &DAG{
		Nodes:  make(map[string]*nodes.Node),
		Edges:  make(map[string][]string),
		logger: logger,
	}
}

func (d *DAG) AddNode(node *nodes.Node) {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	
	node.State = nodes.Pending
	d.Nodes[node.ID] = node
	
	for _, dep := range node.Dependencies {
		d.Edges[dep] = append(d.Edges[dep], node.ID)
	}
}

func (d *DAG) LoadFromJSON(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var nodes []*nodes.Node
	if err := json.Unmarshal(file, &nodes); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	for _, node := range nodes {
		d.AddNode(node)
	}
	return nil
}