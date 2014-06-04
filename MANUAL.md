% LSLEASES(1) lsleases Manual 
% j-keck [jhyphenkeck@gmail.com]
% June, 2014
  
# NAME

lsleases - dhcp leases sniffer


   
# SYNOPSIS

lsleases [*options*]

  
  
# DESCRIPTION

lsleases captures 'DHCP Request' datagrams and displays the ip, mac and host name from computers in the local network with dynamic addresses.

  

# MODES

*client:*

in client mode, lsleases connects to a running lsleases server instance and displays captured ip, mac and host names. 


*server:*

in server mode, lsleases captures 'DHCP Request' datagrams.


  
Because 'DHCP Release' datagrams are no broadcasts, `lsleases` needs to clear expired leases by self. For this task, there are two  modes available:

'aktive mode': check per arp ping if host online (default) 

'passive mode': clear old leases expire based (-p flag)

The check interval (ping / verify expired leases) is with the flag '-t' configurable.


  
# OPTIONS
  
## common
-h
:    print help
  
-v
:    verbose output
  
-V
:    print version

    
## client
-c
:    clear leases

-H
:    scripted mode: no headers, dates as unix time
  
-n
:    list newest leases first

-x
:    shutdown server

    
## server
-s
:    server mode

-p
:    passive mode - no active availability host check - clear leases expire based

-e
:   in passive mode: lease expire duration (valid units: 'd', 'h', 'm', 's') 
  
-t
:   cleanup leases timer duration (valid units: 'd', 'h', 'm', 's') 

in active mode: ping timer

in passive mode: check expired leases timer

-m
:   in active mode: missed pings threshold
remove lease if threshold reached



# EXAMPLES

        j@main:~> lsleases
        Ip               Mac                Name
        192.168.1.189    10:bf:48:xx:xx:xx  android-f6c6dca2130b287
        192.168.1.122    b8:27:eb:xx:xx:xx  raspberrypi
        192.168.1.178    00:22:fb:xx:xx:xx  laptop  