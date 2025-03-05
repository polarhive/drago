package dag

import (
	"testing"

	"go.uber.org/zap"
)

func TestMissingDependency(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d := NewDAG(logger)
	err := d.LoadFromJSON("testdata/invalid_workflow.json")
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	err = d.Validate()
	if err == nil {
		t.Fatal("Expected validation error for missing dependency, got nil")
	}

	expectedErr := "node process has missing dependency: non-existent-node"
	if err.Error() != expectedErr {
		t.Errorf("Expected error: %q, got: %q", expectedErr, err.Error())
	}
}

func TestCycleDetection(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d := NewDAG(logger)
	err := d.LoadFromJSON("testdata/cyclic_workflow.json")
	if err != nil {
		t.Fatalf("Failed to load workflow: %v", err)
	}

	err = d.Validate()
	if err == nil {
		t.Fatal("Expected validation error for cycle detection, got nil")
	}

	expectedErrPrefix := "cycle detected:"
	if len(err.Error()) <= len(expectedErrPrefix) || err.Error()[:len(expectedErrPrefix)] != expectedErrPrefix {
		t.Errorf("Expected error starting with %q, got: %q", expectedErrPrefix, err.Error())
	}
}
