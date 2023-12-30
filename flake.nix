{
  description = "A Go SSH client";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... } @ inputs:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
	  overlays = [
	    (self: super: {
	      GoSSH = self.packages.${system}.GoSSH;
	    })
	  ];
        };
      in
      {
        defaultPackage = self.packages.${system}.GoSSH;

        devShells.default = pkgs.mkShell {
 	  buildInputs = [ pkgs.go ];
  	  shellHook = ''
    	    export PATH=$PATH:${self.packages.${system}.GoSSH}/bin
	  '';
	};

	packages.GoSSH = pkgs.buildGoModule {
          pname = "GoSSH";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
        };
      }
    );
}

