% LSLEASES(1) lsleases Manual 
% j-keck [jhyphenkeck@gmail.com]
% June, 2014
  
# NAME

**lsleases** -- dhcp leases sniffer


   
# SYNOPSIS

## CLIENT
  
**lsleases** [-hvVcHnx] 

## SERVER
  
**lsleases** -s [-k]  [-m missed arpings threshold]  [-t ping interval]\
**lsleases** -s -p [-k]  [-e expire duration]  [-t check expired leases interval]
  
  
# DESCRIPTION

**lsleases** captures broadcast 'DHCP Request' datagrams and displays the ip, mac and host name from computers in the local network with dynamic addresses.

  

# MODES

*client:*

in client mode, **lsleases** connects to a running **lsleases** server instance and displays captured ip, mac and host names. 


*server:*

in server mode, **lsleases** captures broadcast 'DHCP Request' datagrams.



Because 'DHCP Release' datagrams are no broadcasts, **lsleases** can not know about invalidated leases. To workaround this problem, there are two methods implemented:

'active mode': check per arping if host online (default) 

'passive mode': clear old leases expire based (-p flag)

The expiration check interval (arping / verify expired leases) is with the flag '-t' configurable.


  
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
:   interval for checking of leases validity (valid units: 'd', 'h', 'm', 's') 

in active mode: arping interval

in passive mode: check expired leases interval

-m
:   in active mode: missed arpings threshold \
remove lease if threshold reached

-k
:   keep leases over restart\
save leases on shutdown / load on startup



# EXAMPLES

list captured leases
  
    j@main:~> lsleases
    Ip               Mac                Name
    192.168.1.189    10:bf:48:xx:xx:xx  android-f6c6dca2130b287
    192.168.1.122    b8:27:eb:xx:xx:xx  raspberrypi
    192.168.1.178    00:22:fb:xx:xx:xx  laptop

  
start server in active mode - arping interval every 10 minutes, remove offline hosts after 5 missed pings

    j@main:~> lsleases -s -t 10m -m 5

  
start server in passive mode - expire leases after 3 days

    j@mail:~> lsleases -s -p -e 3d