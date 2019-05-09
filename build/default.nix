let

  fetchNixpkgs = {rev, sha256}: builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/${rev}.tar.gz";
    inherit sha256;
  };

  nixpkgs = fetchNixpkgs {
    # nixpkgs-unstable of 2019-03-23T11:23:23+01:00
    rev = "796a8764ab85746f916e2cc8f6a9a5fc6d4d03ac";
    sha256 = "1m57gsr9r96gip2wdvdzbkj8zxf47rg3lrz35yi352x1mzj3by3x";
  };

  pkgs = import nixpkgs {};

  version =
    let mkVersion = pkgs.stdenv.mkDerivation {
      src = ../.;
      name = "mkVersion";
      phases = "buildPhase";
      buildPhase = ''
        mkdir -p $out
        ${pkgs.git}/bin/git -C $src describe --always --tags > $out/version
      '';
    };
    in pkgs.lib.removeSuffix "\n" (builtins.readFile "${mkVersion}/version");


in with pkgs; rec {

  manpage = stdenv.mkDerivation rec {
    inherit version;
    name = "manpage";
    src = ../docs;
    phases = "buildPhase";
    buildPhase = ''
      mkdir -p $out

      # replace all @VARIABLES@ with their values from the environment
      substituteAll $src/lsleases.org $out/lsleases.org
      substituteAll $src/lsleasesd.org $out/lsleasesd.org

      # create man pages
      ${pandoc}/bin/pandoc -s -o $out/lsleases.1 $out/lsleases.org
      ${pandoc}/bin/pandoc -s -o $out/lsleasesd.1 $out/lsleasesd.org
    '';
  };


  lsleases = {arch ? "amd64"}:
    let goModule = if arch == "i386" then pkgsi686Linux.buildGoModule else pkgs.buildGoModule; in
    goModule rec {
      inherit version;
      pname = "lsleases";
      rev = "v${version}";
      src = lib.cleanSource ../.;

      CGO_ENABLED = 0;
      buildFlagsArray = ''
        -ldflags=
        -X main.version=${version}
        -X github.com/j-keck/lsleases/pkg/daemon.version=${version}
      '';

      modSha256 = "sha256:07srf10i1n8fxjgxc8hj8iirsmkbl8pv07q6n93pb1q0ibsxqj21";

      installPhase  = ''
        mkdir -p $out/bin
        cp $GOPATH/bin/lsleases  $out/bin
        cp $GOPATH/bin/lsleasesd $out/bin

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



  package-deb = {arch ? "amd64"}: import ./package-deb.nix {inherit pkgs lsleases arch; };
  package-deb-test = import ./package-deb-test.nix { inherit pkgs package-deb; };

  package-rpm = {arch ? "amd64"}: import ./package-rpm.nix { inherit pkgs lsleases arch; };
}

