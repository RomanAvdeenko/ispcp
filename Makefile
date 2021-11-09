.PHONY: build test
build: 
	go build -v -o pinger ./cmd/pinger/main.go
	sudo setcap cap_net_raw+ep ./pinger
build1: 
	go build -v -o pi ./cmd/test/main.go
	sudo setcap cap_net_raw+ep ./pi


test:
	go test -v -race -timeout 10s ./...

.DEFAULT_GOAL :=build 