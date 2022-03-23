#!/bin/sh

# Loops on executing the golang artifacts syncher

export SLEEP_BETWEEN_ACTIONS_SEC=${SLEEP_BETWEEN_ACTIONS_SEC:-300}  # 5 minutes

# Loop forever
while true
do
	echo $(date '+%d/%m/%Y %H:%M:%S')
	echo Executing: golang-artifacts-syncher
	golang-artifacts-syncher
	if [ "$?" != "0" ]; then echo '' && echo Error - Failure during execution of: golang-artifacts-syncher && echo '' ; fi
	echo ''
	echo $(date '+%d/%m/%Y %H:%M:%S')
	echo Sleeping for: $SLEEP_BETWEEN_ACTIONS_SEC seconds
	sleep $SLEEP_BETWEEN_ACTIONS_SEC
	echo Finished sleeping
	sleep 1
	echo ''
done


