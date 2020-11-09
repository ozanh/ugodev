#!/bin/bash
set -eu
set -o pipefail

MODE="$1"
if [ -z "$MODE" ]; then
    MODE="production"    
fi

ENV_FILE=".env.$MODE.local"
UGO_PATH="$(go env GOPATH)/src/github.com/ozanh/ugo"
JS_WASM_EXEC="$(go env GOROOT)/misc/wasm/go_js_wasm_exec"

if [ "$MODE" = "test" ]; then
    pushd "cmd/wasm"
    GOOS=js GOARCH=wasm go test -cover -exec="$JS_WASM_EXEC" .
    popd
fi

pushd "$UGO_PATH"
if [ "$MODE" = "test" ]; then
    # find all packages in uGO and run tests except for cmd/ugo which is a terminal app.
    find . -name '*.go' -printf '%h\n' | sort -u | grep -v "cmd/ugo" | \
        GOOS=js GOARCH=wasm xargs -n1 -P1 go test -cover -exec="$JS_WASM_EXEC"
fi
popd

mkdir -p dist/static

# build & copy files and write env. vars to .env.[MODE].local file for Vue app
rm -f ./.env.*.local
rm -f ./wasm_exec.*.js ./wasm_exec.js 
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm_exec.js

EXEC_FILE="wasm_exec.js"
if [ "$MODE" = "production" ]; then
    EXEC_FILE="wasm_exec.$(md5sum wasm_exec.js | cut -c1-8).js"
    mv ./wasm_exec.js "$EXEC_FILE"
fi
echo "VUE_APP_WASM_EXEC_FILE=$EXEC_FILE" > "$ENV_FILE"

# get build time in UTC
BUILD_TIME=$(date -u +'%Y-%m-%d_%T')
rm -f ./ugo.*.wasm ./ugo.wasm

pushd cmd/wasm
UGO_VERSION="$(go list -m -f {{.Version}} github.com/ozanh/ugo)"
popd

# create ugo.wasm file in current working dir
GOOS=js GOARCH=wasm go build -o ./ugo.wasm \
    github.com/ozanh/ugodev/playground/cmd/wasm
    

WASM_FILE="ugo.wasm"
if [ "$MODE" = "production" ]; then
    WASM_FILE="ugo.$(md5sum ugo.wasm | cut -c1-8).wasm"
    mv ./ugo.wasm "$WASM_FILE"
fi
cat >> "$ENV_FILE" << EOF
VUE_APP_WASM_FILE=$WASM_FILE
VUE_APP_UGO_VERSION=$UGO_VERSION
VUE_APP_BUILD_TIME=$BUILD_TIME
VUE_APP_LICENSE=../LICENSE
VUE_APP_GO_LICENSE=$UGO_PATH/LICENSE.golang
VUE_APP_TENGO_LICENSE=$UGO_PATH/LICENSE.tengo
EOF
