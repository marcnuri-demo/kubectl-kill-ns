.PHONY: build
build:
	go build -ldflags "-s -w" -o kubectl-kill-ns ./cmd/main.go
