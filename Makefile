#
# Makefile
# @author Maksym Shkolnyi <maksymsh@wix.com>
#

.PHONY: clean build build-static dist deploy

export GO111MODULE=on

NAME := 3ff
MAJOR := $(shell cat VERSION | cut -d. -f 1)
MINOR := $(shell cat VERSION | cut -d. -f 2)
REVISION := $(shell cat VERSION | cut -d. -f 3)
BASEURL := https://tfresdif.s3.eu-central-1.amazonaws.com
OS=$(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH=amd64

LDFLAGS := \
  -X main.major=${MAJOR} \
  -X main.minor=${MINOR} \
  -X main.revision=${REVISION}

default: build

clean:
	@echo -e '\033[0;33mCleaning up...\033[0m'
	@rm dist/* -rf
	@rm bin/* -rf
	@echo -e '\033[0;32mDONE!\033[0m'

build:
	@echo -e '\033[0;33mBuilding...\033[0m'
	go build -v -o ./bin/$(NAME) -ldflags '${LDFLAGS}' .
	@echo -e '\033[0;32mDONE!\033[0m'

build-static:
	@echo -e '\033[0;33mBuilding...\033[0m'
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o ./bin/$(NAME) -ldflags '-s -w --extldflags "-static" ${LDFLAGS}' .
	@echo -e '\033[0;32mDONE!\033[0m'

dist:
	@# For linux 386 when building on linux amd64 you'll need 'libc6-dev-i386' package
	@echo -e '\033[0;33mBuilding dist\033[0m'

	@set -e ;\
#	Commented Windows platform			  "windows amd64 0 .exe "  \l
	for arch in   "linux   amd64 0      "  \
				  "darwin  amd64 0      "; \
	do \
		set -- $$arch ; \
		echo "******************* $$1_$$2 ********************" ;\
		distpath="./dist/$$1_$$2" ;\
		mkdir -p $$distpath ; \
		CGO_ENABLED=$$3 GOOS=$$1 GOARCH=$$2 go build -v -a -o $$distpath/$(NAME)$$4 -ldflags '-s -w --extldflags "-static" ${LDFLAGS}' . ;\
		pushd $$distpath ; md5sum $(NAME) | tee $(NAME).md5 ; popd ;\
		cp "README.md" "LICENSE" $$distpath ;\
	done
	@echo -e '\033[0;32mDONE!\033[0m'


