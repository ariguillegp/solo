.PHONY: validate deploy install

validate:
	gofmt -w ./cmd ./internal
	golangci-lint run --config=.golangci.yml ./...
	go test ./...

deploy:
	mkdir -p ~/.local/bin
	go build -o ~/.local/bin/rv ./cmd/rv

install:
	mkdir -p ~/Projects
	mkdir -p ~/.rivet/worktrees
	$(MAKE) deploy
