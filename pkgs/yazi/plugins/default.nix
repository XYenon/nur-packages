{
  sources,
  lib,
  newScope,
}:

let
  scope =
    self:
    let
      inherit (self) callPackage;
    in
    {
      exifaudio = callPackage ./exifaudio.nix { };
      fg = callPackage ./fg.nix { };
      ouch = callPackage ./ouch.nix { };
      yazi-rs = callPackage ./yazi-rs { source = sources.yazi-rs-plugins; };
    };
in

with lib;
pipe scope [
  (makeScope newScope)
  recurseIntoAttrs
]
