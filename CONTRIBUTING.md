# CONTRIBUTING

Contributions are always welcome, no matter how large or small. Before contributing,
please check to see if the issue is already being tracked and if there is already a PR.

## Setup

> Install Go 1.15.x

Chestnut uses the Go Modules support built into Go 1.11 to build. 
The easiest is to clone Chestnut in a directory outside of GOPATH, 
as in the following example:

```sh
$ git clone https://github.com/jrapoport/chestnut
$ cd chestnut
$ make deps
```

## Running examples

```sh
$ make examples
```

## Testing

```sh
$ make test
```

## Pull Requests

Pull requests are welcome!.

1. Fork the repo and create your branch from `master`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.

```sh
# will run fmt, lint, & vet
$ make pr
```

## License

By contributing to Chestnut, you agree that your contributions will be licensed
under its [MIT license](LICENSE).
