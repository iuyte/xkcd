GO_BUILD_ENV := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
OUT_DIR=$(shell pwd)/bin
OUT_FILE=$(OUT_DIR)/log

all: clean
	mkdir -p $(OUT_DIR)
	$(GO_BUILD_ENV) go build -v -o $(OUT_FILE) .

clean:
	rm -rf $(OUT_DIR)
