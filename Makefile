.PHONY: build
build:
	go build

install: build
	which mfacli && cp -v mfacli `which mfacli` || go install
