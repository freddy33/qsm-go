#! /bin/bash

dbLoc="was-not-set"
dbConfFile="was-not-set"
curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
. $curDir/functions.sh $curDir

case "$1" in
	start)
		ensureRunningPg
		exit $?
		;;
	stop)
		pg_ctl -D $dbLoc stop
		exit $?
		;;
	conf)
		checkDbConf
		exit $?
		;;
	rmconf)
		rm $dbConfFile
		exit $?
		;;
	status)
		pg_ctl -D $dbLoc status
		exit $?
		;;
	*)
		echo "ERROR: Command $1 unknown"
		exit 1
esac

