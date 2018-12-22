# Custom Makefile
# Inspired by https://gist.github.com/subfuzion/0bd969d08fe0d8b5cc4b23c795854a13

TARGET := $(shell echo $${PWD\#\#*/})
TARGET1 := "rrst"
TARGET2 := "rrstd"
VERSION := $(shell cat VERSION)
IMPORTPATH := "github.com/catay"

# Inject version info into the targets
LDFLAGS=-ldflags "-X $(IMPORTPATH)/$(TARGET)/version.Version=$(VERSION)"

.PHONY:        build dep clean $(TARGET1) $(TARGET2)

build: dep $(TARGET1) $(TARGET2)

dep:
	@dep ensure

$(TARGET1):
	@go build $(LDFLAGS) -o $(TARGET1) cmd/$(TARGET1)/*go

$(TARGET2):
	@go build $(LDFLAGS) -o $(TARGET2) cmd/$(TARGET2)/*.go

clean:
	@rm -f $(TARGET1) $(TARGET2)
