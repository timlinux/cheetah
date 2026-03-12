# SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
# SPDX-License-Identifier: MIT
{
  description = "Cheetah - RSVP Speed Reading Application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Hugo documentation site
        docs = pkgs.stdenv.mkDerivation {
          pname = "cheetah-docs";
          version = "0.1.0";
          src = ./hugo;

          nativeBuildInputs = [ pkgs.hugo ];

          buildPhase = ''
            hugo --minify
          '';

          installPhase = ''
            cp -r public $out
          '';
        };

        # Script to run Hugo dev server
        docs-serve = pkgs.writeShellScriptBin "cheetah-docs-serve" ''
          cd ${toString ./hugo}
          ${pkgs.hugo}/bin/hugo server -D --bind 0.0.0.0 --port 1313
        '';

        # Script to build docs
        docs-build = pkgs.writeShellScriptBin "cheetah-docs-build" ''
          cd ${toString ./hugo}
          ${pkgs.hugo}/bin/hugo --minify
          echo "Documentation built in hugo/public/"
        '';

        # Script to open docs in browser
        docs-open = pkgs.writeShellScriptBin "cheetah-docs-open" ''
          ${pkgs.xdg-utils}/bin/xdg-open http://localhost:1313 2>/dev/null || \
          open http://localhost:1313 2>/dev/null || \
          echo "Open http://localhost:1313 in your browser"
        '';

      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "cheetah";
            version = "0.3.0";
            src = ./.;
            vendorHash = "sha256-6xoFDUOvYBku2t+O+8U9cSlyk9nkpNiaKw4aJ2HXsw8=";

            meta = with pkgs.lib; {
              description = "RSVP Speed Reading Application - read at 1000+ WPM";
              homepage = "https://github.com/timlinux/cheetah";
              license = licenses.mit;
              maintainers = [ ];
              mainProgram = "cheetah";
            };
          };

          # Documentation packages
          docs = docs;
          docs-serve = docs-serve;
          docs-build = docs-build;
          docs-open = docs-open;
        };

        # Apps for `nix run`
        apps = {
          default = {
            type = "app";
            program = "${self.packages.${system}.default}/bin/cheetah";
          };

          # nix run .#docs-serve - Start Hugo dev server
          docs-serve = {
            type = "app";
            program = "${docs-serve}/bin/cheetah-docs-serve";
          };

          # nix run .#docs-build - Build documentation
          docs-build = {
            type = "app";
            program = "${docs-build}/bin/cheetah-docs-build";
          };

          # nix run .#docs-open - Open docs in browser
          docs-open = {
            type = "app";
            program = "${docs-open}/bin/cheetah-docs-open";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            golangci-lint
            hugo
            nodejs_20
            pre-commit
          ];

          shellHook = ''
            echo "🐆 Cheetah development environment"
            echo ""
            echo "Commands:"
            echo "  make build      - Build Go binary"
            echo "  make run        - Run the app"
            echo "  make test       - Run tests"
            echo "  make docs-dev   - Start Hugo dev server"
            echo "  make docs-build - Build documentation"
            echo ""
            echo "Nix run commands:"
            echo "  nix run .#docs-serve - Start Hugo dev server"
            echo "  nix run .#docs-build - Build documentation"
            echo ""
          '';
        };
      }
    );
}
