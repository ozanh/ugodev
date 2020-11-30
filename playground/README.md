# uGO Playground

uGO Playground is a single page web application to create playground for
[uGO](https://github.com/ozanh/ugo) script language. Playground is built for
WebAssembly.

As of now (Go 1.15) Go's WebAssembly support is experimental so is uGO. Although
it is experimental, all uGO tests are passed with `GOOS=js GOARCH=wasm`
environment variables. Use following command to test and see
[Makefile](Makefile) script to check how each package is tested.

```sh
GOOS=js GOARCH=wasm go test -cover \
        -exec="$(go env GOROOT)/misc/wasm/go_js_wasm_exec" \
        github.com/ozanh/ugo/...
```

## Why WebAssembly

Thanks to Go's WebAssembly builds, scripts run in clients' web browser or in
nodejs which removes server communication and sandboxing requirements.
In addition, WebAssembly is the future of the web.

> *Go's WebAssembly binaries are big! But not that big*. Built wasm file size
is about 4.7MB and which can be served as gzipped. In the end, client loads about
1.3MB of data to run playground including all wasm, js, css and other assets.

## Project setup

Install followings:

- go v1.15
- node v12
- npm

```sh
go get -u github.com/ozanh/ugo
```

It is recommended to install `Vue CLI` npm packages globally for
development purposes. See detailed `Vue CLI` installation instructions
[here](https://cli.vuejs.org/guide/installation.html).

```sh
npm install -g @vue/cli @vue/cli-service-global
```

Install all node dependencies with the following:

```sh
npm install
```

Use `vue ui` command to access to awesome Vue GUI to serve/build/test instantly.

### Compiles and hot-reloads for development

```sh
npm run serve
```

### Compiles and minifies for production

```sh
npm run build
```

Built files are placed in `dist` directory.

There is a simple Go web server in the package at `cmd/server` directory to
access web application which can be run with the following command:

```sh
go run cmd/server/main.go
```

```sh
go run cmd/server/main.go -h
Usage
  -dir string
        file server root directory (default "dist")
  -listen string
        bind address:port (default ":9090")
```

### Lints and fixes files

```sh
npm run lint
```

### Test Go and JS

```sh
npm run test
```

### Customize configuration

See [Makefile](Makefile) file for testing and building.

See [vue.config.js](vue.config.js) file for Vue settings ([Configuration
Reference](https://cli.vuejs.org/config/)).

See [package.json](package.json) file for other settings.

## TODO

- [ ] Import `uGO` scripts as modules from local files, http(s) addresses and
  github gists.

## LICENSE

uGO Playground is licensed under the MIT License.

See [LICENSE](LICENSE) for the full license text.
