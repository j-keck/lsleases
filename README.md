# lsleases - dhcp leases sniffer

lsleases captures broadcast 'DHCP Request' datagrams and displays the ip, mac and host name from computers in the local network with dynamic addresses.


## Installation

### From source

  1. export GOPATH: `export GOPATH=<PATH>/lsleases-build`
  2. download and build: `go get github.com/j-keck/lsleases`. This will build the binary in `$GOPATH/bin`
  3. copy the binary to your preferred bin dir `cp $GOPATH/bin/lsleases $HOME/bin/`

  **Linux**
  4. to allow non-root users to open port less than 1024 (dhcp sniffer) and use raw sockets (arping) set the corresponding capabilities
  
     `setcap 'cap_net_raw,cap_net_bind_service+ep' $HOME/bin/lsleases`

  **FreeBSD**
  4. allow non-root users to open port less than 1024 (dhcp sniffer)
  
        echo net.inet.ip.portrange.reservedhigh=0 >> /etc/sysctl.conf
        service sysctl restart

  *arping as non-root under FreeBSD not supported*  

  **Windows**
  *no additional steps necessary*


  
### Binary packages
  1. download from http://github.com/j-keck/lsleases/releases/latest

  *deb packages register and starts an server instance on installation*

  
## Usage

1. start an server instance if not installed from package `j@main:~> nohup lsleases -s &`
2. replug / startup any device with dynamic addresses
3. display captured ip, mac and host names. 

        j@main:~> lsleases
        Ip               Mac                Name
        192.168.1.189    10:bf:48:xx:xx:xx  android-f6c6dca2130b287
        192.168.1.122    b8:27:eb:xx:xx:xx  raspberrypi
        192.168.1.178    00:22:fb:xx:xx:xx  laptop


## Hints

- CentOS / RHEL based distros doesn't send the hostname in the 'DHCP Request' datagram by default.
  To include the hostname in the datagram:

        echo 'DHCP_HOSTNAME=$(hostname -s)' >> /etc/sysconfig/network-scripts/ifcfg-eth0
        