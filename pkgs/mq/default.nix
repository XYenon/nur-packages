{
  lib,
  makeRustPlatform,
  rust-bin,
  fetchFromGitHub,
  installShellFiles,
  stdenv,
  nix-update-script,
}:

(makeRustPlatform {
  rustc = rust-bin.beta.latest.minimal;
  cargo = rust-bin.beta.latest.minimal;
}).buildRustPackage
  {
    pname = "mq";
    version = "0-unstable-2025-03-09";

    src = fetchFromGitHub {
      owner = "harehare";
      repo = "mq";
      rev = "d1b759474d89798ff723d0a445a3a0983125e53d";
      hash = "sha256-SMdrIHKDGx0aVVU+/K14tObid8rGLCnqDQLythpzW5M=";
    };

    useFetchCargoVendor = true;
    cargoHash = "sha256-lku0aZjsaYZNUCC+yy64EdJhTToiAgAWehCmjPFZC9s=";

    nativeBuildInputs = [ installShellFiles ];

    postInstall = lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) ''
      installShellCompletion --cmd mq \
        --bash <($out/bin/mq completion --shell bash) \
        --fish <($out/bin/mq completion --shell fish) \
        --zsh <($out/bin/mq completion --shell zsh)
    '';

    passthru.updateScript = nix-update-script { extraArgs = [ "--version=branch" ]; };

    meta = with lib; {
      description = "Jq like markdown processing tool";
      homepage = "https://github.com/harehare/mq";
      license = licenses.mit;
      maintainers = with maintainers; [ xyenon ];
      mainProgram = "mq";
    };
  }
