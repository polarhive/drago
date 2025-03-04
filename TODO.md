# todo

- [x] Zap for logging
- [ ] Ingest, validate workflow from JSON
	- [x] Create a JSON file that represents the workflow nodes and their dependencies.
	- [ ] Modify the code to load nodes from this JSON file instead of the hardcoded sample.
	- [ ] Add validation steps:
		- [ ] Check that all dependencies exist in the node list.
		- [ ] Check that there are no cycles in the graph.
	- [ ] If validation passes, proceed to run the simulation; otherwise, output errors.

### Think about how to structure the project??

```
drago/
├── main.go
├── dag/
│   ├── dag.go       # structure and methods
│   └── validate.go  # how to validate
├── workflow/
│   └── workflow.go  # execution logic
└── nodes/
    └── nodes.go     # definitions and handlers
```
