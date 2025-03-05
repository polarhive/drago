# todo

- [x] Zap for logging
- [x] Step mode for debugging
- [x] Ingest, validate workflow from JSON
	- [x] Create a JSON file that represents the workflow nodes and their dependencies.
	- [x] Modify the code to load nodes from this JSON file instead of the hardcoded sample.
	- [x] Add validation steps:
		- [x] Check that all dependencies exist in the node list.
		- [x] Check that there are no cycles in the graph.
	- [x] If validation passes, proceed to run the simulation; otherwise, output errors.
- [x] Add tests

## Think about how to structure the project??

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

## Performance?

- memoization?
- metrics?
- benchmarks

## Caching, memory? Write to mock db

- [x] KV.go?

- [ ] JSON configureable rules 

## Usecases

### Max temp

- a sample workflow that ingests max temp from sensors
	- set kv 
	- more sensors ingest data
	- get kv
	- rule check > max 
	- set or skip
	- quit

### Chaos Test

### Ride/Cab booking example

- Spawn multiple producers/consumers
- Race