package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polarhive/drago/dag"
	"github.com/polarhive/drago/workflow"
	"go.uber.org/zap"
)

func main() {
	// Define command-line flags for configuring workers, API port, and workflow file
	workers := flag.Int("workers", 2, "Number of parallel workers")
	apiPort := flag.Int("port", 8080, "API server port")
	workflowFile := flag.String("workflow", "workflow.json", "Workflow file")
	stepMode := flag.Bool("step", false, "Step mode")

	flag.Parse()

	// Initialize a logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize the DAG and load the workflow from JSON
	d := dag.NewDAG(logger)
	if err := d.LoadFromJSON(*workflowFile); err != nil {
		logger.Fatal("Failed to load workflow", zap.Error(err))
	}

	// Validate workflow structure
	if err := d.Validate(); err != nil {
		logger.Error("Workflow validation failed",
			zap.Error(err),
			zap.String("action", "validation failed, aborting execution"),
		)
		os.Exit(1)
	}

	// Set up the HTTP API using Gin
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API endpoint to trigger the workflow
	router.POST("/trigger", func(c *gin.Context) {
		start := time.Now()

		// Create a new workflow instance
		wf := workflow.NewWorkflow(d, *workers, *stepMode, logger)

		// Parse input data from request body
		var input map[string]interface{}
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		// Store input data in workflow storage
		for k, v := range input {
			wf.Set(k, v)
		}

		// Run the workflow asynchronously
		go func() {
			duration := time.Since(start)
			defer logger.Info("Workflow completed", zap.Duration("duration", duration))
			wf.Run()
		}()

		c.JSON(http.StatusAccepted, gin.H{
			"message":    "workflow started",
			"started_at": start.Format(time.RFC3339),
		})
	})

	logger.Info("Starting API server", zap.Int("port", *apiPort))
	router.Run(fmt.Sprintf(":%d", *apiPort))
}
