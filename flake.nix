{ go, lib, libpcap, stdenv }:
let
    fs = lib.fileset;
    sourceFiles = fs.gitTracked ./.;
in

fs.trace sourceFiles

stdenv.mkDerivation {
    name = "capture";
    src = fs.toSource {
        root = ./.;
        fileset = sourceFiles;
    };
    buildInputs = [ go libpcap ];

    buildPhase = ''
    HOME=$TMPDIR go build -o capture main.go
    '';

    installPhase = ''
    mkdir -p $out/bin
    cp -r capture $out/bin/capture
    '';
}

#{
#    inputs = {
#        nixpkgs.url = "nixpkgs/nixos-unstable"
#
#        gomod2nix = {
#            url = "github:tweag/gomod2nix"
#            inputs.nixpkgs.follows = "nixpkgs"
#        };
#    };
#
#    outputs = { self, nixpkgs, gomod2nix }@attrs:
#        utils.lib.eachSystem [
#            "x86_64-linux"
#            "aarch64-darwin"
#        ]
#    };
#}
