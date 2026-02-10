# This is a template for the Homebrew formula
# GoReleaser will automatically generate and push the actual formula to homebrew-tap
# This file is for reference only

class Moko < Formula
  desc "CLI for querying Deutsche Bahn real-time transit information with interactive TUI"
  homepage "https://github.com/mobil-koeln/moko-cli"
  url "https://github.com/mobil-koeln/moko-cli/releases/download/v0.3.0/moko_0.3.0_darwin_amd64.tar.gz"
  sha256 "GENERATED_BY_GORELEASER"
  license "MIT"
  version "0.3.0"

  def install
    bin.install "moko"
  end

  test do
    system "#{bin}/moko", "--version"
  end
end
