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
  endc
end
