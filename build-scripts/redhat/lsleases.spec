%define name lsleases
%define version %(echo $VERSION)
%define release 1
%define _rpmdir %(echo $BUILD_OUTPUT)


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
/etc/init.d/%{name}
%{_bindir}/%{name}
%{_mandir}/man1/%{name}.1



%post
/usr/sbin/setcap cap_net_raw,cap_net_bind_service=+ep %{_bindir}/%{name}

echo "enable autostart ..."
/sbin/chkconfig --add %{name}
/sbin/chkconfig %{name} on

echo "startup server ..."
/sbin/service %{name} start

