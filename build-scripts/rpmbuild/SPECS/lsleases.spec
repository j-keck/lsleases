%define name lsleases
%define version %(echo $VERSION)
%define release 1
%define srcdir %(echo $BUILD_DIR)
%define _rpmdir %(echo $PACKAGE_DIR)
%define gopath $RPM_BUILD_DIR/%{name}/gobuild

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

mkdir -p %{gopath}
export GOPATH=%{gopath}

git clone %{srcdir} %{gopath}/src/%{name}
cd %{gopath}/src/%{name}

go get -v -d


%build
cd %{gopath}/src/%{name}
export GOPATH=%{gopath}

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
install -m 0755 %{gopath}/src/%{name}/%{name} $RPM_BUILD_ROOT/usr/bin

mkdir -p %{buildroot}/%{_libdir}/systemd/system
install %{srcdir}/rpmbuild/%{name}.service %{buildroot}/%{_libdir}/systemd/system

mkdir -p $RPM_BUILD_ROOT/etc/init.d
install -m 0755 %{srcdir}/rpmbuild/init.d/%{name} $RPM_BUILD_ROOT/etc/init.d

mkdir -p $RPM_BUILD_ROOT/%{_mandir}/man1
install -m 0755 %{gopath}/src/%{name}/%{name}.1 $RPM_BUILD_ROOT/%{_mandir}/man1


%files
/etc/init.d/%{name}
%{_bindir}/%{name}
%{_libdir}/systemd/system/%{name}.service
%{_mandir}/man1/%{name}.1.gz


%post
/usr/sbin/setcap cap_net_raw,cap_net_bind_service=+ep %{_bindir}/%{name}

echo "enable autostart ..."
/sbin/chkconfig --add %{name}
/sbin/chkconfig %{name} on

echo "startup server ..."
/sbin/service %{name} start

