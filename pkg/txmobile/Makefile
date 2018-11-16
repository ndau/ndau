ANDROID_HOME ?= $(HOME)/Library/Android/sdk/

.PHONY: default all test build clean sources ios android

default: all

all: generate test build ios android txmobile.zip

generate:
	go generate

test:
	go test

build: ios android

ios: Txmobile.framework

sources: *.go

Txmobile.framework: sources
	gomobile bind -target ios -v

txmobile.zip: Txmobile.framework
	rm -f txmobile.zip
	zip -ur txmobile Txmobile.framework/

android: txmobile.aar

txmobile.aar: sources
	gomobile bind -target android -v

clean:
	rm -rf Txmobile.framework
	rm -f txmobile.aar
	rm -f txmobile-sources.jar
