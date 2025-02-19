{
  lib,
  fetchFromGitHub,
  buildGoModule,
  unstableGitUpdater,
}:

buildGoModule {
  pname = "telemikiya";
  version = "0-unstable-2025-02-21";

  src = fetchFromGitHub {
    owner = "XYenon";
    repo = "TeleMikiya";
    rev = "b3fe3baae5ea6cd300f5dba2d72f91bf03779ff4";
    hash = "sha256-XwsSchD8h9SuiLPqkCDZBfn33SgUqdTjuQTZrivtQkE=";
  };

  vendorHash = "sha256-87zQuJyQVuCHvUg4x+OuF+S+Lhv9zmjQUEDiqXBhn9Y=";

  subPackages = [ "." ];
  ldflags = [
    "-s"
    "-w"
  ];

  passthru.updateScript = unstableGitUpdater { };

  meta = with lib; {
    mainProgram = "telemikiya";
    description = "Hybrid message search tool for Telegram";
    homepage = "https://github.com/XYenon/TeleMikiya";
    license = licenses.agpl3Plus;
    maintainers = with maintainers; [ xyenon ];
  };
}
