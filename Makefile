
.PHONY: build
build: ## Run go build.
	go build example.go

.PHONY: run
run:
	go run example.go

