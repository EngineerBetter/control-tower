#!/bin/bash
set -euo pipefail

version=$(cat release/version)

darwin_cli_sha256=$(openssl dgst -sha256 release/control-tower-darwin-amd64 | cut -d ' ' -f 2)
linux_cli_sha256=$(openssl dgst -sha256 release/control-tower-linux-amd64 | cut -d ' ' -f 2)

pushd homebrew-tap
  cat <<EOF > control-tower.rb
class ControlTower < Formula
  desc "Deploy and operate Concourse CI in a single command"
  homepage "https://www.engineerbetter.com"
  license "Apache-2.0"
  version "${version}"

  if OS.mac?
    url "https://github.com/EngineerBetter/control-tower/releases/download/#{version}/control-tower-darwin-amd64"
    sha256 "${darwin_cli_sha256}"
  elsif OS.linux?
    url "https://github.com/EngineerBetter/control-tower/releases/download/#{version}/control-tower-linux-amd64"
    sha256 "${linux_cli_sha256}"
  end

  depends_on :arch => :x86_64

  def install
    binary_name = "control-tower"
    if OS.mac?
      bin.install "control-tower-darwin-amd64" => binary_name
    elsif OS.linux?
      bin.install "control-tower-linux-amd64" => binary_name
    end
  end

  test do
    system "#{bin}/control-tower --help"
  end
end
EOF

  git add control-tower.rb

  git config --global user.email "systems@engineerbetter.com"
  git config --global user.name "CI"
  git commit -m "Release control-tower version ${version}"
popd

git clone ./homebrew-tap ./homebrew-tap-updated
