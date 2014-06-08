#!/bin/sh
### BEGIN INIT INFO
# Provides:          lsleases
# Required-Start:    $local_fs $network $remote_fs $syslog
# Required-Stop:     $local_fs $network $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: dhcp leases sniffer
# Description:       lsleases captures 'DHCP Request' datagrams and displays the ip, mac and host name
#                    from computers in the local network with dynamic addresses.
### END INIT INFO

# Author: j <jhyphenkeck@gmail.com>

# Do NOT "set -e"

# PATH should only include /usr/* if it runs after the mountnfs.sh script
PATH=/sbin:/usr/sbin:/bin:/usr/bin
DESC="dhcp leases sniffer"
NAME=lsleases
DAEMON=/usr/bin/lsleases
DAEMON_ARGS="-s"
DAEMON_USER="nobody"
PIDFILE=/var/run/$NAME.pid
SCRIPTNAME=/etc/init.d/$NAME
LOGFILE="/var/log/${NAME}.log"

# Exit if the package is not installed
[ -x "$DAEMON" ] || exit 0

# Read configuration variable file if it is present
[ -r /etc/default/$NAME ] && . /etc/default/$NAME

# Load the VERBOSE setting and other rcS variables
. /lib/init/vars.sh

# Define LSB log_* functions.
# Depend on lsb-base (>= 3.2-14) to ensure that this file is present
# and status_of_proc is working.
. /lib/lsb/init-functions

#
# Function that starts the daemon/service
#
do_start()
{
    # create logfile and set ownership
    touch $LOGFILE && chown $DAEMON_USER $LOGFILE

    # Return
    #   0 if daemon has been started
    #   1 if daemon was already running
    #   2 if daemon could not be started
    start-stop-daemon --start --quiet --pidfile $PIDFILE --exec $DAEMON --test > /dev/null \
        || return 1
    start-stop-daemon --start --quiet --pidfile $PIDFILE  --background --chuid $DAEMON_USER \
        --exec /bin/bash -- -c "$DAEMON $DAEMON_ARGS $DAEMON_OPTS > $LOGFILE 2>&1" \
        || return 2

}

#
# Function that stops the daemon/service
#
do_stop()
{
    $DAEMON -x
    RETVAL="$?"

    # Remove pid file
    rm -f $PIDFILE

    # Remove sock file
    rm -f /tmp/$NAME.sock
    
    return "$RETVAL"
}

case "$1" in
    start)
        [ "$VERBOSE" != no ] && log_daemon_msg "Starting $DESC" "$NAME"
        do_start
        case "$?" in
            0|1) [ "$VERBOSE" != no ] && log_end_msg 0 ;;
            2) [ "$VERBOSE" != no ] && log_end_msg 1 ;;
        esac
        ;;
    stop)
        [ "$VERBOSE" != no ] && log_daemon_msg "Stopping $DESC" "$NAME"
        do_stop
        case "$?" in
            0|1) [ "$VERBOSE" != no ] && log_end_msg 0 ;;
            2) [ "$VERBOSE" != no ] && log_end_msg 1 ;;
        esac
        ;;
    status)
        status_of_proc "$DAEMON" "$NAME" && exit 0 || exit $?
        ;;
    restart)
        log_daemon_msg "Restarting $DESC" "$NAME"
        do_stop
        case "$?" in
            0|1)
                do_start
                case "$?" in
                    0) log_end_msg 0 ;;
                    1) log_end_msg 1 ;; # Old process is still running
                    *) log_end_msg 1 ;; # Failed to start
                esac
                ;;
            *)
                # Failed to stop
                log_end_msg 1
                ;;
        esac
        ;;
    *)
        echo "Usage: $SCRIPTNAME {start|stop|status|restart}" >&2
        exit 3
        ;;
esac

:
