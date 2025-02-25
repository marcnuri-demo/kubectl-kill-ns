.PHONY: build
build:
	go build -ldflags "-s -w" -o kubectl-kill-ns ./cmd/main.go

.PHONY: test
test:
	go test ./...

.PHONY: update-go-deps
update-go-deps:
	@echo ">> updating Go dependencies"
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get $$m; \
	done
	go mod tidy
ifneq (,$(wildcard vendor))
	go mod vendor
endif
