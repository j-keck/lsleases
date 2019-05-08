{ pkgs, package-deb }:
pkgs.dockerTools.buildImage {
    name = "lsleases-package-deb-test";
    tag = "latest";

    fromImage = pkgs.dockerTools.pullImage {
      imageName = "debian";
      sha256 = "1lqk2ab4mn255plaixshcbyqb54zm87zymv4376hpbx6mqf8ajz3";

      # nix-shell --packages skopeo jq --command "skopeo  inspect docker://docker.io/debian | jq -r '.Digest'"
      imageDigest = "sha256:724b0fbbda7fda6372ffed586670573c59e07a48c86d606bab05db118abe0ef5";
    };

    contents = [ (package-deb {}) pkgs.libcap_progs ];

    runAsRoot = ''
      #!${pkgs.stdenv.shell}

      cp "${pkgs.writeScript "send-dummy-broadcast.pl" ''
        #!${pkgs.perl}/bin/perl
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


      cp "${pkgs.writeScript "test.sh" ''
        #!${pkgs.stdenv.shell}
        dpkg -i /lsleases-${(package-deb {}).version}.deb
        lsleasesd &
        sleep 1
        lsleases -V
        lsleases
        /send-dummy-broadcast.pl
        lsleases
      ''}" /test.sh

     # fix "fatal: unable to access 'https://github.com/...': SSL certificate problem: unable to get local issuer certificate"
     #      git config --system http.sslCAInfo ${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt
    '';

    config.Cmd = [ "/test.sh" ];

}
