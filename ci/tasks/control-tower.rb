class ControlTower < Formula
  desc "Deploy and operate Concourse CI in a single command"
  homepage "https://www.engineerbetter.com"
  version "__version__"
  license "Apache-2.0"

  is_arm64 = RUBY_PLATFORM.include("arm64")

  if OS.mac?
    if is_arm64
      url "https://github.com/EngineerBetter/control-tower/releases/download/#{version}/control-tower-darwin-arm64"
      sha256 "__darwin_cli_arm64_sha256__"
    else
      url "https://github.com/EngineerBetter/control-tower/releases/download/#{version}/control-tower-darwin-amd64"
      sha256 "__darwin_cli_amd64_sha256__"
    end
  elsif OS.linux?
    url "https://github.com/EngineerBetter/control-tower/releases/download/#{version}/control-tower-linux-amd64"
    sha256 "__linux_cli_amd64_sha256__"
  end

  def install
    binary_name = "control-tower"
    if OS.mac?
      if is_arm64
        bin.install "control-tower-darwin-arm64" => binary_name
      else
        bin.install "control-tower-darwin-amd64" => binary_name
      end
    elsif OS.linux?
      bin.install "control-tower-linux-amd64" => binary_name
    end
  end

  test do
    system "#{bin}/control-tower", "--help"
  end
end
