{
  lib,
  runCommand,
  coreutils,
  mq,
  stdenvNoCC,
  fetchFromGitHub,
  nix-update-script,
  newScope,
}:

let
  scope =
    self:
    let
      src = import ./src.nix {
        inherit
          lib
          stdenvNoCC
          fetchFromGitHub
          nix-update-script
          ;
      };
      plugins =
        with lib;
        mapAttrs' (
          name: _value:
          nameValuePair (removeSuffix ".yazi" name) {
            description = trim (
              builtins.readFile (
                runCommand "${name}-description" {
                  nativeBuildInputs = [
                    coreutils
                    mq
                  ];
                } ''mq -q '.text | gsub("([\\.!])(\\s+|$)", "")' ${src}/${name}/README.md | head -qn 1 > "$out"''
              )
            );
          }
        ) (filterAttrs (n: v: (hasSuffix ".yazi" n) && v == "directory") (builtins.readDir src));
    in
    {
      inherit src;
    }
    // (lib.mapAttrs (
      name: value:
      stdenvNoCC.mkDerivation {
        pname = name;
        inherit (src) version;

        inherit src;

        dontBuild = true;

        installPhase = ''
          runHook preInstall

          cp -rL '${name}.yazi' $out

          runHook postInstall
        '';

        meta = with lib; {
          inherit (value) description;
          homepage = "https://github.com/yazi-rs/plugins/tree/main/${name}.yazi";
          license = licenses.mit;
          maintainers = with maintainers; [ xyenon ];
        };
      }
    ) plugins);
in

with lib;
pipe scope [
  (makeScope newScope)
  recurseIntoAttrs
]
