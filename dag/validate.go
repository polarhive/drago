package dag

import "fmt"

func (d *DAG) Validate() error {
	d.Mu.RLock()
	defer d.Mu.RUnlock()

	// Validate dependencies exist
	for _, node := range d.Nodes {
		for _, dep := range node.Dependencies {
			if _, exists := d.Nodes[dep]; !exists {
				return fmt.Errorf("node %s has missing dependency: %s", node.ID, dep)
			}
		}
	}

	// Validate no cycles
	if cyclic, cycle := d.isCyclic(); cyclic {
		return fmt.Errorf("cycle detected: %v", cycle)
	}

	// Validate node types
	validTypes := map[string]bool{
		"trigger":  true,
		"compute":  true,
		"decision": true,
		"api":      true,
		"action":   true,
	}
	for _, node := range d.Nodes {
		if !validTypes[node.Type] {
			return fmt.Errorf("invalid node type %s for node %s", node.Type, node.ID)
		}
	}

	return nil
}

func (d *DAG) isCyclic() (bool, []string) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycle []string

	for nodeID := range d.Nodes {
		if !visited[nodeID] {
			if d.detectCycle(nodeID, visited, recStack, &cycle) {
				return true, cycle
			}
		}
	}
	return false, nil
}

func (d *DAG) detectCycle(nodeID string, visited, recStack map[string]bool, cycle *[]string) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, neighbor := range d.Edges[nodeID] {
		if !visited[neighbor] {
			if d.detectCycle(neighbor, visited, recStack, cycle) {
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