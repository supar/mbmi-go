#!/bin/bash
# SBSS service level module init script


. /lib/lsb/init-functions

NAME=mbmi-go
DAEMON=/usr/sbin/${NAME}
PIDFILE=/var/run/${NAME}.pid

[ -f /etc/${NAME}/${NAME}-server.conf ] && . /etc/${NAME}/${NAME}-server.conf

[ -z "$SERVER" ] || ARGS="$ARGS -L $SERVER"
[ -z "$VERBOSE" ] || ARGS="$ARGS -v $VERBOSE"
[ -z "$DBHOST" ] || ARGS="$ARGS -Dh $DBHOST"
[ -z "$DBUSER" ] || ARGS="$ARGS -Du $DBUSER"
[ -z "$DBPASS" ] || ARGS="$ARGS -Dp $DBPASS"
[ -z "$DBNAME" ] || ARGS="$ARGS -Db $DBNAME"
[ -z "$ASSETS" ] || ARGS="$ARGS -A $ASSETS"

status_service() {
    if [ -e $PIDFILE ]; then
        status_of_proc -p $PIDFILE $DAEMON "$NAME process" && exit 0 || exit $?
    else
        log_daemon_msg "$NAME process is not running"
        log_end_msg 0
    fi
}

start_service() {
    if [ -e $PIDFILE ]; then
        status_of_proc -p $PIDFILE ${DAEMON} "$NAME process" && status="0" || status="$?"

        if [ $status = "0" ]; then
            exit
        fi
    fi

    log_daemon_msg "Starting the process" "$NAME"
    echo

    if start-stop-daemon --start --quiet --background --oknodo --make-pidfile --pidfile $PIDFILE --exec ${DAEMON} -- $ARGS; then
        log_end_msg 0
    else
        log_end_msg 1
    fi
}

stop_service() {
    if [ -e $PIDFILE ]; then
        status_of_proc -p $PIDFILE ${DAEMON} "Stoppping the $NAME process" && status="0" || status="$?"

        if [ "$status" = 0 ]; then
            start-stop-daemon --stop --quiet --oknodo --pidfile $PIDFILE
            /bin/rm -rf $PIDFILE
        fi
     else
        log_daemon_msg "$NAME process is not running"
        log_end_msg 0
    fi
}

case "$1" in
    start)
        start_service
    ;;
    stop)
        stop_service
    ;;
    restart)
        stop_service
        sleep 2
        start_service
    ;;
    status)
        status_service
    ;;
    *)
        echo "Usage: /etc/init.d/${NAME} {start|stop|restart|status}"
    ;;
esac

exit 0
