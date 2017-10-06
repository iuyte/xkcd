GO_BUILD_ENV := CGO_ENABLED=0 GOOS=linux GOARCH=arm
OUT_DIR=$(shell pwd)/bin
OUT_FILE=$(OUT_DIR)/`basename $(PWD)`
DOCKER_ID=`docker build -q .`

all:
	mkdir -p $(OUT_DIR)
	$(GO_BUILD_ENV) go build -v -o $(OUT_FILE) .

clean:
	rm -rf $(OUT_DIR)

docker: all
	docker run $(DOCKER_ID)
