ROOT_DIR       := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SHELL          := $(shell which bash)
ARGS            = $(filter-out $@,$(MAKECMDGOALS))

.SILENT: ;               # no need for @
.ONESHELL: ;             # recipes execute in same shell
.NOTPARALLEL: ;          # wait for this target to finish
.EXPORT_ALL_VARIABLES: ; # send all vars to shell
default: build;             # default target
Makefile: ;              # skip prerequisite discovery

build:
# go version go1.11 linux/amd64
	go build -i  -o bin/releaser src/main.go

run:
	bin/releaser

%:
	@:
