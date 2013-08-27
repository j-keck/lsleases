# lsleases - dhcp leases sniffer

lsleases captures 'DHCP Request' datagrams and displays the ip, mac and host name from computers in the local network with dynamic addresses.


## Installation

### From source

  1. export GOPATH: `export GOPATH=<PATH>/lsleases-workdir`
  2. download and build: `go get github.com/j-keck/lsleases`. This will build the binary in $GOPATH/bin

  **Linux**
  3. set capabilities to open port less than 1024 (dhcp sniff) and use raw sockets (arp ping)
  
     `setcap 'cap_net_raw,cap_net_bind_service+ep' <PATH>/lsleases`

  **FreeBSD**
  3. allow to open port less than 1024 (dhcp sniff)
  
        echo net.inet.ip.portrange.reservedhigh=0 >> /etc/sysctl.conf
        service sysctl restart

  *arping as non-root under FreeBSD currently not supported*  

 

### Binary packages
  1. download from http://github.com/j-keck/lsleases/releases/latest

  *deb packages register and starts an server instance on installation*

  
## Usage

1. start an server instance `nohup lsleases -s &`
2. replug / startup any device with dynamic addresses
3. display captured ip, mac and host names. 

        j@main:~> lsleases
        Ip               Mac                Name
        192.168.1.189    10:bf:48:xx:xx:xx  android-f6c6dca2130b287
        192.168.1.122    b8:27:eb:xx:xx:xx  raspberrypi
        192.168.1.178    00:22:fb:xx:xx:xx  laptop

