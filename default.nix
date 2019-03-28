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

in with (import nixpkgs {}); rec {

  manpage = stdenv.mkDerivation rec {
    name = "manpage";
    src = ./MANUAL.md;
    phases = "buildPhase";
    buildPhase = ''
      mkdir -p $out
      ${pandoc}/bin/pandoc -s -o $out/lsleases.1 $src
    '';
  };

  lsleases = {arch ? "amd64"}:
    let goModule = if arch == "i386" then pkgsi686Linux.buildGoModule else pkgs.buildGoModule; in
    goModule rec {
      pname = "lsleases";
      version = "1.5.0";
      rev = "v${version}";
      src = pkgs.lib.cleanSource ./.;

      CGO_ENABLED = 0;
      goPackagePath = "github.com/j-keck/lsleases";
      buildFlagsArray = ''
        -ldflags=
        -X main.VERSION=${version}
      '';

      modSha256 = "0w83mbl3mqlg545gx79450h6knvkwnzcyglyjrmvy7xdw4w8lqzz";

      installPhase  = ''
        mkdir -p $out/bin
        cp $GOPATH/bin/lsleases $out/bin

        mkdir -p $out/man/man1
        cp ${manpage}/lsleases.1 $out/man/man1
      '';

      meta = with pkgs.stdenv.lib; {
        description = "DHCP leases sniffer";
        homepage = "https://github.com/j-keck/lsleases";
        license = licenses.mit;
        maintainers = maintainers.j-keck;
      };
  };



  package-dpkg = {arch ? "amd64"}: stdenv.mkDerivation rec {
    name = "package-dpkg";
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

      cp "${writeScript "postinst" ''
        mkdir -p /var/lib/lsleases && chown nobody:nogroup /var/lib/lsleases

        setcap cap_net_raw,cap_net_bind_service=+ep /usr/bin/lsleases
      ''}" DEBIAN/postinst

      cp "${writeScript "postrm" ''
        if [ -d /var/lib/lsleases ]; then rm -rf /var/lib/lsleases fi
      ''}" DEBIAN/postrm

      mkdir -p usr/bin
      cp ${pkg}/bin/lsleases usr/bin/lsleases

      mkdir -p usr/local/man/man1
      cp ${pkg}/share/man/man1/lsleases.1.gz usr/local/man/man1

      mkdir -p etc/systemd/system
      cp "${writeScript "lsleases.service" ''
        [Unit]
        Description=dhcp leases sniffer
        After=network.target

        [Service]
        Type=simple
        ExecStart=/usr/bin/lsleases -s -k
        ExecStop=/usr/bin/lsleases -x
        Restart=on-failure
        User=nobody
        Group=nogroup

        [Install]
        WantedBy=multi-user.target
      ''}" etc/systemd/system/lsleases.service

      mkdir -p $out
      fakeroot dpkg -b . $out/lsleases-${pkg.version}.deb
    '';
  };



  package-dpkg-test = pkgs.dockerTools.buildImage {
    name = "lsleases-package-dpkg-test";
    tag = "latest";

    fromImage = pkgs.dockerTools.pullImage {
      imageName = "debian";
      sha256 = "1lqk2ab4mn255plaixshcbyqb54zm87zymv4376hpbx6mqf8ajz3";

      # nix-shell --packages skopeo jq --command "skopeo  inspect docker://docker.io/debian | jq -r '.Digest'"
      imageDigest = "sha256:724b0fbbda7fda6372ffed586670573c59e07a48c86d606bab05db118abe0ef5";
    };

    contents = [ (package-dpkg {}) libcap_progs ];

    runAsRoot = ''
      #!${stdenv.shell}

      cp "${writeScript "send-dummy-broadcast.pl" ''
        #!${perl}/bin/perl
        use Socket;

        my $dummy_host_name = "dummy-test-host";
        my $host_name_option = sprintf("%02x%s", length($dummy_host_name), unpack("H*", $dummy_host_name));
        my $dummy_dhcp_request = "010106002ccab3380000000000000000000000000000000000000000080027f2975a" .
        "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
        "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
        "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
        "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" .
        "638253633501033204". "01010101" ."0c".$host_name_option."3712011c02030f06770c2c2f1a792a79f921fc2aff0000000" .
        "000000000000000000000000000";


        socket(SOCKET, AF_INET, SOCK_DGRAM, getprotobyname("udp")) or die "Couldn't create raw socket: $!";
        my $broadcast = sockaddr_in(67, INADDR_BROADCAST);
        setsockopt(SOCKET, SOL_SOCKET, SO_BROADCAST, 1);
        send(SOCKET, pack("H*", $dummy_dhcp_request), 0, $broadcast);
        close SOCKET;
      ''}" /send-dummy-broadcast.pl


      cp "${writeScript "test.sh" ''
        #!${stdenv.shell}
        dpkg -i /lsleases-${(package-dpkg {}).version}.deb
        lsleases -s &
        lsleases -V
        lsleases
        /send-dummy-broadcast.pl
        lsleases
      ''}" /test.sh

     # fix "fatal: unable to access 'https://github.com/...': SSL certificate problem: unable to get local issuer certificate"
     #      git config --system http.sslCAInfo ${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt
    '';

    config.Cmd = [ "/test.sh" ];

  };
}

