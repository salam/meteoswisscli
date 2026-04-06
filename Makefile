MODULE  := $(shell head -1 go.mod | awk '{print $$2}')
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w

.PHONY: build clean all

build:
	go build -ldflags "$(LDFLAGS) -X '$(MODULE)/internal/meteoswiss/cmd.version=$(VERSION)'" -o meteoswiss ./cmd/meteoswiss
	go build -ldflags "$(LDFLAGS) -X '$(MODULE)/internal/whiterisk/cmd.version=$(VERSION)'" -o whiterisk ./cmd/whiterisk

clean:
	rm -f meteoswiss whiterisk
	rm -rf dist

all: clean
	@for app in meteoswiss whiterisk; do \
		for os_arch in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64; do \
			GOOS=$${os_arch%/*} GOARCH=$${os_arch#*/} \
			go build -ldflags "$(LDFLAGS) -X '$(MODULE)/internal/$${app}/cmd.version=$(VERSION)'" \
				-o dist/$${app}-$${os_arch%/*}-$${os_arch#*/}$$([ "$${os_arch%/*}" = windows ] && echo .exe) \
				./cmd/$${app}; \
		done \
	done
