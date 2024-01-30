{ lib
, stdenv
, fetchFromGitHub
, fetchFromGitLab
, cmake
, pkg-config
, immer-unstable
, zug-unstable
, boost
, openssl
, curl
, nlohmann_json
, olm
, cryptopp
, catch2_3
, unstableGitUpdater
}:

let
  lager = stdenv.mkDerivation rec {
    pname = "lager";
    version = "unstable";
    src = fetchFromGitHub {
      owner = "arximboldi";
      repo = pname;
      rev = "8fd338d6d67dd2efc4f38c0634c2f849a7db8774";
      hash = "sha256-SJ3+DjMQ3ZRUtCov9rscNhk4Zsdqizcz4ae3szs/wro=";
    };
    nativeBuildInputs = [ cmake pkg-config ];
    buildInputs = [ boost immer-unstable zug-unstable ];
    cmakeFlags = [ "-Dlager_BUILD_TESTS=ON" "-Dlager_BUILD_EXAMPLES=OFF" "-Dlager_BUILD_DOCS=OFF" ];
    meta = with lib; {
      homepage = "https://github.com/arximboldi/lager";
      description = "C++ library for value-oriented design using the unidirectional data-flow architecture — Redux for C++";
      license = licenses.mit;
    };
  };

  cpr = stdenv.mkDerivation rec {
    pname = "cpr";
    version = "1.6.0";
    src = fetchFromGitHub {
      owner = "libcpr";
      repo = pname;
      rev = "refs/tags/${version}";
      hash = "sha256-Y2jCINX4Gb5bKLCibYmtDllk1HrxwZE62BBHABK4TX0=";
    };
    nativeBuildInputs = [ cmake ];
    buildInputs = [ openssl ];
    propagatedBuildInputs = [ curl ];
    cmakeFlags = [ "-DCPR_FORCE_USE_SYSTEM_CURL=ON" "-DCPR_BUILD_TESTS=OFF" ];
    meta = with lib; {
      homepage = "https://github.com/libcpr/cpr";
      description = "C++ Requests: Curl for People, a spiritual port of Python Requests";
      license = licenses.mit;
    };
  };
in
stdenv.mkDerivation rec {
  pname = "libkazv";
  version = "unstable-2024-01-13";

  src = fetchFromGitLab {
    domain = "lily-is.land";
    owner = "kazv";
    repo = pname;
    rev = "4502fbe9194b2c05c9eacf34dfd3fbba46eb74ae";
    hash = "sha256-Qv31KzuL8pts7VrnxJlX8UKvZjZY6h2OE9VK22NScTA=";
  };

  nativeBuildInputs = [ cmake pkg-config ];

  propagatedBuildInputs = [
    lager
    immer-unstable
    zug-unstable
    boost
    nlohmann_json
    olm
    cryptopp

    cpr

    catch2_3
  ];

  cmakeFlags = [
    "-Dlibkazv_BUILD_KAZVJOB=ON"
    "-Dlibkazv_INSTALL_HEADERS=ON"
    "-Dlibkazv_BUILD_EXAMPLES=OFF"
    "-Dlibkazv_BUILD_TESTS=ON"
  ];

  passthru.updateScript = unstableGitUpdater { };

  meta = with lib; {
    description = "A sans-io C++ (gnu++17) matrix client library";
    homepage = "https://lily-is.land/kazv/libkazv";
    license = licenses.agpl3Plus;
    maintainers = with maintainers; [ xyenon ];
  };
}
