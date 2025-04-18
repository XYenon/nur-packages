# This file was generated by nvfetcher, please do not modify it manually.
{ fetchgit, fetchurl, fetchFromGitHub, dockerTools }:
{
  "WeidiDeng/caddy-cloudflare-ip" = {
    pname = "WeidiDeng/caddy-cloudflare-ip";
    version = "f53b62aa13cb7ad79c8b47aacc3f2f03989b67e5";
    src = fetchFromGitHub {
      owner = "WeidiDeng";
      repo = "caddy-cloudflare-ip";
      rev = "f53b62aa13cb7ad79c8b47aacc3f2f03989b67e5";
      fetchSubmodules = false;
      sha256 = "sha256-b2JPVdXBgbI91zRzXGut7QKYT/pe15Gcr3JvIKxvTH4=";
    };
    moduleName = "github.com/WeidiDeng/caddy-cloudflare-ip";
    date = "20231130082422+0800";
  };
  "caddy-dns/cloudflare" = {
    pname = "caddy-dns/cloudflare";
    version = "1fb64108d4debf196b19d7398e763cb78c8a0f41";
    src = fetchFromGitHub {
      owner = "caddy-dns";
      repo = "cloudflare";
      rev = "1fb64108d4debf196b19d7398e763cb78c8a0f41";
      fetchSubmodules = false;
      sha256 = "sha256-nLpiXMHTKTfmc5TBkPErkvXf/d2tWBlv2h4A+ELrraU=";
    };
    moduleName = "github.com/caddy-dns/cloudflare";
    date = "20250228105314-0700";
  };
  "mholt/caddy-l4" = {
    pname = "mholt/caddy-l4";
    version = "87e3e5e2c7f986b34c0df373a5799670d7b8ca03";
    src = fetchFromGitHub {
      owner = "mholt";
      repo = "caddy-l4";
      rev = "87e3e5e2c7f986b34c0df373a5799670d7b8ca03";
      fetchSubmodules = false;
      sha256 = "sha256-aLAGmbWXhwtuP8JKbceXU9ZtHDgncQC2p3hOZuGRMds=";
    };
    moduleName = "github.com/mholt/caddy-l4";
    date = "20250124164235-0700";
  };
  "mholt/caddy-webdav" = {
    pname = "mholt/caddy-webdav";
    version = "42168ba04c9dc2cd228ab8c453dbab27654e52e6";
    src = fetchFromGitHub {
      owner = "mholt";
      repo = "caddy-webdav";
      rev = "42168ba04c9dc2cd228ab8c453dbab27654e52e6";
      fetchSubmodules = false;
      sha256 = "sha256-LX4R3DQiNPCFirM2M0ddDl0JRaA/EzQJU11Y2Mv9Y7A=";
    };
    moduleName = "github.com/mholt/caddy-webdav";
    date = "20241008102340-0600";
  };
  rime-ice = {
    pname = "rime-ice";
    version = "2025.04.06";
    src = fetchFromGitHub {
      owner = "iDvel";
      repo = "rime-ice";
      rev = "2025.04.06";
      fetchSubmodules = false;
      sha256 = "sha256-s3r8cdEliiPnKWs64Wgi0rC9Ngl1mkIrLnr2tIcyXWw=";
    };
  };
  rime-ice-unstable = {
    pname = "rime-ice-unstable";
    version = "5b1999c3c3bc1bfd84bffacd7ff5c5e0879ea683";
    src = fetchFromGitHub {
      owner = "iDvel";
      repo = "rime-ice";
      rev = "5b1999c3c3bc1bfd84bffacd7ff5c5e0879ea683";
      fetchSubmodules = false;
      sha256 = "sha256-7DodE2/bcYSVoXCyJ9YAeYH80lARIG7XJ9AImh75Kcc=";
    };
    date = "2025-04-16";
  };
  rime-moegirl = {
    pname = "rime-moegirl";
    version = "20250409";
    src = fetchurl {
      url = "https://github.com/outloudvi/mw2fcitx/releases/download/20250409/moegirl.dict.yaml";
      sha256 = "sha256-/MBYX1Zb66RK6Nowky4LdMB6Vt/4n4UAxRTN0KODdwk=";
    };
  };
  yazi-rs-plugins = {
    pname = "yazi-rs-plugins";
    version = "b12a9ab085a8c2fe2b921e1547ee667b714185f9";
    src = fetchFromGitHub {
      owner = "yazi-rs";
      repo = "plugins";
      rev = "b12a9ab085a8c2fe2b921e1547ee667b714185f9";
      fetchSubmodules = false;
      sha256 = "sha256-LWN0riaUazQl3llTNNUMktG+7GLAHaG/IxNj1gFhDRE=";
    };
    date = "2025-04-13";
  };
}
