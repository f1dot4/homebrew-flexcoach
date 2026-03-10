class Flexcli < Formula
  desc "Management CLI for FlexCoach AI fitness platform"
  homepage "https://github.com/f1dot4/homebrew-flexcli"
  url "https://github.com/f1dot4/homebrew-flexcli/archive/refs/tags/v0.1.2.tar.gz"
  sha256 "baaf40137d980204f18be5e24e4ae225901e67adff8433ffeb730ef9ab3b46c3"
  license "MIT"

  depends_on "go" => :build

  def install
    # Build the binary from the cmd/flexcli directory
    # std_go_args handles common flags for brew-built Go apps
    system "go", "build", *std_go_args(output: bin/"flexcli"), "./cmd/flexcli"
  end

  test do
    # Simple check to ensure the binary runs and shows help
    output = shell_output("#{bin}/flexcli help")
    assert_match "FlexCLI - FlexCoach Command Line Interface", output
  end
end
