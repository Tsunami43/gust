BINARY := gust
PKG := ./...

.PHONY: build test vet lint run install clean tidy

## build: compile the gust binary
build:
	go build -o $(BINARY) ./cmd/gust

## test: run the test suite
test:
	go test $(PKG)

## vet: run go vet
vet:
	go vet $(PKG)

## lint: run golangci-lint (if installed)
lint:
	golangci-lint run

## run: build and run gust
run: build
	./$(BINARY)

## install: install gust into GOBIN / GOPATH/bin
install:
	go install ./cmd/gust

## tidy: tidy the module
tidy:
	go mod tidy

## clean: remove build artefacts
clean:
	rm -f $(BINARY)
