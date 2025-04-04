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
    version = "a5f5404e369100fcfc5562f86f1205827453e31c";
    src = fetchFromGitHub {
      owner = "iDvel";
      repo = "rime-ice";
      rev = "a5f5404e369100fcfc5562f86f1205827453e31c";
      fetchSubmodules = false;
      sha256 = "sha256-s3r8cdEliiPnKWs64Wgi0rC9Ngl1mkIrLnr2tIcyXWw=";
    };
    date = "2025-04-01";
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
    version = "9a095057d698aaaedc4dd23d638285bd3fd647e9";
    src = fetchFromGitHub {
      owner = "yazi-rs";
      repo = "plugins";
      rev = "9a095057d698aaaedc4dd23d638285bd3fd647e9";
      fetchSubmodules = false;
      sha256 = "sha256-Lx+TliqMuaXpjaUtjdUac7ODg2yc3yrd1mSWJo9Mz2Q=";
    };
    date = "2025-04-04";
  };
}
