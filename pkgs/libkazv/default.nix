{
  lib,
  stdenv,
  fetchFromGitLab,
  cmake,
  pkg-config,
  lager,
  immer,
  zug,
  boost,
  nlohmann_json,
  vodozemac-bindings-kazv-unstable,
  cryptopp,
  libcpr,
  libhttpserver,
  libmicrohttpd,
  catch2_3,
  nix-update-script,
}:

stdenv.mkDerivation (finalAttrs: {
  pname = "libkazv";
  version = "0.8.0-unstable-2025-07-06";

  src = fetchFromGitLab {
    domain = "lily-is.land";
    owner = "kazv";
    repo = "libkazv";
    rev = "a02757c5db9de927db3781e87379b24ed7b0ce90";
    hash = "sha256-/zprR8INPGtfAnWHviXAgwAp+rvT5OSAIPs0x6EpFfc=";
  };

  nativeBuildInputs = [
    cmake
    pkg-config
  ];

  buildInputs = [
    lager
    immer
    zug
    boost
    nlohmann_json
    vodozemac-bindings-kazv-unstable
    cryptopp

    libcpr
    libhttpserver
    libmicrohttpd
  ];

  strictDeps = true;

  cmakeFlags = [ (lib.cmakeBool "libkazv_BUILD_TESTS" finalAttrs.doCheck) ];

  doCheck = true;

  checkInputs = [ catch2_3 ];

  passthru.updateScript = nix-update-script { extraArgs = [ "--version=branch" ]; };

  meta = with lib; {
    description = "Sans-io C++ (gnu++17) matrix client library";
    homepage = "https://lily-is.land/kazv/libkazv";
    license = licenses.agpl3Plus;
    maintainers = with maintainers; [ xyenon ];
  };
})
