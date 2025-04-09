
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS += --warn-undefined-variables

ENV_FILE=.env.local

GO_VERSION=$(shell go env GOVERSION)
UGO_PATH=$(shell go list -m -f {{.Dir}} github.com/ozanh/ugo)
UGO_VERSION=$(shell go list -m -f {{.Version}} github.com/ozanh/ugo)

GOOS=js
GOARCH=wasm
GOLDFLAGS=-s -w

BUILD_TIME=$(shell date -u +'%Y-%m-%d_%T')
JS_WASM_EXEC="$(shell go env GOROOT)/misc/wasm/go_js_wasm_exec"

.PHONY: all
all: build test-all

.PHONY: build
build: production
	yarn build

.PHONY: serve
serve: development
	yarn dev

.PHONY: test-all
test-all: lint test

.PHONY: lint
lint:

.PHONY: install
install: ugo.wasm wasm_exec.js
	test -f ../LICENSE
	test -f "$(UGO_PATH)/LICENSE.golang"
	test -f "$(UGO_PATH)/LICENSE.tengo"
	printf "VITE_GO_VERSION=$(GO_VERSION)\n"\
	"VITE_UGO_VERSION=$(UGO_VERSION)\n"\
	"VITE_BUILD_TIME=$(BUILD_TIME)\n"\
	"VITE_GO_LICENSE=\"$(UGO_PATH)/LICENSE.golang\"\n"\
	"VITE_TENGO_LICENSE=\"$(UGO_PATH)/LICENSE.tengo\"\n" > $(ENV_FILE)

wasm_exec.js: # copy wasm_exec.js to current working dir
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm_exec.js

ugo.wasm: # create ugo.wasm file in current working dir
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(GOLDFLAGS)" -o ./ugo.wasm \
    	github.com/ozanh/ugodev/playground/cmd/wasm

.PHONY: development
development: clean install
	mv $(ENV_FILE) .env.$@.local

.PHONY: production
production: distclean install
	mv $(ENV_FILE) .env.$@.local

.PHONY: test
test: clean install
	mv $(ENV_FILE) .env.$@.local
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -cover -exec="$(JS_WASM_EXEC)" \
        github.com/ozanh/ugo/...
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -v -cover -exec="$(JS_WASM_EXEC)" \
        github.com/ozanh/ugodev/playground/cmd/wasm

.PHONY: clean
clean:
	rm -rf ./.env.local ./.env.*.local ./wasm_exec.js ./ugo.wasm ./ThirdPartyNotices.txt

.PHONY: distclean
distclean: clean
	rm -rf ./dist ./coverage ./node_modules/.cache
