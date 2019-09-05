# Development of Control Tower

[CI Pipeline](https://ci.engineerbetter.com/teams/main/pipelines/control-tower) (deployed with Control Tower!)

## Development

### Pre-requisites

To build and test you'll need:

- Golang 1.11+
- to have installed `github.com/kevinburke/go-bindata`

### Building locally

`control-tower` uses [golang compile-time variables](https://github.com/golang/go/wiki/GcToolchainTricks#including-build-information-in-the-executable) to set the release versions it uses. To build locally use the `build_local.sh` script, rather than running `go build`.

You will also need to clone [`control-tower-ops`](https://github.com/EngineerBetter/control-tower-ops) to the same level as `control-tower` to get the manifest and ops files necessary for building. Check the latest release of `control-tower` for the appropriate tag of `control-tower-ops`

### Tests

Tests use the [Ginkgo](https://onsi.github.io/ginkgo/) Go testing framework. The tests require you to have set up AWS authentication locally.

Install ginkgo and run the tests with:

```sh
go get github.com/onsi/ginkgo/ginkgo
ginkgo -r
```

```sh
go get github.com/onsi/ginkgo/ginkgo
ginkgo -r
```

Go linting, shell linting, and unit tests can be run together in the same docker image CI uses with `./run_tests_local.sh`. This should be done before committing or raising a PR.

### Bumping Manifest/Ops File versions

The pipeline listens for new patch or minor versions of `manifest.yml` and `ops/versions.json` coming from the `control-tower-ops` repo. In order to pick up a new major version first make sure it exists in the repo then modify `tag_filter: X.*.*` in the `control-tower-ops` resource where `X` is the major version you want to pin to.
