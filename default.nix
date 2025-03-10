# This file describes your repository contents.
# It should return a set of nix derivations
# and optionally the special attributes `lib`, `modules` and `overlays`.
# It should NOT import <nixpkgs>. Instead, you should take pkgs as an argument.
# Having pkgs default to <nixpkgs> is fine though, and it lets you use short
# commands such as:
#     nix-build -A mypackage

let
  lock = builtins.fromJSON (builtins.readFile ./flake.lock);
in
{
  pkgs ? import <nixpkgs> { overlays = [ rust-overlay ]; },
  rust-overlay ? import (
    (import <nixpkgs> { }).fetchzip {
      url = "https://github.com/oxalica/rust-overlay/archive/${lock.nodes.rust-overlay.locked.rev}.zip";
      hash = lock.nodes.rust-overlay.locked.narHash;
    }
  ),
}:

with pkgs;
rec {
  # The `lib`, `modules`, and `overlays` names are special
  lib = import ./lib { inherit pkgs; }; # functions
  modules = import ./modules; # NixOS modules
  hmModules = import ./hm-modules; # Home Manager modules
  overlays = import ./overlays; # nixpkgs overlays

  go-check = callPackage ./pkgs/go-check { };
  catp = callPackage ./pkgs/catp { };
  vodozemac-bindings-kazv-unstable = callPackage ./pkgs/vodozemac-bindings-kazv { };
  libkazv = callPackage ./pkgs/libkazv {
    inherit vodozemac-bindings-kazv-unstable;
    libcpr = pkgs.libcpr_1_10_5 or libcpr;
  };
  kazv = kdePackages.callPackage ./pkgs/kazv { inherit vodozemac-bindings-kazv-unstable libkazv; };
  nginxModules = callPackage ./pkgs/nginx/modules { };
  nginxStable =
    let
      nginxStable' = pkgs.nginxStable;
    in
    nginxStable'.override {
      modules =
        nginxStable'.modules ++ (with nginxModules; [ (http_proxy_connect nginxStable'.version) ]);
    };
  nginxMainline =
    let
      nginxMainline' = pkgs.nginxMainline;
    in
    nginxMainline'.override {
      modules =
        nginxMainline'.modules ++ (with nginxModules; [ (http_proxy_connect nginxMainline'.version) ]);
    };
  nginx = nginxStable;
  mpvScripts = callPackage ./pkgs/mpv/scripts { };
  anime4k = callPackage ./pkgs/anime4k { };
  yaziPlugins = callPackage ./pkgs/yazi/plugins { inherit mq; };
  telemikiya = callPackage ./pkgs/telemikiya { };
  mq = callPackage ./pkgs/mq { };
}
