BINARY := bin/file-converter
IMAGE  := file-converter:dev
PORT   := 8080

.PHONY: build run test fmt fmt-check vet tidy check docker-build docker-run clean

build:
	CGO_ENABLED=1 go build -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

fmt:
	gofmt -w .

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Not gofmt'd:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	go vet ./...

tidy:
	go mod tidy

check: fmt-check vet test build

docker-build:
	docker build -t $(IMAGE) .

docker-run: docker-build
	docker run --rm -p $(PORT):$(PORT) $(IMAGE)

clean:
	rm -rf bin
