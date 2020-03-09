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

pg_ext=""
dbLocExe="$dbLoc"
if [ "$is_windows" == "yes" ]; then
  pg_ext=".exe"
  dbLocExe="$(wslpath -w "$dbLoc")"
fi

dbLogFile="$logDir/pgout.log"

# Add postgresql bin if exists
is_pg_10="no"
if [ -d "/usr/lib/postgresql/10/bin" ]; then
  echo "INFO: Adding /usr/lib/postgresql/10/bin to path"
  sudo chmod a+w /var/run/postgresql
  export PATH=/usr/lib/postgresql/10/bin:$PATH
  is_pg_10="yes"
fi

rotateDbLog() {
    if [ -e "$dbLogFile" ]; then
        mv "$dbLogFile" "$logDir/pgout-$(date "+%Y-%m-%d_%H-%M-%S").log"
    fi
}

getServerStatus() {
  serverStatus="$( pg_ctl$pg_ext -D "$dbLocExe" status 2>&1 )"
  serverStatus="${serverStatus//$'\r'}"
}

ensureRunningPg() {
    getServerStatus
    echo "testing if $serverStatus is a db"
    if [[ "$serverStatus" == *"not a database cluster directory" ]]; then
        echo "INFO: PostgreSQL folder $dbLocExe not initialized as DB. Initializing DB"
        initdb$pg_ext -D "$dbLocExe"
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Could initialize DB directory"
            exit $RES
        fi
        getServerStatus
    fi

    executed_pgctl="no"
    if [[ $serverStatus == *"no server running" ]]; then
        echo "INFO: PostgreSQL server not running. Starting PostgreSQL server"
        rotateDbLog
        if [ "$is_pg_10" == "no" ] && [ "$is_windows" == "no" ]; then
          echo "INFO: Copying basic PostgreSQL server configuration"
          cp $confDir/postgresql.conf $dbLoc/postgresql.conf
        fi
        if [ "$is_windows" == "yes" ]; then
          dbLogFileExe="$(wslpath -w "$dbLogFile")"
        else
          dbLogFileExe="$dbLocFile"
        fi
        echo "Executing 'pg_ctl$pg_ext -o \"-F -p $dbPort\" -w -D \"$dbLocExe\" start -l \"$dbLogFileExe\"'"
        pg_ctl$pg_ext -o "-F -p $dbPort" -w -D "$dbLocExe" start -l "$dbLogFileExe" &
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Could not start postgresql DB server"
            tail -50 "$dbLogFile"
            exit $RES
        fi
        sleep 1
        executed_pgctl="yes"
        getServerStatus
    fi

    echo "INFO: Current DB status $serverStatus"

    if [[ $serverStatus == *"server is running"* ]]; then
        echo "INFO: PostgreSQL server at $dbLocExe up and running"
        return 0
    fi

    echo "ERROR: PostgreSQL server at $dbLocExe not running."
    if [ "$executed_pgctl" == "yes" ]; then
        echo "ERROR: Executed the pg control command to start server and still not up. The log shows:"
        tail -50 "$dbLogFile"
    fi
    exit 11
}

ensureUser() {
    echo "INFO: Checking user $dbUser"
    checkUser=$(psql$pg_ext -h $dbHost -p $dbPort -qAt -c "select 1 from pg_catalog.pg_user u where u.usename='$dbUser';" postgres)
    checkUser="${checkUser//$'\r'}"
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for user presence"
        exit $dbError
    fi
    if [ "x$checkUser" == "x1" ]; then
        echo "INFO: User $dbUser already exists"
    else
        echo "INFO: Creating user $dbUser"
        psql$pg_ext -h $dbHost -p $dbPort -qAt -c "create user $dbUser with encrypted password '$dbPassword';" postgres
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
    checkDb=$(psql$pg_ext -h $dbHost -p $dbPort -qAt -c "SELECT 1 FROM pg_database WHERE datname='$dbName';" postgres)
    checkDb="${checkDb//$'\r'}"
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for DB presence"
        exit $dbError
    fi
    if [ "x$checkDb" == "x1" ]; then
        echo "INFO: Database $dbName already exists"
    else
        echo "INFO: Creating database $dbName"
        psql$pg_ext -h $dbHost -p $dbPort -qAt postgres <<EOF
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
    checkUser=$(psql$pg_ext -h $dbHost -p $dbPort -qAt -c "select 1 from pg_catalog.pg_user u where u.usename='$dbUser';" postgres)
    checkUser="${checkUser//$'\r'}"
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for user presence"
        exit $dbError
    fi
    if [ "x$checkUser" == "x1" ]; then
        echo "INFO: User $dbUser exists => deleting it"
        psql$pg_ext -h $dbHost -p $dbPort -qAt -c "drop user $dbUser;" postgres
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
    checkDb=$(psql$pg_ext -h $dbHost -p $dbPort -qAt -c "SELECT 1 FROM pg_database WHERE datname='$dbName';" postgres)
    checkDb="${checkDb//$'\r'}"
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: Failed to check for DB presence"
        exit $dbError
    fi
    if [ "x$checkDb" == "x1" ]; then
        echo "INFO: Database $dbName exists => Dropping it"
        psql$pg_ext -h $dbHost -p $dbPort -qAt postgres <<EOF
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
    checkDbConf || exit $?
    ensureRunningPg
    exit $?
    ;;
  stop)
    pg_ctl$pg_ext -D "$dbLocExe" stop
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
    psql$pg_ext -h $dbHost -p $dbPort -U "$dbUser" $dbName
    exit $?
    ;;
  dump)
    checkDbConf || exit $?
    pg_dump$pg_ext -U "$dbUser" $dbName | gzip > "$dumpDir/$dbName-$(date "+%Y-%m-%d_%H-%M-%S").dump.sql.gz"
    exit $?
    ;;
  rmconf)
    rm "$dbConfFile"
    exit $?
    ;;
  status)
    checkDbConf || exit $?
    echo "INFO: Checking PostgreSQL using:"
    echo -ne "QSM_ENV_NUMBER=${QSM_ENV_NUMBER}\ndbName=$dbName\ndbUser=$dbUser\n"
    pg_ctl$pg_ext -D "$dbLoc" status
    exit $?
    ;;
  *)
    echo "ERROR: Command $1 unknown"
    exit 1
esac
