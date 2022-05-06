# uGO Playground

uGO Playground is a single page web application to create playground for
[uGO](https://github.com/ozanh/ugo) script language. Playground is built for
WebAssembly.

## Project setup

Install followings:

- go v1.17
- node v14
- npm
- yarn

```sh
go install github.com/ozanh/ugo@latest
```

It is recommended to install `Vue CLI` npm packages globally for
development purposes. See detailed `Vue CLI` installation instructions
[here](https://cli.vuejs.org/guide/installation.html).

```sh
yarn global add @vue/cli
```

Install all node dependencies with the following:

```sh
yarn install
```

Use `vue ui` command to access to awesome Vue GUI to serve/build/test instantly.

### Compiles and hot-reloads for development

```sh
make development
yarn run serve
```

### Compiles and minifies for production

```sh
make production
yarn run build
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
yarn run lint
```

### Test Go and JS

```sh
make test
yarn run test
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
