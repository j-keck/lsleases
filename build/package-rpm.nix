{pkgs, lsleases, arch}:
pkgs.stdenv.mkDerivation rec {
  name = "package-rpm";
  src = "";
  pkg = lsleases {inherit arch;};

  version = pkg.version;

  pkgVersion = builtins.replaceStrings ["-"] ["."] version;
  target = if arch == "i386"
              then "i386"
              else "x86_64";

  buildInputs = with pkgs; [ rpm ];
  maintainer = let j-keck = pkgs.stdenv.lib.maintainers.j-keck;
               in j-keck.name + " <" + j-keck.email + ">";

  buildCommand = ''

    cp -v "${pkgs.writeText "lsleases.spec" ''
      %define name lsleases
      %define version ${pkgVersion}
      %define release 1
      %define _build_id_links none
      %define _tmppath %(echo $out)


      Name: %{name}
      Version: %{version}
      Release: %{release}
      Summary: dhcp leases sniffer
      License: MIT
      URL: http://github.com/j-keck/lsleases
      AutoReqProv: no

      %description
      dhcp leases sniffer


      %files
      /etc/systemd/system/lsleasesd.service
      /usr/bin/lsleases
      /usr/bin/lsleasesd
      /usr/local/man/man1/lsleases.1.gz

      %post
      /usr/sbin/setcap cap_net_raw,cap_net_bind_service=+ep /usr/bin/lsleasesd

      %postun
      if [ -d /var/cache/lsleasesd ]; then rm -rf /var/cache/lsleasesd; fi
      if [ -d /var/run/lsleasesd ]; then rm -rf /var/run/lsleasesd; fi

    ''}" lsleases.spec;


    mkdir -p pkg/usr/bin
    cp ${pkg}/bin/lsleases  pkg/usr/bin/lsleases
    cp ${pkg}/bin/lsleasesd pkg/usr/bin/lsleasesd


    mkdir -p pkg/usr/local/man/man1
    cp ${pkg}/share/man/man1/lsleases.1.gz pkg/usr/local/man/man1

    mkdir -p pkg/etc/systemd/system
    cp "${pkgs.writeScript "lsleasesd.service" ''
      [Unit]
      Description=dhcp leases sniffer
      After=network.target

      [Service]
      Type=simple
      PermissionsStartOnly=true
      ExecStartPre=-/usr/bin/mkdir /var/run/lsleasesd
      ExecStartPre=/usr/bin/chown nobody:nobody /var/run/lsleasesd
      ExecStart=/usr/bin/lsleasesd -k
      ExecStop=/usr/bin/lsleases -x
      Restart=on-failure
      User=nobody
      Group=nobody

      [Install]
      WantedBy=multi-user.target
    ''}" pkg/etc/systemd/system/lsleasesd.service


    # build the rpm package
    HOME=$PWD rpmbuild -bb --buildroot=$PWD/pkg --target ${target} --nodeps lsleases.spec


    mkdir -p $out
    cp -v $PWD/rpmbuild/RPMS/${target}/lsleases-${pkgVersion}-1.${target}.rpm $out/lsleases-${version}-${arch}.rpm
  '';
}
