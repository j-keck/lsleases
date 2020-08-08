let

  fetchNixpkgs = {rev, sha256}: builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs-channels/archive/${rev}.tar.gz";
    inherit sha256;
  };

  nixpkgs = fetchNixpkgs {
    # nixos-20.03 of 08.08.2020
    rev = "4364ff933ebec0ef856912b182f4f9272aa7f98f";
    sha256 = "19ig1ywd2jq7qqzwpw6f1li90dq4kk3v0pbqgn6lzdabzf95bz6z";
  };

  pkgs = import nixpkgs {};

  version =
    let mkVersion = pkgs.stdenv.mkDerivation {
      src = ./.;
      name = "mkVersion";
      phases = "buildPhase";
      buildPhase = ''
        mkdir -p $out
        ${pkgs.git}/bin/git -C $src describe --always --tags > $out/version
      '';
    };
    in pkgs.lib.removeSuffix "\n" (builtins.readFile "${mkVersion}/version");

  manpage = pkgs.stdenv.mkDerivation rec {
    inherit version;
    name = "manpage";
    src = ./docs;
    phases = "buildPhase";
    buildPhase = ''
      mkdir -p $out

      # replace all @VARIABLES@ with their values from the environment
      substituteAll $src/lsleases.org $out/lsleases.org
      substituteAll $src/lsleasesd.org $out/lsleasesd.org

      # create man pages
      ${pkgs.pandoc}/bin/pandoc -s -o $out/lsleases.1 $out/lsleases.org
      ${pkgs.pandoc}/bin/pandoc -s -o $out/lsleasesd.1 $out/lsleasesd.org
    '';
  };


  lsleases = {arch ? "amd64", goos ? "linux" }:
    let goModule = if arch == "i386" then pkgs.pkgsi686Linux.buildGoModule else pkgs.buildGoModule; in
    goModule rec {
      inherit version goos;
      pname = "lsleases";
      rev = "v${version}";
      src = pkgs.lib.cleanSource ./.;

      CGO_ENABLED = 0;
      buildFlagsArray = ''
        -ldflags=
        -X main.version=${version}
        -X github.com/j-keck/lsleases/pkg/daemon.version=${version}
      '';

      modSha256 = "sha256:1yjrdl73yilyg9vp7khqwjc5li88frc602ir3vb0xl3apbv9z0km";

      preBuild = ''
        export GOOS=${goos}
      '';

      installPhase  = ''
        BIN_PATH=${if goos == pkgs.stdenv.buildPlatform.parsed.kernel.name
                   then "$GOPATH/bin"
                   else "$GOPATH/bin/${goos}_$GOARCH"}

        mkdir -p $out/bin
        cp $BIN_PATH/lsleases  $out/bin
        cp $BIN_PATH/lsleasesd $out/bin

        mkdir -p $out/man/man1
        cp ${manpage}/lsleases.1 $out/man/man1
        cp ${manpage}/lsleasesd.1 $out/man/man1
      '';

      meta = with pkgs.stdenv.lib; {
        description = "DHCP leases sniffer";
        homepage = "https://github.com/j-keck/lsleases";
        license = licenses.mit;
        maintainers = maintainers.j-keck;
      };
  };

in rec {

  inherit lsleases;

  package-deb = {arch ? "amd64"}: import ./build/package-deb.nix {inherit pkgs lsleases arch; };
  package-deb-test = import ./build/package-deb-test.nix { inherit pkgs package-deb; };

  package-rpm = {arch ? "amd64"}: import ./build/package-rpm.nix { inherit pkgs lsleases arch; };

  package-osx = { arch ? "amd64"}: import ./build/package-osx.nix { inherit pkgs lsleases arch; };
}

