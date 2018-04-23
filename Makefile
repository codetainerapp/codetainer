NAME="codetainer"
DESCRIPTION=""
WEBSITE="http://codetainer.org"

SHA := $(shell git rev-parse --short HEAD)
VERSION=$(shell cat $(NAME).go | grep "Version =" | sed 's/Version\ \=//' | sed 's/"//g' | tr -d '[[:space:]]')
CWD=$(shell pwd)

GITHUB_USER=recruit2class

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
UNAME := $(shell uname -s)

ifeq ($(UNAME),Darwin)
	ECHO=echo
else
	ECHO=/bin/echo -e
endif

all: install_deps bindata
	@mkdir -p bin/
	@$(ECHO) "$(OK_COLOR)==> Building $(NAME) $(VERSION) ($(SHA)) $(NO_COLOR)"
	@go build -o bin/$(NAME) -ldflags "-w -s -X main.Build=$(SHA)" cmd/*.go
	@mkdir -p bin/util/
	@$(ECHO) "$(OK_COLOR)==> Building utils $(NO_COLOR)"
	@go build -o bin/util/files cmd/util/files.go
	@chmod +x bin/$(NAME)
	@chmod +x bin/util/*
	@$(ECHO) "$(OK_COLOR)==> Done building$(NO_COLOR)"

updatedeps:
	@$(ECHO) "$(OK_COLOR)==> Updating all dependencies$(NO_COLOR)"
	@go get -d -v ./...
	@echo $(DEPS) | xargs -n1 go get -d

bindata:
	@$(ECHO) "$(OK_COLOR)==> Embedding assets$(NO_COLOR)"
	@$(GOPATH)/bin/go-bindata -pkg="codetainer" web/...

test:
	@$(ECHO) "$(OK_COLOR)==> Testing $(NAME)...$(NO_COLOR)"
	godep go test ./... -v

clean:
	@$(ECHO) "$(OK_COLOR)==> Cleaning$(NO_COLOR)"
	@rm -rf .Version
	@rm -rf bindata.go
	@rm -rf bin/
	@rm -rf pkg/
	@rm -rf release/
	@rm -rf package/

setup:
	@$(ECHO) "$(OK_COLOR)==> Building packages $(NAME)$(NO_COLOR)"
	@echo $(VERSION) > .Version
	@mkdir -p pkg/
	@mkdir -p package/root/usr/bin
	@cp -R pkg/linux-amd64/$(NAME) package/root/usr/bin
	@mkdir -p release/

install: clean all
	@mkdir -p $(GOPATH)/bin/
	@cp -r ./bin/codetainer $(GOPATH)/bin
	@cp -r ./bin/util/* $(GOPATH)/bin/

install_deps:
	@go get github.com/jteeuwen/go-bindata/...
	@go get github.com/elazarl/go-bindata-assetfs/...

uninstall: clean

docs: 
	mkdir -p doc/
	@go get -u github.com/go-swagger/go-swagger/cmd/swagger
	$(GOPATH)/bin/swagger generate spec > doc/swagger.json

.PHONY: all clean deps
