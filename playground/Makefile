
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS += --warn-undefined-variables

ENV_FILE=.env.local

UGO_PATH="$(shell go list -m -f {{.Dir}} github.com/ozanh/ugo)"
UGO_VERSION="$(shell go list -m -f {{.Version}} github.com/ozanh/ugo)"

GOOS=js
GOARCH=wasm
GOLDFLAGS=-s -w

BUILD_TIME=$(shell date -u +'%Y-%m-%d_%T')
JS_WASM_EXEC="$(shell go env GOROOT)/misc/wasm/go_js_wasm_exec"

VUE_CLI_SERVICE=./node_modules/.bin/vue-cli-service

.PHONY: all
all: build test-all

.PHONY: build
build: production
	$(VUE_CLI_SERVICE) build

.PHONY: serve
serve: development
	$(VUE_CLI_SERVICE) serve

.PHONY: test-all
test-all: lint test
	$(VUE_CLI_SERVICE) test:unit

.PHONY: lint
lint:
	$(VUE_CLI_SERVICE) lint --fix

.PHONY: install
install: ugo.wasm wasm_exec.js
	test -f ../LICENSE
	test -f "$(UGO_PATH)/LICENSE.golang"
	test -f "$(UGO_PATH)/LICENSE.tengo"
	printf "VUE_APP_WASM_FILE=ugo.wasm\n"\
	"VUE_APP_UGO_VERSION=$(UGO_VERSION)\n"\
	"VUE_APP_BUILD_TIME=$(BUILD_TIME)\n"\
	"VUE_APP_LICENSE=../LICENSE\n"\
	"VUE_APP_GO_LICENSE=$(UGO_PATH)/LICENSE.golang\n"\
	"VUE_APP_TENGO_LICENSE=$(UGO_PATH)/LICENSE.tengo\n" > $(ENV_FILE)

.PHONY: --append
--append:
	echo "VUE_APP_WASM_EXEC_FILE=wasm_exec.js" >> $(ENV_FILE)

wasm_exec.js: # copy wasm_exec.js to current working dir
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm_exec.js

ugo.wasm: # create ugo.wasm file in current working dir
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(GOLDFLAGS)" -o ./ugo.wasm \
    	github.com/ozanh/ugodev/playground/cmd/wasm

.PHONY: development
development: clean install --append
	mv $(ENV_FILE) .env.$@.local

.PHONY: production
production: distclean install
	mv wasm_exec.js "wasm_exec.$(shell md5sum wasm_exec.js | cut -c1-8).js"
	echo "VUE_APP_WASM_EXEC_FILE=$(wildcard wasm_exec.*.js)" >> $(ENV_FILE)
	mv $(ENV_FILE) .env.$@.local

.PHONY: test
test: clean install --append
	mv $(ENV_FILE) .env.$@.local
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -cover -exec="$(JS_WASM_EXEC)" \
        github.com/ozanh/ugo/...
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -cover -exec="$(JS_WASM_EXEC)" \
        github.com/ozanh/ugodev/playground/cmd/wasm

.PHONY: clean
clean:
	rm -rf ./.env.local ./.env.*.local ./wasm_exec.js ./wasm_exec.*.js ./ugo.wasm

.PHONY: distclean
distclean: clean
	rm -rf ./dist ./coverage ./node_modules/.cache
