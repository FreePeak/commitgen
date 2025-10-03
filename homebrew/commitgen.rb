# Documentation: https://docs.brew.sh/Formula-Cookbook
#                https://rubydoc.brew.sh/Formula
# PLEASE REMOVE ALL GENERATED COMMENTS BEFORE SUBMITTING YOUR PULL REQUEST!

class Commitgen < Formula
  desc "AI-powered git commit message generator"
  homepage "https://github.com/FreePeak/commitgen"
  url "https://github.com/FreePeak/commitgen/archive/refs/tags/v0.1.0.tar.gz"
  sha256 ""
  license "MIT"
  head "https://github.com/FreePeak/commitgen.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "-o", bin/"commitgen", "main.go"
  end

  test do
    system "#{bin}/commitgen", "--help"
  end
end