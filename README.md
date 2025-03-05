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