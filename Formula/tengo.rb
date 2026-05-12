class Tengo < Formula
  desc "Fast terminal editor for structured config files"
  homepage "https://github.com/helloWorld44-89/tengo"
  url "https://github.com/helloWorld44-89/tengo/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "PLACEHOLDER_RUN_SHA256_AFTER_TAGGING"
  license "MIT"
  head "https://github.com/helloWorld44-89/tengo.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build",
      *std_go_args(ldflags: "-s -w -X main.version=#{version}"),
      "."

    man1.install "tengo.1"

    (bash_completion/"tengo").write Utils.safe_popen_read(bin/"tengo", "-completion", "bash")
    (zsh_completion/"_tengo").write Utils.safe_popen_read(bin/"tengo", "-completion", "zsh")
    (fish_completion/"tengo.fish").write Utils.safe_popen_read(bin/"tengo", "-completion", "fish")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/tengo -version")
    (testpath/"test.json").write('{"key":"value"}')
    assert_match "valid", shell_output("#{bin}/tengo #{testpath}/test.json -validate").downcase
  end
end
