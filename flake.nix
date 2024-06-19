{
  description = "Development environment for Dapla Ctrl";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = ["x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin"];
      perSystem = {pkgs, ...}: {
        devShells.default = pkgs.mkShell {
          shellHook = ''
            export DAPLA_TEAM_API_URL=https://dapla-team-api-v2.staging-bip-app.ssb.no
            export PORT=3000
            export DAPLA_CTRL_ADMIN_GROUPS=dapla-stat-developers,dapla-skyinfra-developers,dapla-utvik-developers
            export DAPLA_CTRL_DOCUMENTATION_URL=https://statistics-norway.atlassian.net/wiki/x/EYC24g
          '';
          packages = with pkgs; [
            nixd
            nodejs
            nodePackages.nodemon
            nodePackages.typescript-language-server
            pandoc
          ];
        };
        formatter = pkgs.alejandra;
      };
    };
}
