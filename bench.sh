# build bins
go build
go build -o loadtest cmd/loadtest/main.go

# run ./drago
./loadtest -requests 1000 -concurrency 50 -url http://localhost:8080/trigger
