APP_NAME := rasgui
BUILD_DIR := build

.PHONY: fmt test e2e tidy run build build-windows build-linux clean

fmt:
	gofmt -w cmd internal

test:
	go test ./...

e2e:
	npm.cmd run test:e2e

tidy:
	go mod tidy

run:
	go run ./cmd/rasgui

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/rasgui

build-windows:
	go build -o $(BUILD_DIR)/$(APP_NAME).exe ./cmd/rasgui

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/rasgui

clean:
	- powershell -Command "if (Test-Path '$(BUILD_DIR)') { Remove-Item -Recurse -Force '$(BUILD_DIR)' }"
