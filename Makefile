TARGET := url-shortener
CGO_ENABLED := 1
GOGC := 100
GOOS := linux
GOARCH := amd64
CMD_API_DIR := cmd/url-shortener
BIN_DIR := bin

GO_ENV_VAR := GOGC=$(GOGC) CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH)
FLAGS := -ldflags="-s -w"

all: $(TARGET)
.PHONY: all

build: $(TARGET)
.PHONY: build

$(TARGET):
	$(GO_ENV_VAR) go build $(FLAGS) -o $(BIN_DIR)/$(TARGET) $(CMD_API_DIR)/main.go

launch: clean all
	./$(BIN_DIR)/$(TARGET)
.PHONY: launch

run:
	go run $(CMD_API_DIR)/main.go
.PHONY: run

unit-test:
	go test -cover -race -count=1 -v ./internal/... -coverprofile=coverage.out
	go tool cover -func=coverage.out
.PHONY: unit-test

test-html:
	go tool cover -html=coverage.out
.PHONY: cover-html

mock:
	go generate ./...
.PHONY: mock

swag-init:
	swag init -g internal/app/app.go
.PHONY: swag-init

build:
	docker compose build
.PHONY: build

up:
	docker compose up -d
.PHONY: up

down:
	docker compose down
.PHONY: down

clean:
	-rm $(BIN_DIR)/*
.PHONY: clean
