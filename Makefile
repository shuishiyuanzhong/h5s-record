PROJECT_NAME=h5s-record

.PHONY: all dep lint vet test test-coverage build clean

all: build

dep: ## Get the dependencies
	@go mod tidy

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic ${PKG_LIST}
	@cat cover.out >> coverage.txt

build: dep ## Build the binary file
	@go build -o ./bin/$(PROJECT_NAME)

linux: dep ## Build the binary file
	@GOOS=linux GOARCH=amd64 go build  -o ./bin/$(PROJECT_NAME)

amd: dep ## Build the binary file
	@GOOS=linux GOARCH=amd64 go build  -o ./bin/$(PROJECT_NAME)-amd64

arm: dep ## Build the binary file
	@GOOS=linux GOARCH=arm64 go build  -o ./bin/$(PROJECT_NAME)-arm64

windows: dep ## Build the binary file
	@GOOS=windows go build  -o ./bin/$(PROJECT_NAME)-windows.exe

mac-arm64: dep ## Build the binary file
	@GOOS=darwin GOARCH=arm64 go build  -o ./bin/$(PROJECT_NAME)-darwin-arm64

mac-amd64: dep ## Build the binary file
	@GOOS=darwin GOARCH=amd64 go build  -o ./bin/$(PROJECT_NAME)-darwin-amd64

build-all-platform: dep ## Build the binary file
	@GOOS=linux GOARCH=amd64 go build  -o ./bin/$(PROJECT_NAME)-amd64
	@GOOS=linux GOARCH=arm64 go build  -o ./bin/$(PROJECT_NAME)-arm64
	@GOOS=windows go build  -o ./bin/$(PROJECT_NAME)-windows.exe
	@GOOS=darwin GOARCH=amd64 go build  -o ./bin/$(PROJECT_NAME)-darwin-amd64
	@GOOS=darwin GOARCH=arm64 go build  -o ./bin/$(PROJECT_NAME)-darwin-arm64

#run: # Run Develop server
#	@go run  start -f etc/app.toml

clean: ## Remove previous build
	@rm -rf ./bin

#push: # push git to multi repo
#	@git push -u gitee
#	@git push -u origin

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'