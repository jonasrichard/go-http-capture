let
    pkgs = import <nixpkgs> { };
in
{
    capture = pkgs.callPackage ./flake.nix { };
}
