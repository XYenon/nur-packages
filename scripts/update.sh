#!/usr/bin/env bash
# See https://discourse.nixos.org/t/25274
set -Eeuo pipefail

root="$(readlink --canonicalize -- "$(dirname -- "$0")/..")"

rust_overlay=$(
	cat <<-EOF
		let
			lock = builtins.fromJSON (builtins.readFile ./flake.lock);
		in
		import (builtins.fetchTarball {
			url = "https://github.com/oxalica/rust-overlay/archive/\${lock.nodes.rust-overlay.locked.rev}.zip";
			sha256 = lock.nodes.rust-overlay.locked.narHash;
		})
	EOF
)

# Run update scripts
nixpkgs="$(nix-instantiate --eval --expr '<nixpkgs>')"
nix-shell "$nixpkgs/maintainers/scripts/update.nix" --show-trace \
	--arg include-overlays "[ (import ./overlay.nix) ($rust_overlay) ]" \
	--argstr keep-going 'true' \
	--argstr commit 'true' \
	--arg predicate "(
    let prefix = \"$root/pkgs/\"; prefixLen = builtins.stringLength prefix;
    getPosition = p: p.meta.position or (builtins.trace p.meta \"\");
    in (path: p: (builtins.substring 0 prefixLen (getPosition p)) ==  prefix)
  )"

# Clean up
if [[ -f "$root/update-git-commits.txt" ]]; then
	cat "$root/update-git-commits.txt" && rm "$root/update-git-commits.txt"
fi
