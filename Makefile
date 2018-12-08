PACKAGE_NAME := github.com/yusufsyaifudin/go-opentracing-example
PROJECT_DIR := $(PWD)

run: install-dep
	go run $(PROJECT_DIR)/main.go

install-dep:
	go get -v -u github.com/golang/dep/cmd/dep
	dep ensure -v
