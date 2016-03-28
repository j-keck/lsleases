# Changelog
  
*only notable changes are listed*  

##1.4.x##

####1.4.1####
- bugfix in address selection if host has also a v6 address
- windows: fix update in non default directory
- windows: add version in title

[all changes since 1.4.0](https://github.com/j-keck/lsleases/compare/1.4.0...1.4.1)

####1.4.0####
- watch for new leases via '-w' flag - client polls server every second for new leases
- windows installer uninstalls already installed old version

##1.3##
- persist leases over restarts via '-k' flag (disable by default)
- active alive check under windows (per icmp ping) (enabled by default)

*1.3.1 (windows only)*

  - fix pipe permission issue if running as windows service
 
##1.2##
- windows installer
- rework binary packages
- internal build / test structure perl based

##1.1##
- shutdown server from client per '-x' flag
- rewording help usage
- rpm packages
- FreeBSD packages
- windows zip with hacky .bat scripts to start/stop an server instance and list leases
- set host name to \<UNKNOW\> if not existing in the datagram
  
##1.0##
- initial public release
