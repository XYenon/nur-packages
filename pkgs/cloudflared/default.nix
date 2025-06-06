{
  lib,
  stdenv,
  buildGoModule,
  buildPackages,
  fetchFromGitHub,
  callPackage,
  gitUpdater,
}:

let
  # https://github.com/cloudflare/cloudflared/issues/1151#issuecomment-1888819250
  # https://github.com/cloudflare/cloudflared/blob/master/.teamcity/install-cloudflare-go.sh
  buildGoModule' = buildGoModule.override {
    go = buildPackages.go_1_22.overrideAttrs {
      pname = "cloudflare-go";
      version = "1.22.5-devel-cf";
      src = fetchFromGitHub {
        owner = "cloudflare";
        repo = "go";
        rev = "af19da5605ca11f85776ef7af3384a02a315a52b";
        hash = "sha256-6VT9CxlHkja+mdO1DeFoOTq7gjb3T5jcf2uf9TB/CkU=";
      };
    };
  };
in

buildGoModule' rec {
  pname = "cloudflared";
  version = "2025.5.0";

  src = fetchFromGitHub {
    owner = "cloudflare";
    repo = "cloudflared";
    tag = version;
    hash = "sha256-ZnkE9x4A9HoiSXzvYuzyW/dH08r0aJUk/q6gFVgtTjk=";
  };

  vendorHash = null;

  ldflags = [
    "-s"
    "-w"
    "-X main.Version=${version}"
    "-X github.com/cloudflare/cloudflared/cmd/cloudflared/updater.BuiltForPackageManager=nixpkgs"
  ];

  preCheck = ''
    # Workaround for: sshgen_test.go:74: mkdir /homeless-shelter/.cloudflared: no such file or directory
    export HOME="$(mktemp -d)"

    # Workaround for: protocol_test.go:11:
    #   lookup protocol-v2.argotunnel.com on [::1]:53: read udp [::1]:51876->[::1]:53: read: connection refused
    substituteInPlace "edgediscovery/protocol_test.go" \
      --replace "TestProtocolPercentage" "SkipProtocolPercentage"

    # Workaround for: origin_icmp_proxy_test.go:46:
    #   cannot create ICMPv4 proxy: socket: permission denied nor ICMPv6 proxy: socket: permission denied
    substituteInPlace "ingress/origin_icmp_proxy_test.go" \
      --replace "TestICMPRouterEcho" "SkipICMPRouterEcho"

    # Workaround for: origin_icmp_proxy_test.go:110:
    #   cannot create ICMPv4 proxy: socket: permission denied nor ICMPv6 proxy: socket: permission denied
    substituteInPlace "ingress/origin_icmp_proxy_test.go" \
      --replace "TestConcurrentRequestsToSameDst" "SkipConcurrentRequestsToSameDst"

    # Workaround for: origin_icmp_proxy_test.go:242:
    #   cannot create ICMPv4 proxy: socket: permission denied nor ICMPv6 proxy: socket: permission denied
    substituteInPlace "ingress/origin_icmp_proxy_test.go" \
      --replace "TestICMPRouterRejectNotEcho" "SkipICMPRouterRejectNotEcho"

    # Workaround for: origin_icmp_proxy_test.go:108:
    #   Received unexpected error: cannot create ICMPv4 proxy: Group ID 100 is not between ping group 65534 to 65534 nor ICMPv6 proxy: socket: permission denied
    substituteInPlace "ingress/origin_icmp_proxy_test.go" \
      --replace "TestTraceICMPRouterEcho" "SkipTraceICMPRouterEcho"

    # Workaround for: icmp_posix_test.go:28: socket: permission denied
    substituteInPlace "ingress/icmp_posix_test.go" \
      --replace "TestFunnelIdleTimeout" "SkipFunnelIdleTimeout"

    # Workaround for: icmp_posix_test.go:88: Received unexpected error: Group ID 100 is not between ping group 65534 to 65534
    substituteInPlace "ingress/icmp_posix_test.go" \
      --replace "TestReuseFunnel" "SkipReuseFunnel"

    # Workaround for: manager_test.go:197:
    #   Should be false
    substituteInPlace "datagramsession/manager_test.go" \
      --replace "TestManagerCtxDoneCloseSessions" "SkipManagerCtxDoneCloseSessions"
  '';

  doCheck = !stdenv.hostPlatform.isDarwin;

  passthru = {
    tests.simple = callPackage ./tests.nix { inherit version; };
    updateScript = gitUpdater { };
  };

  meta = with lib; {
    description = "Cloudflare Tunnel daemon, Cloudflare Access toolkit, and DNS-over-HTTPS client";
    homepage = "https://www.cloudflare.com/products/tunnel";
    changelog = "https://github.com/cloudflare/cloudflared/releases/tag/${version}";
    license = licenses.asl20;
    platforms = platforms.unix ++ platforms.windows;
    maintainers = with maintainers; [ xyenon ];
    mainProgram = "cloudflared";
    # cloudflared requires a fork of go 1.22 to build (see override above),
    # but go 1.22 is EOL and the toolchain has been removed from nixpkgs.
    broken =
      (!lib.hasAttr "go_1_22" buildPackages) || !(builtins.tryEval buildPackages.go_1_22).success;
  };
}
