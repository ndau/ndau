ANDROID_HOME ?= $(HOME)/Library/Android/sdk/

.PHONY: default all test build clean sources ios android

default: all

all: generate test build ios android keyaddr.zip

generate:
	go generate

test:
	go test

build: ios android

ios: Keyaddr.framework

sources: *.go

Keyaddr.framework: sources
	gomobile bind -target ios -v

keyaddr.zip: Keyaddr.framework
	rm -f keyaddr.zip
	zip -ur keyaddr Keyaddr.framework/

android: keyaddr.aar

keyaddr.aar: sources
	gomobile bind -target android -v

clean:
	rm -rf Keyaddr.framework
	rm -f keyaddr.aar
	rm -f keyaddr-sources.jar
