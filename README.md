# go-prettier

go-prettier is a distribution of [prettier][1], that can be built with Go. It does not actually reimplement core
functionality of prettier in Go (though does reimplement the CLI), instead packaging it with the lightweight JS
runtime [QuickJS][3], and executing with the pure Go Wasm runtime [wazero][2]. This means that `go install` or `go run`
can be used to execute it, with no need to rely on separate package managers such as pnpm, on any platform that Go
supports.

Prettier can format files like YAML and Markdown which are commonly used in Go projects, and integrates well
with IDEs like VSCode. This project is primarily designed for Go, or other non-JS projects, that would like to
still use prettier to format such non-logic files.

## Limitations

- External plugins are not supported. Currently, only the built-in plugins are included.
- Caching is not supported.
- Config must be JSON, YAML, or TOML. JS configs are not supported.
- Performance is worse for many files. The intent is to format a few yaml or markdown type files
  in a Go repository but not to replace formatting in a full NodeJS project. It is recommended to specify globs
  for the files that should be formatted rather than relying on auto-detection on a large directory.

## Installation

Precompiled binaries are available in the [releases](https://github.com/wasilibs/go-prettier/releases).
Alternatively, install the plugin you want using `go install`.

```bash
$ go install github.com/wasilibs/go-prettier/cmd/prettier@latest
```

To avoid installation entirely, it can be convenient to use `go run`

```bash
$ go run github.com/wasilibs/go-prettier/cmd/prettier@latest -o formatted.sql unformatted.sql
```

[1]: https://github.com/prettier/prettier
[2]: https://wazero.io/
[3]: https://bellard.org/quickjs/
