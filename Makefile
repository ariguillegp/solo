.PHONY: validate deploy install

validate:
	gofmt -w ./cmd ./internal
	golangci-lint run --config=.golangci.yml ./...
	go test ./...

deploy:
	install -d ~/.local/bin
	go build -o ~/.local/bin/solo ./cmd/solo

install:
	./scripts/install.sh
