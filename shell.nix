{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.pkg-config
    pkgs.libusb1
    pkgs.libftdi1
    pkgs.bluez
  ];
}
