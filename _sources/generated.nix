# This file was generated by nvfetcher, please do not modify it manually.
{ fetchgit, fetchurl, fetchFromGitHub, dockerTools }:
{
  rime-ice = {
    pname = "rime-ice";
    version = "2024.12.12";
    src = fetchFromGitHub {
      owner = "iDvel";
      repo = "rime-ice";
      rev = "2024.12.12";
      fetchSubmodules = false;
      sha256 = "sha256-2QZdlLGZwWIesbjYTE/2yhM1hHGVVp7jR02bR0oqxV0=";
    };
  };
  rime-ice-unstable = {
    pname = "rime-ice-unstable";
    version = "75b225e84f5e95ddf0b5a50fa6e2cff9800ee7b7";
    src = fetchFromGitHub {
      owner = "iDvel";
      repo = "rime-ice";
      rev = "75b225e84f5e95ddf0b5a50fa6e2cff9800ee7b7";
      fetchSubmodules = false;
      sha256 = "sha256-/sKoufIf4m5h1Zds6AqIyeCjDaZsFU9ot5ln3ykuNyU=";
    };
    date = "2025-03-11";
  };
  rime-moegirl = {
    pname = "rime-moegirl";
    version = "20250309";
    src = fetchurl {
      url = "https://github.com/outloudvi/mw2fcitx/releases/download/20250309/moegirl.dict.yaml";
      sha256 = "sha256-LUKh0UMU0WVIWAGWogzWjqEVKCY7TTvWRueMeFxVslk=";
    };
  };
  yazi-rs-plugins = {
    pname = "yazi-rs-plugins";
    version = "cb6165b25515b653a70f72a67889579d190facfe";
    src = fetchFromGitHub {
      owner = "yazi-rs";
      repo = "plugins";
      rev = "cb6165b25515b653a70f72a67889579d190facfe";
      fetchSubmodules = false;
      sha256 = "sha256-XDz67eHmVM5NrnQ/uPXN/jRgmrShs80anWnHpVmbPO8=";
    };
    date = "2025-03-12";
  };
}
