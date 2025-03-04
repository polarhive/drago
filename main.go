package main

import (
	"flag"
	"os"

	"github.com/polarhive/drago/dag"
	"github.com/polarhive/drago/workflow"
	"go.uber.org/zap"
)

func main() {
	workers := flag.Int("workers", 2, "Number of parallel workers")
	stepMode := flag.Bool("step", false, "Step mode")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize DAG and load workflow
	d := dag.NewDAG(logger)
	if err := d.LoadFromJSON("workflow.json"); err != nil {
		logger.Fatal("Failed to load workflow", zap.Error(err))
	}

	// Validate workflow
	if err := d.Validate(); err != nil {
		logger.Error("Workflow validation failed",
			zap.Error(err),
			zap.String("action", "validation failed, aborting execution"),
		)
		os.Exit(1)
	}

	// Create and run workflow
	wf := workflow.NewWorkflow(d, *workers, *stepMode, logger)
	wf.Run()
}
