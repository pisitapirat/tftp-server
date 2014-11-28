#!/bin/bash

# Quick and dirty testing.  
# Tested on OSX Yosemite.
# 
# Tested packet loss using pf with 1% probability of loss

function run_test {
	FILENAME=test$1
	FILESIZE=$2
	echo "> Generating test file"
	dd if=/dev/urandom of=$FILENAME bs=$2 count=1
	tftp << EOQ
mode binary
connect 127.0.0.1 9229
put $FILENAME $FILENAME
get $FILENAME $FILENAME-result
quit
EOQ

	EXIT_CODE=0
	diff $FILENAME $FILENAME-result
	if [[ $? != 0 ]]; then
		echo "TEST FAILURE"
		EXIT_CODE=1
	else
		echo "TEST SUCCESS"	
	fi

	rm $FILENAME $FILENAME-result
	exit $EXIT_CODE
}


# Test a file doesn't exist

tftp << EOQ
mode binary
connect 127.0.0.1 9229
get bogus-filename
EOQ
if [[ $? != 1 ]]; then
	echo "TEST FAILURE"
else
	echo "TEST SUCCESS"
fi
rm bogus-filename


# Test parallelism and getting different file sizes

run_test 1 102400 &
run_test 2 102410 & 
run_test 3 202400 &
run_test 4 2400 &
run_test 5 100 &

wait

# In-flight saves shouldn't be visible
# (race-condition here)
run_test 6 1024000 &

tftp << EOQ
mode binary
connect 127.0.0.1 9229
get test6 bogus-filename
EOQ
if [[ $? != 1 ]]; then
	echo "TEST FAILURE"
else
	echo "TEST SUCCESS"
fi
rm bogus-filename

wait