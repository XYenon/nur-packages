{ lib, newScope }:

let
  scope = self:
    let inherit (self) callPackage; in
    {
      fg = callPackage ./fg.nix { };
    };
in

with lib;
pipe scope [
  (makeScope newScope)
  recurseIntoAttrs
]
