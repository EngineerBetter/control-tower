#!/bin/bash
set -euo pipefail

version=$(cat release/version)

darwin_cli_amd64_sha256=$(openssl dgst -sha256 release/control-tower-darwin-amd64 | cut -d ' ' -f 2)
darwin_cli_arm64_sha256=$(openssl dgst -sha256 release/control-tower-darwin-arm64 | cut -d ' ' -f 2)
linux_cli_amd64_sha256=$(openssl dgst -sha256 release/control-tower-linux-amd64 | cut -d ' ' -f 2)

pushd homebrew-tap
  sed -i -e "s/__darwin_cli_arm64_sha256__/$darwin_cli_arm64_sha256/g" control-tower.rb
  sed -i -e "s/__darwin_cli_amd64_sha256__/$darwin_cli_amd64_sha256/g" control-tower.rb
  sed -i -e "s/__linux_cli_amd64_sha256__/$linux_cli_amd64_sha256/g" control-tower.rb
  sed -i -e "s/__version__/$version/g" control-tower.rb

  git add control-tower.rb

  git config --global user.email "systems@engineerbetter.com"
  git config --global user.name "CI"
  git commit -m "Release control-tower version ${version}"
popd

git clone ./homebrew-tap ./homebrew-tap-updated
