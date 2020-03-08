#!/usr/bin/env bash

if [ -z "$QSM_HOME" ]; then
    echo "ERROR: QSM scripts not be called out of the qsm launcher"
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

go_exe="$(which go)"
if [ $? -eq 0 ]; then
  is_windows="no"
else
  go_exe="$(which go.exe)"
  if [ $? -eq 0 ]; then
    is_windows="yes"
  else
    echo "ERROR: did not find go or go.exe"
    exit 14
  fi
fi

buildDir="$rootDir/build"
logDir="$buildDir/log"

if [ ! -d "$logDir" ]; then
    echo "INFO: Creating log out dir $logDir"
    mkdir -p $logDir
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: could not create log dir $logDir"
        exit 13
    fi
fi

dumpDir="$buildDir/dump"

if [ ! -d "$dumpDir" ]; then
    echo "INFO: Creating dump dir $dumpDir"
    mkdir -p $dumpDir
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: could not create dump dir $dumpDir"
        exit 13
    fi
fi

dbLoc="$buildDir/postgres"

if [ ! -d "$dbLoc" ]; then
    echo "INFO: Creating database dir $dbLoc"
    mkdir -p $dbLoc
    RES=$?
    if [ $RES -ne 0 ]; then
        echo "ERROR: could not create database dir $dbLoc"
        exit 13
    fi
fi

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
    dbPort="$( cat $dbConfFile | jq -r .port )"
    dbHost="$( cat $dbConfFile | jq -r .host )"
    dbUser="$( cat $dbConfFile | jq -r .user )"
    dbPassword="$( cat $dbConfFile | jq -r .password )"
    dbName="$( cat $dbConfFile | jq -r .dbName )"
    if [ -z "$dbUser" ] || [ -z "$dbPassword" ] || [ -z "$dbName" ]; then
        echo "ERROR: Reading conf file $dbConfFile failed since one of '$dbUser' '$dbPassword' '$dbName' is empty"
        exit 15
    fi
    echo "INFO: Using user $dbUser on $dbHost:$dbPort/$dbName"
}
