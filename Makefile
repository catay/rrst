# Custom Makefile
# Inspired by https://gist.github.com/subfuzion/0bd969d08fe0d8b5cc4b23c795854a13

TARGET := $(shell echo $${PWD\#\#*/})
VERSION := $(shell cat VERSION)
IMPORTPATH := "github.com/catay"

# Inject version info into the targets
LDFLAGS=-ldflags "-X $(IMPORTPATH)/$(TARGET)/version.Version=$(VERSION)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY:        build dep clean $(TARGET)

build: dep $(TARGET)

dep:
	@dep ensure

$(TARGET): dep $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

clean:
	@rm -f $(TARGET)

