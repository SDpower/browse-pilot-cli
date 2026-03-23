.PHONY: build test lint clean extension-build extension-lint

BINARY = bp
DIST_DIR = dist

build:
	@mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY) ./cmd/bp/

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f $(DIST_DIR)/$(BINARY)
	rm -rf $(DIST_DIR)/firefox $(DIST_DIR)/chrome $(DIST_DIR)/edge

extension-build:
	bash scripts/build-extensions.sh

extension-lint:
	npx eslint extension/

all: build extension-build
