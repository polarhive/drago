package dag

import (
	"fmt"
)

// Validate checks the DAG for missing dependencies, invalid node types, and cycles.
func (d *DAG) Validate() error {
	d.Mu.RLock()
	defer d.Mu.RUnlock()

	if err := d.checkDependenciesAndTypes(); err != nil {
		return err
	}
	if cyclic, cycle := d.detectCycles(); cyclic {
		return fmt.Errorf("cycle detected: %v", cycle)
	}
	return nil
}

// Combined check for missing dependencies and invalid node types.
func (d *DAG) checkDependenciesAndTypes() error {
	validTypes := map[string]struct{}{
		"trigger": {}, "compute": {}, "decision": {}, "api": {}, "action": {},
	}

	for _, node := range d.Nodes {
		if _, valid := validTypes[node.Type]; !valid {
			return fmt.Errorf("invalid node type '%s' for node %s", node.Type, node.ID)
		}
		for _, dep := range node.Dependencies {
			if _, exists := d.Nodes[dep]; !exists {
				return fmt.Errorf("node %s has missing dependency: %s", node.ID, dep)
			}
		}
	}
	return nil
}

// checks for cycles in the DAG using DFS.
func (d *DAG) detectCycles() (bool, []string) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycle []string

	for nodeID := range d.Nodes {
		if !visited[nodeID] && d.hasCycle(nodeID, visited, recStack, &cycle) {
			return true, cycle
		}
	}
	return false, nil
}

// DFS to detect cycles in the DAG.
func (d *DAG) hasCycle(nodeID string, visited, recStack map[string]bool, cycle *[]string) bool {
	visited[nodeID], recStack[nodeID] = true, true

	for _, neighbor := range d.Edges[nodeID] {
		if !visited[neighbor] {
			if d.hasCycle(neighbor, visited, recStack, cycle) {
				*cycle = append(*cycle, nodeID)
				return true
			}
		} else if recStack[neighbor] {
			*cycle = append(*cycle, neighbor, nodeID)
			return true
		}
	}
	recStack[nodeID] = false
	return false
}
