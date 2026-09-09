{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixgl = {
      url = "github:nix-community/nixGL";
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

  outputs =
    {
      self,
      nixpkgs,
      gitignore,
      nixgl,
      version,
      xc,
    }:
    let
      lib = nixpkgs.lib;
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];

      forAllSystems =
        f:
        nixpkgs.lib.genAttrs allSystems (
          system:
          f {
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
          }
        );

      v = nixpkgs.lib.strings.trim (builtins.readFile ./.version);

      # Vendor hash covering all Go dependencies including Wails.
      vendorHash = "sha256-oXowhV2SBVTJOOFijShNUjwZyk0yGmzCLx07opmPF7k=";

      # Build the cap CLI (pure Go, no CGO).
      app =
        {
          name,
          pkgs,
          system,
        }:
        pkgs.buildGoModule {
          pname = name;
          version = v;
          src = gitignore.lib.gitignoreSource ./.;
          inherit vendorHash;
          subPackages = [ "cmd/${name}" ];
          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${v}"
          ];
          # Skip tests, we run those as part of CI.
          doCheck = false;
        };

      # Build the capui desktop app for the host platform (requires CGO + WebView).
      capuiApp =
        { pkgs, system }:
        pkgs.buildGoModule {
          pname = "capui";
          version = v;
          src = gitignore.lib.gitignoreSource ./.;
          inherit vendorHash;
          subPackages = [ "cmd/capui" ];
          nativeBuildInputs = webviewCGOInputs system pkgs;
          buildInputs = webviewCGOInputs system pkgs;
          # The nixpkgs Go hook sets CGO_ENABLED=0 by default; override via preBuild
          # so we do not conflict with the env attr that buildGoModule manages internally.
          preBuild = "export CGO_ENABLED=1";
          # webkit2_41 tells Wails to use webkit2gtk-4.1 pkg-config name;
          # webkitgtk_4_0 was removed from nixpkgs 26.05.
          tags = [
            "desktop"
            "production"
          ]
          ++ lib.optionals (builtins.elem system [
            "x86_64-linux"
            "aarch64-linux"
          ]) [ "webkit2_41" ];
          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${v}"
          ];
          doCheck = false;
        };

      # Cross-compile the capui desktop app to Windows from any Linux host.
      # go-webview2 loads WebView2 via pure Go syscall (no CGO required), so the
      # standard Go GOOS/GOARCH mechanism is all that is needed — no MinGW toolchain.
      capuiWindowsApp =
        { pkgs, system }:
        pkgs.buildGoModule {
          pname = "capui";
          version = v;
          src = gitignore.lib.gitignoreSource ./.;
          inherit vendorHash;
          subPackages = [ "cmd/capui" ];
          # Skip the standard buildPhase; go install cross-targets output to
          # $GOPATH/bin/GOOS_GOARCH/ which buildGoModule does not copy to $out.
          # The installPhase sets GOOS/GOARCH/CGO_ENABLED and builds in one step.
          buildPhase = "true";
          installPhase = ''
            runHook preInstall
            export GOOS=windows
            export GOARCH=amd64
            export CGO_ENABLED=0
            mkdir -p $out/bin
            go build -tags desktop,production \
              -ldflags "-s -w -H=windowsgui -X main.Version=${v}" \
              -o $out/bin/capui.exe \
              ./cmd/capui
            runHook postInstall
          '';
          doCheck = false;
          meta.platforms = lib.platforms.all;
        };

      # Development tools used.
      devTools = pkgs: [
        pkgs.adr-tools
        pkgs.gh
        pkgs.git
        pkgs.go
        pkgs.goreleaser
        pkgs.govulncheck
        pkgs.templ
        pkgs.version
        pkgs.wails
        pkgs.xc
      ];

      # CGO build inputs required by the Wails WebView per platform.
      # Windows uses WebView2 loaded via pure Go syscall (no CGO inputs needed).
      # macOS uses the system SDK frameworks via Xcode CLT; no explicit Nix inputs needed.
      webviewCGOInputs =
        system: pkgs:
        if
          builtins.elem system [
            "x86_64-linux"
            "aarch64-linux"
          ]
        then
          [
            pkgs.pkg-config
            pkgs.gtk3
            pkgs.webkitgtk_4_1
          ]
        else
          [ ];

      name = "cap";
    in
    {
      # `nix build` builds the cap CLI.
      # `nix build ".#capui"` builds the desktop app for the current platform.
      # `nix build ".#capui-windows"` cross-compiles the desktop app for Windows (Linux hosts only).
      packages = forAllSystems (
        { system, pkgs }: rec {
          default = app {
            name = name;
            pkgs = pkgs;
            system = system;
          };
          cap = default;
          capui = capuiApp { inherit pkgs system; };
          capui-windows =
            if
              builtins.elem system [
                "x86_64-linux"
                "aarch64-linux"
              ]
            then
              capuiWindowsApp { inherit pkgs system; }
            else
              throw "capui-windows can only be built on Linux";
        }
      );

      # `nix develop` provides a shell containing required tools.
      devShells = forAllSystems (
        { system, pkgs }: {
          default = pkgs.mkShell {
            buildInputs =
              (devTools pkgs)
              ++ (webviewCGOInputs system pkgs)
              ++ lib.optionals (builtins.elem system [
                "x86_64-linux"
                "aarch64-linux"
              ]) [ nixgl.packages.${system}.nixGLIntel ];
          };
        }
      );
    };
}
