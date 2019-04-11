NAME		 := mget
SHORT_NAME   := mget
VERSION      := $(shell git describe --always --tags | sed 's/^v//')
PKGS         := $(shell go list ./... | grep -v /vendor/)
PKG_NAME     := $(SHORT_NAME)-$(VERSION)

all:
	@echo "Nothing to build"
	@echo "possible targets are: build, format, tarball, clean"

build: mget
	@echo ">> building binaries"
	@go build -i \
	    -o $(NAME) \
		-ldflags "-X main.Version=$(VERSION) -X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S'` -X main.Githash=`git rev-parse HEAD`" \
		./cmd/*.go

format:
	@echo ">> formatting code"
	@go fmt $(PKGS)

mget:
	@echo ">> fetching mget"

tarball: build_root
	@echo ">> packaging binaries and scripts"
	@mkdir -p dist
	@cd .build; zip -r ../dist/$(PKG_NAME).zip  $(PKG_NAME)
	@echo "Builded dist/$(PKG_NAME).zip"

build_root: build_linux_amd64 build_linux_i686 build_mac_amd64
	@echo ">> making build root"
	@rm -rf .build
	@mkdir -p .build
	@install -m 0755 -d .build/$(PKG_NAME)
	@install -m 0755 .bin/$(NAME)_linux32 .build/$(PKG_NAME)/$(NAME)_linux32
	@install -m 0755 .bin/$(NAME)_linux64 .build/$(PKG_NAME)/$(NAME)_linux64
	@install -m 0755 .bin/$(NAME)_mac64 .build/$(PKG_NAME)/$(NAME)_mac64
	@install -m 0755 ./script/install.sh .build/$(PKG_NAME)/install.sh
	@install -m 0755 ./README.md .build/$(PKG_NAME)/README.md
	@echo "Builded .build/$(PKG_NAME)"

build_linux_amd64: mget
	@echo ">> building linux amd64 binaries"
	@mkdir -p .bin
	@GOOS=linux GOARCH=amd64 \
		go build \
			-o .bin/$(NAME)_linux64 \
	     	-ldflags "-X main.Version=$(VERSION) -X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S'` -X main.Githash=`git rev-parse HEAD`" \
			./cmd/*.go
	@echo "Builded .bin/$(NAME)_linux64"

build_linux_i686: mget
	@echo ">> building linux i686 binaries"
	@mkdir -p .bin
	@GOOS=linux GOARCH=386 \
		go build \
			-o .bin/$(NAME)_linux32 \
	     	-ldflags "-X main.Version=$(VERSION) -X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S'` -X main.Githash=`git rev-parse HEAD`" \
			./cmd/*.go
	@echo "Builded .bin/$(NAME)_linux32"

build_mac_amd64: mget
	@echo ">> building mac amd64 binaries"
	@mkdir -p .bin
	@GOOS=darwin GOARCH=amd64 \
		go build \
			-o .bin/$(NAME)_mac64 \
	     	-ldflags "-X main.Version=$(VERSION) -X main.Buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S'` -X main.Githash=`git rev-parse HEAD`" \
			./cmd/*.go
	@echo "Builded .bin/$(NAME)_mac64"

clean:
	@rm -rf .build
	@rm -rf .bin
	@rm -rf $(NAME)

.PHONY: format build mget tarball build_root build_linux_amd64 build_linux_i686 build_mac_amd64 clean
