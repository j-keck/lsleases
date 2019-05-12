{ pkgs, arch, lsleases }:
pkgs.stdenv.mkDerivation rec {
  name = "package-osx";
  src = pkgs.lib.cleanSource ../.;
  pkg = lsleases { inherit arch; goos = "darwin"; };
  version = pkg.version;

  buildInputs = with pkgs; [ pandoc zip ];

  buildCommand = with pkgs; ''
    mkdir lsleases

    cp "${writeScript "capture-leases.sh" ''
      #!/bin/sh
      BASE_PATH=$(dirname $0)
      echo "start sniffer to capture ip addresses ..."
      $BASE_PATH/lsleases -s
    ''}" lsleases/capture-leases.sh

    substituteAll $src/docs/lsleases.org lsleases.org
    ${pandoc}/bin/pandoc -s --toc --toc-depth=1 lsleases.org -o lsleases/lsleases.html

    substituteAll $src/docs/manual.org manual.org
    ${pandoc}/bin/pandoc -s --toc --toc-depth=1 manual.org -o lsleases/manual.html

    cp ${pkg}/bin/lsleases lsleases/

    mkdir -p $out
    zip -r $out/lsleases-${version}-${arch}-osx-standalone.zip lsleases
  '';
}
