%define name lsleases
%define version 1.0
%define release 1
%define srcdir %(echo $BUILD_DIR)
%define _rpmdir %(echo $PACKAGE_DIR)

Name: %{name}
Version: %{version}
Release: %{release}
Summary: dhcp leases sniffer
License: MIT
URL: http://github.com/j-keck/lsleases
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}
AutoReqProv: no

%description
dhcp leases sniffer


%prep
rm -rf $RPM_BUILD_DIR/%{name}
mkdir -p $RPM_BUILD_DIR/%{name}
git clone %{srcdir} $RPM_BUILD_DIR/%{name}


%build
cd $RPM_BUILD_DIR/%{name}

%ifarch i386
export GOARCH=386
%else
export GOARCH=%{_target_cpu}
%endif

go build -v 
pandoc -s -t man MANUAL.md -o lsleases.1

%install
rm -rf $RPM_BUILD_ROOT

mkdir -p $RPM_BUILD_ROOT/usr/bin
install -m 0755 $RPM_BUILD_DIR/%{name}/%{name} $RPM_BUILD_ROOT/usr/bin

mkdir -p $RPM_BUILD_ROOT/etc/init.d
install -m 0755 %{srcdir}/rpmbuild/init.d/%{name} $RPM_BUILD_ROOT/etc/init.d

mkdir -p $RPM_BUILD_ROOT/%{_mandir}/man1
install -m 0755 $RPM_BUILD_DIR/%{name}/%{name}.1 $RPM_BUILD_ROOT/%{_mandir}/man1


%files
/etc/init.d/%{name}
%{_bindir}/%{name}
%{_mandir}/man1/%{name}.1.gz


%post
/usr/sbin/setcap cap_net_raw,cap_net_bind_service=+ep %{_bindir}/%{name}

echo "enable autostart ..."
/sbin/chkconfig --add %{name}
/sbin/chkconfig %{name} on

echo "startup server ..."
/sbin/service %{name} start

