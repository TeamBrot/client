# Contributing guidelines

## Before contributing

Please familiarize yourself with the codebase before writing any code. If you have any questions, feel free to reach out to us! You can write us emails or open an issue.

You can find our full documentation under [https://github.com/TeamBrot/paper](https://github.com/TeamBrot/paper).

## Contributions

Always welcome are:

- Improving documentation
- Adding tests
- Improving logging / error messages
- Improving error message
- Improving performance / security
- Minor refactorings

For other types of contributions, consider opening an issue first so we can chat about the possible change.

If you are considering adding a client, please create a new `.go` file containing a structure that implements the `Client` interface that can be found in [`client.go`](client/client.go). You can look at [`clientBasic.go`](client/clientBasic.go) for an example.

## Style

Please use govet and gofmt to check your code style. Please run `go test` to check that all tests still pass.

## Code of conduct

When contributing, please also check out our code of conduct in [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

