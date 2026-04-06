class Meteoswisscli < Formula
  desc "CLI tools for Swiss weather and avalanche data"
  homepage "https://github.com/salam/meteoswisscli"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/salam/meteoswisscli/releases/download/v#{version}/meteoswisscli_#{version}_darwin_arm64.tar.gz"
    end
    on_intel do
      url "https://github.com/salam/meteoswisscli/releases/download/v#{version}/meteoswisscli_#{version}_darwin_amd64.tar.gz"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/salam/meteoswisscli/releases/download/v#{version}/meteoswisscli_#{version}_linux_arm64.tar.gz"
    end
    on_intel do
      url "https://github.com/salam/meteoswisscli/releases/download/v#{version}/meteoswisscli_#{version}_linux_amd64.tar.gz"
    end
  end

  def install
    bin.install "meteoswiss"
    bin.install "whiterisk"
  end

  test do
    system "#{bin}/meteoswiss", "--version"
    system "#{bin}/whiterisk", "--version"
  end
end
