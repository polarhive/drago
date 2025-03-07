# drago

> workflow engine POC

wip: see [todo](TODO.md) and [reading](READING.md)

## Example Workflow

```json
[
  {
    "id": "sensor-ingest",
    "type": "trigger",
    "dependencies": []
  },
  {
    "id": "validate-data",
    "type": "compute",
    "dependencies": ["sensor-ingest"]
  },
  {
    "id": "store-db",
    "type": "api",
    "dependencies": ["validate-data"]
  }
]
```

### Load testing

```sh
go build -o loadtest cmd/loadtest/main.go
./loadtest -requests 1000 -concurrency 50 -url http://localhost:8080/trigger
```