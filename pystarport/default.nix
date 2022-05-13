{ pkgs ? import <nixpkgs> {} }:
let chain = import ../. { inherit pkgs; };
in
  pkgs.poetry2nix.mkPoetryApplication {
    projectDir = ./.;
    preBuild = ''
    sed -i -e "s@CHAIN = 'astrad'  # edit by nix-build@CHAIN = '${chain}/bin/astrad'@" pystarport/cluster.py
    '';
  }