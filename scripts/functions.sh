#!/usr/bin/env bash

if [ -z "$QSM_HOME" ]; then
    echo "ERROR: QSM scripts not cqlled from qsm launcher"
    exit 12
fi

rootDir=$QSM_HOME
if [ ! -e "$rootDir/.git" ]; then
    echo "ERROR: Did not find root git repo dir under $rootDir"
    exit 13
fi

confDir="$rootDir/conf"
if [ ! -e "$confDir" ]; then
    echo "ERROR: Did not find conf dir $confDir"
    exit 14
fi

logDir="$rootDir/build/log"

if [ ! -d "$logDir" ]; then
    echo "INFO: Creating log out dir $logDir"
    mkdir -p $logDir
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: could not create log dir $logDir"
        exit 13
    fi
fi

dbLoc="/usr/local/var/postgres"
dbLogFile="$logDir/pgout.log"

rotateDbLog() {
    if [ -e "$dbLogFile" ]; then
        mv "$dbLogFile" "$logDir/pgout-$(date "+%Y-%m-%d_%H-%M-%S").log"
    fi
}

ensureRunningPg() {
    serverNotRunning="$( pg_ctl -D $dbLoc status | grep "no server running" )"

    if [ -n "$serverNotRunning" ]; then
        echo "INFO: PostgreSQL server not running. Status returned $serverNotRunning"
        echo "INFO: Starting PostgreSQL server"
        rotateDbLog
		pg_ctl -w -D $dbLoc start -l $dbLogFile
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Could start server"
            exit $RES
        fi
    fi

    serverRunning="$( pg_ctl -D $dbLoc status | grep "server is running" )"

    if [ -z "$serverRunning" ]; then
        echo "ERROR: PostgreSQL server not running. Status returned $serverRunning"
        exit 11
    fi

    echo "INFO: PostgreSQL server up and running"
}

QSM_ENV_NUMBER=${QSM_ENV_NUMBER:=1}
dbConfFile="$confDir/dbconn${QSM_ENV_NUMBER}.json"

checkDbConf() {
    if [ -e "$dbConfFile" ]; then
        echo "INFO: $dbConfFile already exists"
    else
        echo "INFO: Creating conf file for test number ${QSM_ENV_NUMBER} at $dbConfFile"
        genUser="qsmu${QSM_ENV_NUMBER}"
        genPassword="qsm$RANDOM"
        genName="qsmdb${QSM_ENV_NUMBER}"
        cat $confDir/db-template.json | jq --arg pass "$genPassword" --arg user "$genUser" --arg db "$genName" '.password=$pass | .user=$user | .dbName=$db' > $dbConfFile
        RES=$?
        if [ $RES -ne 0 ]; then
            echo "ERROR: Could create conf file $dbConfFile"
            exit $RES
        fi
    fi

    echo "INFO: Reading existing conf file for test number ${QSM_ENV_NUMBER} at $dbConfFile"
    dbUser="$( cat $dbConfFile | jq -r .user )"
    dbPassword="$( cat $dbConfFile | jq -r .password )"
    dbName="$( cat $dbConfFile | jq -r .dbName )"
    if [ -z "$dbUser" ] || [ -z "$dbPassword" ] || [ -z "$dbName" ]; then
        echo "ERROR: Reading conf file $dbConfFile failed since one of '$dbUser' '$dbPassword' '$dbName' is empty"
        exit 15
    fi
    echo "INFO: Using user $dbUser on $dbName"
}
