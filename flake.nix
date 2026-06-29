{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    version = {
      url = "github:a-h/version";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    xc = {
      url = "github:joerdav/xc";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, gitignore, version, xc, }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        system = system;
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            (final: prev: {
              xc = xc.outputs.packages.${system}.xc;
              version = version.outputs.packages.${system}.default;
            })
          ];
        };
      });

      v = nixpkgs.lib.strings.trim (builtins.readFile ./.version);

      # Build app.
      app = { name, pkgs, system, }: pkgs.buildGoModule {
        pname = name;
        version = v;
        src = gitignore.lib.gitignoreSource ./.;
        vendorHash = "sha256-gDk56lw+XBvjmnGvpF0jJwAltfKPsEDaajdleK/HYUc=";
        subPackages = [ "cmd/${name}" ];
        ldflags = [
          "-s"
          "-w"
          "-X main.Version=${v}"
        ];
        # Skip tests, we run those as part of CI.
        doCheck = false;
      };

      # Development tools used.
      devTools = pkgs: [
        pkgs.adr-tools
        pkgs.gh
        pkgs.git
        pkgs.go
        pkgs.goreleaser
        pkgs.version
        pkgs.xc
      ];

      name = "cap";
    in
    {
      # `nix build` builds the app.
      packages = forAllSystems ({ system, pkgs }: rec {
        default = app { name = name; pkgs = pkgs; system = system; };
        cap = default;
      });

      # `nix develop` provides a shell containing required tools.
      devShells = forAllSystems ({ system, pkgs }: {
        default = pkgs.mkShell {
          buildInputs = (devTools pkgs);
        };
      });
    };
}
