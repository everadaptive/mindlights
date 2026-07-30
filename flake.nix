{
  description = "mindlights dev environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      devShells.${system} = {
        default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.pkg-config
            pkgs.libusb1
            pkgs.libftdi1
            pkgs.bluez
          ];
        };

        web = pkgs.mkShell {
          buildInputs = [
            pkgs.nodejs_22
          ];
          shellHook = ''
            cd cmd/stream/app
            echo "Run: npm install && npm start"
          '';
        };
      };

      apps.${system}.web = {
        type = "app";
        program = toString (pkgs.writeShellScript "run-web-ui" ''
          set -e
          export PATH="${pkgs.nodejs_22}/bin:$PATH"
          cd "$PWD/cmd/stream/app"
          if [ ! -d node_modules ]; then
            npm install --legacy-peer-deps
          fi
          exec npm start
        '');
      };
    };
}
