{
  lib,
  stdenv,
  fetchFromGitHub,
  cmake,
  catch2_3,
  unstableGitUpdater,
  ...
}:

stdenv.mkDerivation rec {
  pname = "immer";
  version = "0.8.1-unstable-2024-08-22";

  src = fetchFromGitHub {
    owner = "arximboldi";
    repo = pname;
    rev = "d3860a6783446392176907eb8e9b8d69be57d80c";
    hash = "sha256-xHAqMhk+LFELQNre3ZR+iqUwho/wXUcAelLFw4e1tlw=";
  };

  nativeBuildInputs = [ cmake ];
  buildInputs = [ catch2_3 ];

  cmakeFlags = [
    "-Dimmer_BUILD_TESTS=ON"
    "-Dimmer_BUILD_EXAMPLES=OFF"
    "-Dimmer_BUILD_DOCS=OFF"
    "-Dimmer_BUILD_EXTRAS=OFF"
  ];

  passthru.updateScript = unstableGitUpdater { tagPrefix = "v"; };

  meta = with lib; {
    homepage = "http://sinusoid.es/immer";
    description = "Postmodern immutable and persistent data structures for C++ — value semantics at scale";
    license = licenses.boost;
    maintainers = with maintainers; [ xyenon ];
  };
}
