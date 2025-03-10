{
  lib,
  stdenvNoCC,
  fetchFromGitHub,
  nix-update-script,
}:

stdenvNoCC.mkDerivation {
  pname = "yazi-rs-plugins";
  version = "0-unstable-2025-03-04";

  src = fetchFromGitHub {
    owner = "yazi-rs";
    repo = "plugins";
    rev = "a1b678dfacfd2726fad364607aeaa7e1fded3cfa";
    hash = "sha256-Vvq7uau+UNcriuLE7YMK5rSOXvVaD0ElT59q+09WwdQ=";
  };

  dontBuild = true;

  installPhase = ''
    runHook preInstall

    cp -r $src $out

    runHook postInstall
  '';

  passthru.updateScript = nix-update-script { extraArgs = [ "--version=branch" ]; };

  meta = with lib; {
    description = "Place code snippets from docs into this monorepo, so that users can update more easily via package manager";
    homepage = "https://github.com/yazi-rs/plugins";
    license = licenses.mit;
    maintainers = with maintainers; [ xyenon ];
  };
}
