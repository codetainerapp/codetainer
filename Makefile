NAME="codetainer"
DESCRIPTION=""
WEBSITE="http://codetainer.org"

VERSION=$(shell cat $(NAME).go | grep "Version =" | sed 's/Version\ \=//' | sed 's/"//g' | tr -d '[[:space:]]')
CWD=$(shell pwd)

GITHUB_USER=codetainerapp
CCOS=linux
CCARCH=386 amd64
CCOUTPUT="pkg/{{.OS}}-{{.Arch}}/$(NAME)"

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

all: bindata
	@mkdir -p bin/
	@$(ECHO) "$(OK_COLOR)==> Building $(NAME) $(VERSION) $(NO_COLOR)"
	@godep go build -o bin/$(NAME) cmd/*.go
	@mkdir -p bin/util/
	@$(ECHO) "$(OK_COLOR)==> Building utils $(NO_COLOR)"
	@godep go build -o bin/util/files cmd/util/files.go
	@chmod +x bin/$(NAME)
	@chmod +x bin/util/*
	@$(ECHO) "$(OK_COLOR)==> Done building$(NO_COLOR)"

build: bindata all
	@$(ECHO) "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@godep get

updatedeps:
	@$(ECHO) "$(OK_COLOR)==> Updating all dependencies$(NO_COLOR)"
	@go get -d -v ./...
	@echo $(DEPS) | xargs -n1 go get -d 
	@godep update ...

bindata:
	@$(ECHO) "$(OK_COLOR)==> Embedding assets$(NO_COLOR)"
	@go-bindata -pkg="codetainer" web/...

test:
	@$(ECHO) "$(OK_COLOR)==> Testing $(NAME)...$(NO_COLOR)"
	godep go test ./... -v

goxBuild:
	@CGO_ENABLED=1 gox -os="$(CCOS)" -arch="$(CCARCH)" -build-toolchain

gox: 
	@$(ECHO) "$(OK_COLOR)==> Cross compiling $(NAME)$(NO_COLOR)"
	@mkdir -p Godeps/_workspace/src/github.com/$(GITHUB_USER)/$(NAME)
	@cp -R *.go Godeps/_workspace/src/github.com/$(GITHUB_USER)/$(NAME)
	@CGO_ENABLED=1 GOPATH=$(shell godep path) gox -ldflags "-s" -os="$(CCOS)" -arch="$(CCARCH)" -output=$(CCOUTPUT)
	@rm -rf Godeps/_workspace/src/github.com/$(GITHUB_USER)/$(NAME)

release: clean all test gox setup package
	@for os in $(CCOS); do \
		for arch in $(CCARCH); do \
			cd pkg/$$os-$$arch/; \
			tar -zcvf ../../release/$(NAME)-$$os-$$arch.tar.gz $(NAME)* > /dev/null 2>&1; \
			cd ../../; \
		done \
	done
	@$(ECHO) "$(OK_COLOR)==> Done cross compiling $(NAME)$(NO_COLOR)"

clean:
	@$(ECHO) "$(OK_COLOR)==> Cleaning$(NO_COLOR)"
	@rm -rf Godeps/_workspace/src/github.com/$(GITHUB_USER)/$(NAME)
	@rm -rf .Version
	@rm -rf bindata.go
	@rm -rf bin/
	@rm -rf pkg/
	@rm -rf release/
	@rm -rf package/

setup:
	@$(ECHO) "$(OK_COLOR)==> Building packages $(NAME)$(NO_COLOR)"
	@echo $(VERSION) > .Version
	@mkdir -p package/root/usr/bin
	@cp -R pkg/linux-amd64/$(NAME) package/root/usr/bin
	@mkdir -p release/

package: deb386 debamd64

debamd64:
	fpm -s dir -t deb -n $(NAME) -v $(VERSION) -p release/$(NAME)-amd64.deb \
		--deb-priority optional --category admin \
		--force \
		--deb-compression bzip2 \
		--url $(WEBSITE) \
		--description $(DESCRIPTION) \
		-m "codetainer <dev@codetainer.org>" \
		--vendor "Codetainer" -a amd64 \
		--exclude */**.gitkeep \
		package/root/=/

deb386:
	fpm -s dir -t deb -n $(NAME) -v $(VERSION) -p release/$(NAME)-386.deb \
		--deb-priority optional --category admin \
		--force \
		--deb-compression bzip2 \
		--url $(WEBSITE) \
		--description $(DESCRIPTION) \
		-m "codetainer <dev@codetainer.org>" \
		--vendor "Codetainer" -a 386 \
		--exclude */**.gitkeep \
		package/root/=/

install: clean all
  mkdir -p /opt/codetainer
  sudo cp -r ./bin/ /opt/codetainer/bin

uninstall: clean

tar: 

.PHONY: all clean deps
