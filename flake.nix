{
  description = "RWDT sauk reader";

  # A collection of packages for Nix.
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  
  # A utility library for Nix flakes.
  # This is used to simplify the flake configuration.
  inputs.flake-utils.url = "github:numtide/flake-utils";

  # A tool to convert Go modules to Nix expressions.
  # This integrates Go module dependencies into the Nix build system.
  inputs.gomod2nix.url = "github:nix-community/gomod2nix";
  inputs.gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
  inputs.gomod2nix.inputs.flake-utils.follows = "flake-utils";

  # defines what this flake provides, including packages and development shells
  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};

          # The current default sdk for macOS fails to compile go projects, so we use a newer one for now.
          # This has no effect on other platforms.
          callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
        in
        rec {
          # Package definitions
          packages.default = callPackage ./. {
            inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
          };

          # Docker container image definition
          # This creates a Docker image named rwdt-sauk-reader with the version tag 0.1.
          packages.container = pkgs.dockerTools.buildImage {
            name = "rwdt-sauk-reader";
            tag = "0.1";
            created = "now";
            copyToRoot = pkgs.buildEnv {
              name = "image-root";
              paths = [ packages.default pkgs.cacert ];
              pathsToLink = [ "/bin" ];
            };
            config.Cmd = [ "${packages.default}/bin/sauk-reader" ];
            config.Env = [ "SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt" ];
          };
          
          # Development shell
          devShells.default = callPackage ./shell.nix {
            inherit (gomod2nix.legacyPackages.${system}) mkGoEnv gomod2nix;
          };
        })
    );
}
