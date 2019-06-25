#! /bin/bash

echo "INFO: Checking all good for QSM dev"

curDir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
. $curDir/functions.sh $curDir

ensureUser() {
    echo "INFO: Checking user"
}

