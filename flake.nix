{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

    devshell.inputs.nixpkgs.follows = "nixpkgs";
    devshell.url = "github:numtide/devshell";

    flake-parts.inputs.nixpkgs-lib.follows = "nixpkgs";

    flake-parts.url = "github:hercules-ci/flake-parts";

    treefmt-nix.url = "github:numtide/treefmt-nix";
    treefmt-nix.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        # systems for which you want to build the `perSystem` attributes
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      imports = [
        inputs.treefmt-nix.flakeModule
      ];
      perSystem = { self', lib, config, pkgs, ... }:
        let
          devshell = pkgs.callPackage inputs.devshell {
            inherit inputs;
          };
        in
        {
          # Auto formatters. This also adds a flake check to ensure that the
          # source tree was auto formatted.
          treefmt.config = {
            projectRootFile = "flake.nix";
            programs.nixpkgs-fmt.enable = true;

            programs.prettier.enable = true;
            
            programs.shfmt.enable = true;
          };

          devShells.default = devshell.mkShell {
            env = [
              {
                name = "NIX_PATH";
                value = "nixpkgs=${toString pkgs.path}";
              }
            ];

            packages = [
              pkgs.gh
              pkgs.gitAndTools.git-absorb
              config.treefmt.build.wrapper
              pkgs.gcc
              pkgs.go_1_20
              pkgs.gotools
              pkgs.gopls
              pkgs.go-outline
              pkgs.gocode
              pkgs.gopkgs
              pkgs.gocode-gomod
              pkgs.godef
              pkgs.golint
              pkgs.delve
              pkgs.go-tools
              pkgs.chromedriver
            ];
            commands = [
              { name = "head-ref"; category = "dev"; help = "get head ref"; command = ''git rev-parse --short HEAD''; }
            ];
          };
        };
    };
}
