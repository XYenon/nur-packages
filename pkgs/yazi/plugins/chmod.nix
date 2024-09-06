{
  lib,
  stdenvNoCC,
  fetchFromGitHub,
  unstableGitUpdater,
}:

stdenvNoCC.mkDerivation {
  pname = "chmod";
  version = "0-unstable-2024-09-05";

  src = fetchFromGitHub {
    owner = "yazi-rs";
    repo = "plugins";
    rev = "9a1129cf3919f0a75865ba2cb05ba08ceee5a548";
    hash = "sha256-g+NBP0Yg2IpunQckooB2r0KmH83YespD1jDllriNQZ8=";
  };

  dontBuild = true;

  installPhase = ''
    runHook preInstall

    cp -r chmod.yazi $out

    runHook postInstall
  '';

  passthru.updateScript = unstableGitUpdater { };

  meta = with lib; {
    description = "Execute chmod on the selected files to change their mode";
    homepage = "https://github.com/yazi-rs/plugins/tree/main/chmod.yazi";
    license = licenses.mit;
    maintainers = with maintainers; [ xyenon ];
  };
}
