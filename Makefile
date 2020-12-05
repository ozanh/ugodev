SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS += --warn-undefined-variables

.PHONY: cover
cover:
	# wasm builds are protected by build flags
	go test -coverprofile cover.out ./...
	cd playground && COVER_FILE=../pgcover.out $(MAKE) test
	# Assume mode is set for both files and concat them with mode line
	tail -n +2 pgcover.out >> cover.out
