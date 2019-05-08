{pkgs, lsleases, arch}:
pkgs.stdenv.mkDerivation rec {
    name = "package-deb";
    src = "";
    pkg = lsleases {inherit arch;};
    version = pkg.version;
    buildInputs = with pkgs; [ dpkg fakeroot];
    maintainer = let j-keck = pkgs.stdenv.lib.maintainers.j-keck;
                  in j-keck.name + " <" + j-keck.email + ">";

    buildCommand = ''
      mkdir -p DEBIAN

      cp "${pkgs.writeText "control" ''
        Package: lsleases
        Version: ${pkg.version}
        Priority: optional
        Architecture: ${arch}
        Section: custom
        Essential: no
        Depends: libcap2-bin
        Maintainer: ${maintainer}
        Description: ${pkg.meta.description}
      ''}" DEBIAN/control

      cp "${pkgs.writeScript "postinst" ''
        mkdir -p /var/cache/lsleases && chown nobody:nogroup /var/cache/lsleases

        setcap cap_net_raw,cap_net_bind_service=+ep /usr/bin/lsleasesd
      ''}" DEBIAN/postinst

      cp "${pkgs.writeScript "postrm" ''
        if [ -d /var/cache/lsleases ]; then rm -rf /var/cache/lsleases; fi
        if [ -d /var/run/lsleases ]; then rm -rf /var/run/lsleases; fi
      ''}" DEBIAN/postrm

      mkdir -p usr/bin
      cp ${pkg}/bin/lsleases  usr/bin/lsleases
      cp ${pkg}/bin/lsleasesd usr/bin/lsleasesd

      mkdir -p usr/local/man/man1
      cp ${pkg}/share/man/man1/lsleases.1.gz usr/local/man/man1

      mkdir -p etc/systemd/system
      cp "${pkgs.writeScript "lsleases.service" ''
        [Unit]
        Description=dhcp leases sniffer
        After=network.target

        [Service]
        Type=simple
        PermissionsStartOnly=true
        ExecStartPre=-/bin/mkdir /var/run/lsleases
        ExecStartPre=/bin/chown nobody:nogroup /var/run/lsleases
        ExecStart=/usr/bin/lsleasesd -k
        ExecStop=/usr/bin/lsleases -x
        Restart=on-failure
        User=nobody
        Group=nogroup

        [Install]
        WantedBy=multi-user.target
      ''}" etc/systemd/system/lsleases.service

      mkdir -p $out
      fakeroot dpkg -b . $out/lsleases-${pkg.version}-${arch}.deb
    '';
}
