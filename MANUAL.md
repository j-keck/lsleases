% LSLEASES(1) lsleases Manual 
% j-keck [jhyphenkeck@gmail.com]
% June, 2014
  
# NAME

**lsleases** -- dhcp leases sniffer


   
# SYNOPSIS

## CLIENT
  
**lsleases** [-hvVcHnx] 

## SERVER
  
**lsleases** -s [-k]  [-m missed pings threshold]  [-t ping interval]\
**lsleases** -s -p [-k]  [-e expire duration]  [-t check expired leases interval]
  
  
# DESCRIPTION

**lsleases** watches your dhcp network traffic and gives you easy access to assigned adresses and active devices. (wenn du das ok findest, kann auch bei readme verwendet werden)
It captures broadcast 'DHCP Request' datagrams and displays the ip, mac and host name from computers in the local network with dynamic ip address.

  

# MODES

*client:*

in client mode, **lsleases** connects to a running **lsleases** server instance and displays captured ips, macs and host names. 


*server:*

in server mode, **lsleases** captures broadcast 'DHCP Request' datagrams.



Because - unlike 'DHCP Request' - 'DHCP Release' datagrams are no broadcasts, **lsleases** can not know about invalidated leases. To workaround this problem, there are two methods implemented:

'active mode': use ping (icmp on windows, arping others) to check if host online (default)



'passive mode'  (-p flag) : clear old leases based on expiration time (-e flag)

The expiration check interval (ping / verify expired leases) is configurable with the -t flag.


  
# OPTIONS
  ** Multiple Flags have to be specifyed individually and separated by blanks (see Examples) **
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
:    passive mode - no active availability checking - clear leases expiration based

-e
:   in passive mode: lease expiration duration (valid units: 'd', 'h', 'm', 's') 
  
-t
:   interval for checking of leases validity / reachability (valid units: 'd', 'h', 'm', 's') 

in active mode: ping interval

in passive mode: check expired leases interval

-m
:   in active mode: missed pings threshold \
remove lease if threshold reached

-k
:   keep leases over restart (save leases on shutdown and load on startup )


# CONFIGURATION
  (das hier würde ich oben bei den options hinschreiben, hat ja mit configuration nix zu tun)
  **!keep every flag separate - so to enable persistent leases and passive mode, write: "-k -p" - see EXAMPLES!**
  
To configure the server, set the corresponding option flags:

### FreeBSD
  in the file `/etc/rc.conf`:

    `lsleases_flags=""`

### Linux
  in the file `/etc/default/lsleases`:

    `DAEMON_OPTS=""`

### Windows
  ***!keep in mind to let the parameter '-s' untouched!***

  **standalone**
  
    in the file `<INSTALL_PATH>\start-server.bat`

  **installed as service**
  
    in the Registry under: `HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\lsleases\Parameters\AppParameters`

    restart the service per `<INSTALL_PATH>\restart-service.bat` (per right mouse click and "Run as Administrator")

# EXAMPLES


Specify Flags separately (here: Server in passive mode with persistent leases) 

     j@main:~> lsleases -s -k -p  (-skp or -s-k-p  DO NOT WORK)
    
    
list captured leases
  
    j@main:~> lsleases
    Ip               Mac                Name
    192.168.1.189    10:bf:48:xx:xx:xx  android-f6c6dca2130b287
    192.168.1.122    b8:27:eb:xx:xx:xx  raspberrypi
    192.168.1.178    00:22:fb:xx:xx:xx  laptop

  
start server in active mode - ping interval every 10 minutes, remove offline hosts after 5 missed pings

    j@main:~> lsleases -s -t 10m -m 5

  
start server in passive mode - expire leases after 3 days

    j@mail:~> lsleases -s -p -e 3d
