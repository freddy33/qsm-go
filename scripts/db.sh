#! /bin/bash

QSM_ENV_NUMBER=${QSM_ENV_NUMBER:=1}

curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
# shellcheck source=./functions.sh
. "$curDir/functions.sh"

dbError=13

if [ "$logDir" == "was-not-set" ]; then
    echo "ERROR: functions.sh did not set the logDir var"
    exit 5
fi
if [ "$confDir" == "was-not-set" ]; then
    echo "ERROR: functions.sh did not set the confDir var"
    exit 5
fi

dbLogFile="$logDir/pgout.log"

# Add postgresql bin if exists
is_pg_10="no"
if [ -d "/usr/lib/postgresql/10/bin" ]; then
  echo "INFO: Adding /usr/lib/postgresql/10/bin to path"
  export PATH=/usr/lib/postgresql/10/bin:$PATH
  is_pg_10="yes"
fi

rotateDbLog() {
    if [ -e "$dbLogFile" ]; then
        mv "$dbLogFile" "$logDir/pgout-$(date "+%Y-%m-%d_%H-%M-%S").log"
    fi
}

ensureRunningPg() {
    serverStatus="$( pg_ctl -D $dbLoc status 2>&1 )"

    if [[ $serverStatus == *"not a database cluster directory" ]]; then
        echo "INFO: PostgreSQL folder $dbLoc not initialized as DB. Initializing DB"
        initdb -D $dbLoc
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Could initialize DB directory"
            exit $RES
        fi
        serverStatus="$( pg_ctl -D $dbLoc status 2>&1 )"
    fi

    if [[ $serverStatus == *"no server running" ]]; then
        echo "INFO: PostgreSQL server not running. Starting PostgreSQL server"
        rotateDbLog
        if [ "$is_pg_10" == "no" ]; then
          echo "INFO: Copying basic PostgreSQL server configuration"
          cp $confDir/postgresql.conf $dbLoc/postgresql.conf
        fi
		    pg_ctl -w -D $dbLoc start -l $dbLogFile
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Could start postgresql DB server"
            tail -50 $dbLogFile
            exit $RES
        fi
        serverStatus="$( pg_ctl -D $dbLoc status 2>&1 )"
    fi

    echo "INFO: Current DB status $serverStatus"

    if [[ $serverStatus == *"server is running"* ]]; then
        echo "INFO: PostgreSQL server up and running"
        return 0
    fi

    echo "ERROR: PostgreSQL server at $dbLoc not running."
    exit 11
}

ensureUser() {
    echo "INFO: Checking user $dbUser"
    checkUser=$(psql -h $dbHost -p $dbPort -qAt postgres -c "select 1 from pg_catalog.pg_user u where u.usename='$dbUser';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for user presence"
        exit $dbError
    fi
    if [ "x$checkUser" == "x1" ]; then
        echo "INFO: User $dbUser already exists"
    else
        echo "INFO: Creating user $dbUser"
        psql -h $dbHost -p $dbPort -qAt postgres -c "create user $dbUser with encrypted password '$dbPassword';"
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to create user $dbUser"
            exit $dbError
        fi
        echo "INFO: User $dbUser created"
    fi
}

ensureDb() {
    echo "INFO: Checking db $dbName"
    checkDb=$(psql -h $dbHost -p $dbPort -qAt postgres -c "SELECT 1 FROM pg_database WHERE datname='$dbName';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for DB presence"
        exit $dbError
    fi
    if [ "x$checkDb" == "x1" ]; then
        echo "INFO: Database $dbName already exists"
    else
        echo "INFO: Creating database $dbName"
        psql -h $dbHost -p $dbPort -qAt postgres <<EOF
create database $dbName;
grant all privileges on database $dbName to $dbUser;
EOF
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to create database $dbName"
            exit $dbError
        fi
        echo "INFO: Database $dbName created"
    fi
}

dropUser() {
    echo "INFO: Dropping user $dbUser"
    checkUser=$(psql -h $dbHost -p $dbPort -qAt postgres -c "select 1 from pg_catalog.pg_user u where u.usename='$dbUser';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for user presence"
        exit $dbError
    fi
    if [ "x$checkUser" == "x1" ]; then
        echo "INFO: User $dbUser exists => deleting it"
        psql -h $dbHost -p $dbPort -qAt postgres -c "drop user $dbUser;"
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to drop user $dbUser"
            exit $dbError
        fi
        echo "INFO: User $dbUser deleted"
    else
        echo "INFO: User $dbUser already deleted"
    fi
}

dropDb() {
    echo "INFO: Dropping db $dbName"
    checkDb=$(psql -h $dbHost -p $dbPort -qAt postgres -c "SELECT 1 FROM pg_database WHERE datname='$dbName';")
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for DB presence"
        exit $dbError
    fi
    if [ "x$checkDb" == "x1" ]; then
        echo "INFO: Database $dbName exists => Dropping it"
        psql -h $dbHost -p $dbPort -qAt postgres <<EOF
drop database $dbName;
EOF
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Failed to drop database $dbName"
            exit $dbError
        fi
        echo "INFO: Database $dbName dropped"
    else
        echo "INFO: Database $dbName already dropped"
    fi
}

case "$1" in
	check)
	    checkDbConf || exit $?
        echo "INFO: Checking all good for QSM dev using:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
		ensureRunningPg && ensureUser && ensureDb
		exit $?
		;;
	drop)
	    checkDbConf || exit $?
        echo "INFO: Dropping QSM environment:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
	    ensureRunningPg && dropDb && dropUser && rm $dbConfFile
		exit $?
		;;
	dropAll)
	    checkDbConf || exit $?
        echo "INFO: Dropping ALL QSM environments except 1"
        RES=0
        for envId in 2 3 4 5 6 7 8 9 10 11; do
            export QSM_ENV_NUMBER=$envId
            ./qsm db drop
            LOOP_RES=$?
            if [ $LOOP_RES -ne 0 ]; then
                RES=$LOOP_RES
            fi
        done
		exit $RES
		;;
	start)
		ensureRunningPg
		exit $?
		;;
	stop)
		pg_ctl -D $dbLoc stop
		exit $?
		;;
	conf)
	    checkDbConf || exit $?
        echo "INFO: Checked configuration for QSM environment:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
		exit 0
		;;
	shell)
	    checkDbConf || exit $?
	    psql -h $dbHost -p $dbPort -U "$dbUser" $dbName
		exit $?
		;;
	dump)
	    checkDbConf || exit $?
	    pg_dump -U "$dbUser" $dbName | gzip > $dumpDir/$dbName-$(date "+%Y-%m-%d_%H-%M-%S").dump.sql.gz
		exit $?
		;;
	rmconf)
		rm $dbConfFile
		exit $?
		;;
	status)
	    checkDbConf || exit $?
        echo "INFO: Checking PostgreSQL using:"
        echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
		pg_ctl -D $dbLoc status
		exit $?
		;;
	*)
		echo "ERROR: Command $1 unknown"
		exit 1
esac


