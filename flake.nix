{
  description = "🔥 horizontally-scalable, highly-available, multi-tenant continuous profiling aggregation system";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in
      {
        devShell = with pkgs; mkShell {
          buildInputs = [
            go_1_18
            golangci-lint
            gopls
            buf
          ];

          shellHook = ''
            export GOROOT=${pkgs.go_1_18}/share/go
            export "PS1=\n\[\033[1;32m\][\[\e]0;\u@\h: \w\a\]\u@\h:\w]🔥\[\033[0m\] "
          '';

        };
      });
}
