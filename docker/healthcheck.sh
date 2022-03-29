#!/bin/sh

# Check for execution success indicator file existence

export SUCCESS_INDICATOR_FILE_PATH=${SUCCESS_INDICATOR_FILE_PATH:-/usr/bin/syncher_execution_finished_successfully.txt}

# Check for file existence
echo $(date '+%d/%m/%Y %H:%M:%S')
echo Checking file: ${SUCCESS_INDICATOR_FILE_PATH} exists
if [ -f "$SUCCESS_INDICATOR_FILE_PATH" ]; then echo 'OK - File exists' && exit 0 ; fi
echo ''
echo 'NO - File doesnt exist. Exiting with error code: 1'
exit 1


