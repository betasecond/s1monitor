# Makefile for s1monitor

# Binary name
BINARY_NAME=s1monitor

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
MAIN_PATH=./cmd/s1monitor

# Directories
BIN_DIR=bin

# Command options
BUILD_FLAGS=-v

# Make targets
.PHONY: all build clean test run tidy

all: clean build

build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

build-all: clean
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME)_windows_amd64.exe $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME)_linux_amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME)_darwin_amd64 $(MAIN_PATH)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BIN_DIR)

test:
	$(GOTEST) -v ./...

run:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	./$(BINARY_NAME)

daemon:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	./$(BINARY_NAME) -d

tidy:
	$(GOMOD) tidy
