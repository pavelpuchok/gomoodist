{
  description = "Gomoodist";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      packages.${system}.default = pkgs.buildGoModule {
        pname = "gomoodist";
        version = "0.1.0";
        src = ./.;

        vendorHash = "sha256-9jCDXglIJhKifxYOhlLdO+AqjhUUUoL4cRHO6DPX42U=";

        nativeBuildInputs = with pkgs; [
          gcc
          pkg-config
        ];

        buildInputs = with pkgs; [
          alsa-lib.dev
          gtk3.dev
          libayatana-appindicator.dev
        ];

        subPackages = [ "cmd/gomoodist" ];
      };
    };
}
