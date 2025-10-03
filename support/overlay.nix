{ pkgs
}: (final: prev: {
  protoc-states = pkgs.callPackage ./protoc-states.nix { };
})