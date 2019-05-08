# Custom Makefile
# Inspired by https://gist.github.com/subfuzion/0bd969d08fe0d8b5cc4b23c795854a13

TARGET := $(shell echo $${PWD\#\#*/})
VERSION := $(shell cat VERSION)
IMPORTPATH := "github.com/catay"
CURRENT_COMMIT := $(shell git log -n 1 --format="%H")
RELEASE_COMMIT := $(shell git log -n 1 --format="%H" $(VERSION))

# Inject version info into the targets
LDFLAGS=-ldflags "-X $(IMPORTPATH)/$(TARGET)/version.Version=$(VERSION)"

CURRENT_LDFLAGS=-ldflags "-X $(IMPORTPATH)/$(TARGET)/version.Commit=$(CURRENT_COMMIT)"
RELEASE_LDFLAGS=-ldflags "-X $(IMPORTPATH)/$(TARGET)/version.Version=$(VERSION) -X $(IMPORTPATH)/$(TARGET)/version.Commit=$(RELEASE_COMMIT)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY:        build dep clean $(TARGET)

build: dep $(TARGET)

dep:
	@dep ensure

release: dep $(SRC)
 ifeq ($(CURRENT_COMMIT), $(RELEASE_COMMIT))
	@go build $(RELEASE_LDFLAGS) -o $(TARGET)
 else
	@echo "Current and tagged version commit hash don't match. Do nothing !!"
 endif

$(TARGET): dep $(SRC)
	@go build $(CURRENT_LDFLAGS) -o $(TARGET)

clean:
	@rm -f $(TARGET)

