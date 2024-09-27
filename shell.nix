{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.chromedriver
  ];
  shellHook = ''
  '';
}
