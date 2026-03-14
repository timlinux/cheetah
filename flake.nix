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

        # Script to record demo with asciinema
        demo-record = pkgs.writeShellScriptBin "cheetah-demo-record" ''
          #!/usr/bin/env bash
          set -e

          # Use current working directory (should be project root)
          PROJECT_DIR="$(pwd)"
          DEMO_DIR="$PROJECT_DIR/demo"
          CAST_FILE="$DEMO_DIR/cheetah-demo.cast"
          GIF_FILE="$DEMO_DIR/cheetah-demo.gif"

          # Verify we're in the right directory
          if [ ! -f "$PROJECT_DIR/flake.nix" ]; then
            echo "❌ Please run this command from the cheetah project root directory."
            exit 1
          fi

          mkdir -p "$DEMO_DIR"

          echo "🎬 Recording Cheetah demo..."
          echo ""
          echo "Tips for a good demo:"
          echo "  - Open cheetah and load a document"
          echo "  - Show speed controls (j/k keys)"
          echo "  - Demonstrate pause/resume (space)"
          echo "  - Show paragraph navigation (h/l)"
          echo "  - Keep it under 30 seconds"
          echo ""
          echo "Press Enter to start recording, type 'exit' when done..."
          read -r

          ${pkgs.asciinema}/bin/asciinema rec --overwrite "$CAST_FILE"

          echo ""
          echo "✅ Recording saved to $CAST_FILE"
          echo ""
          echo "Converting to GIF for README..."

          ${pkgs.asciinema-agg}/bin/agg --theme monokai "$CAST_FILE" "$GIF_FILE"
          echo "✅ GIF saved to $GIF_FILE"

          # Copy GIF to hugo static folder
          mkdir -p "$PROJECT_DIR/hugo/static/images"
          cp "$GIF_FILE" "$PROJECT_DIR/hugo/static/images/cheetah-demo.gif"
          echo "✅ GIF copied to hugo/static/images/"

          echo ""
          echo "🎉 Demo recording complete!"
          echo ""
          echo "The demo is now available at:"
          echo "  - demo/cheetah-demo.cast (asciinema format)"
          echo "  - demo/cheetah-demo.gif (animated GIF)"
          echo "  - hugo/static/images/cheetah-demo.gif (for docs)"
          echo ""
          echo "README.md and docs will automatically use the new demo."
        '';

        # Script to play demo locally
        demo-play = pkgs.writeShellScriptBin "cheetah-demo-play" ''
          #!/usr/bin/env bash
          PROJECT_DIR="$(pwd)"
          CAST_FILE="$PROJECT_DIR/demo/cheetah-demo.cast"

          if [ ! -f "$CAST_FILE" ]; then
            echo "❌ No demo recording found at $CAST_FILE"
            echo "Run 'nix run .#demo-record' to create one."
            exit 1
          fi

          echo "🎬 Playing Cheetah demo..."
          ${pkgs.asciinema}/bin/asciinema play "$CAST_FILE"
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
          demo-record = demo-record;
          demo-play = demo-play;
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

          # nix run .#demo-record - Record a demo with asciinema
          demo-record = {
            type = "app";
            program = "${demo-record}/bin/cheetah-demo-record";
          };

          # nix run .#demo-play - Play the demo locally
          demo-play = {
            type = "app";
            program = "${demo-play}/bin/cheetah-demo-play";
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
            asciinema
            asciinema-agg
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
            echo "  nix run .#docs-serve  - Start Hugo dev server"
            echo "  nix run .#docs-build  - Build documentation"
            echo "  nix run .#demo-record - Record demo with asciinema"
            echo "  nix run .#demo-play   - Play recorded demo"
            echo ""
          '';
        };
      }
    );
}
