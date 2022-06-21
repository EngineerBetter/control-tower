# Development of Control Tower

[CI Pipeline](https://ci.engineerbetter.com/teams/main/pipelines/control-tower) (deployed with Control Tower!)

## Development

### Pre-requisites

To build and test you'll need:

- Golang 1.16+
- [`control-tower-ops`](https://github.com/EngineerBetter/control-tower-ops) cloned at the same level as this repository (i.e. a sibling directory). Check the latest release of `control-tower` for the appropriate tag of `control-tower-ops`.

### Building locally

`control-tower` uses [golang compile-time variables](https://github.com/golang/go/wiki/GcToolchainTricks#including-build-information-in-the-executable) to set the release versions it uses. To build locally use the `build_local.sh` script, rather than running `go build`.

### Linting

Run our linting script (requires `gometalinter` to be installed):

```sh
./ci/tasks/lint.sh
```

### Unit Tests

To Run tests from your host:

```sh
go install github.com/maxbrunsfeld/counterfeiter/v6

cp ../control-tower-ops/manifest.yml opsassets/assets/
cp -R ../control-tower-ops/ops opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json opsassets/assets/

go generate ./...
go test ./...
```

Alternatively, you can run both the tests and linting script with Docker using the same container image that we use in our [CI pipeline](https://github.com/EngineerBetter/control-tower/blob/master/ci/pipeline.yml) by running `./run_tests_local.sh`

This should be done before committing or raising a PR.

### Bumping Manifest/Ops File versions

The pipeline listens for new patch or minor versions of `manifest.yml` and `ops/versions.json` coming from the `control-tower-ops` repo. In order to pick up a new major version first make sure it exists in the repo then modify `tag_filter: X.*.*` in the `control-tower-ops` resource where `X` is the major version you want to pin to.
