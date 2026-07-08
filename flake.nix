{
  description = "Development environment for the dapla API and frontend";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
      perSystem = { pkgs, ... }: {

        devShells.default = pkgs.mkShell {
          name = "dapla devenv";

          packages = with pkgs; [
            go
            gopls
            mise
            nixd
            nodejs
            pnpm
            protobuf
            protobuf-language-server
            protoc-gen-go
            protoc-gen-go-grpc
            yaml-language-server
          ];
        };

        formatter = pkgs.alejandra;
      };
    };
}
