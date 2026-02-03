.PHONY: validate deploy install

validate:
	gofmt -w ./cmd ./internal
	golangci-lint run --config=.golangci.yml ./...
	go test ./...

deploy:
	mkdir -p ~/.local/bin
	go build -o ~/.local/bin/solo ./cmd/solo

install:
	mkdir -p ~/Projects
	mkdir -p ~/.solo/worktrees
	$(MAKE) deploy
