- [x] Zap for logging
- [ ] Ingest workflow from JSON
	  - [ ] Create a JSON file that represents the workflow nodes and their dependencies.
	  - [ ] Modify the code to load nodes from this JSON file instead of the hardcoded sample.
	  - [ ] Add validation steps:
		  - [ ] a. Check that all dependencies exist in the node list.
		  - [ ] b. Check that there are no cycles in the graph.
	 - [ ] If validation passes, proceed to run the simulation; otherwise, output errors.

Think about how to structure the JSON??