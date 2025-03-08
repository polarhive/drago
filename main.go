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
	workers := flag.Int("workers", 2, "Number of parallel workers")
	apiPort := flag.Int("port", 8080, "API server port")
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

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Workflow trigger endpoint
	router.POST("/trigger", func(c *gin.Context) {
		start := time.Now()

		wf := workflow.NewWorkflow(d, *workers, *stepMode, logger)

		// Store input data
		var input map[string]interface{}
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		// Set input data in workflow storage
		for k, v := range input {
			wf.Set(k, v)
		}

		go func() {
			defer func() {
				logger.Info("Workflow completed",
					zap.Duration("duration", time.Since(start)),
				)
			}()
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
