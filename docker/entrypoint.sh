#!/bin/sh

# Use env: SYNCHER_APP_EXECUTION_ARGS for execution args
# Loops on executing the golang artifacts syncher

export SLEEP_BETWEEN_ACTIONS_SEC=${SLEEP_BETWEEN_ACTIONS_SEC:-300}  # 5 minutes
export SYNCHER_APP_EXECUTION_ARGS=${SYNCHER_APP_EXECUTION_ARGS}  # App execution args


# Loop forever
while true
do
	echo $(date '+%d/%m/%Y %H:%M:%S')
	echo Executing: golang-artifacts-syncher ${SYNCHER_APP_EXECUTION_ARGS}
	golang-artifacts-syncher ${SYNCHER_APP_EXECUTION_ARGS}
	if [ "$?" != "0" ]; then echo '' && echo Error - Failure during execution of: golang-artifacts-syncher && echo '' ; fi
	echo ''
	echo Executed: golang-artifacts-syncher ${SYNCHER_APP_EXECUTION_ARGS}
	echo ''
	echo $(date '+%d/%m/%Y %H:%M:%S')
	echo Sleeping for: $SLEEP_BETWEEN_ACTIONS_SEC seconds
	sleep $SLEEP_BETWEEN_ACTIONS_SEC
	echo Finished sleeping
	sleep 1
	echo ''
done


