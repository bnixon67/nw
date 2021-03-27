SIZE=$((2**16+1))

if ! [[ -x ./nw ]]
then
	echo "Executable does not exist. Building executable."
	if ! go build
	then
		echo "Build failed!"
		exit 1
	fi
fi

# define log file
LOG_FILE=test.log
rm -f ${LOG_FILE}

# temp files for send and receive
SEND_FILE=$(mktemp -t nw)
RECEIVE_FILE=$(mktemp -t nw)

# common args
ARGS="-logFileName ${LOG_FILE} -host localhost"

# create file with random bytes
head -c ${SIZE} < /dev/urandom > ${SEND_FILE}

# startup receiver
./nw ${ARGS} -overwrite receive ${RECEIVE_FILE} &
WAIT_PID=$!

# delay to allow time for receiver to start
sleep 0.1

# send file
./nw ${ARGS} send ${SEND_FILE}

# wait for receive to finish
wait ${WAIT_PID}

# show log
cat ${LOG_FILE}

# compare send and receive file
ls -l ${SEND_FILE} ${RECEIVE_FILE}
if cmp -s ${SEND_FILE} ${RECEIVE_FILE}
then
	echo "Files matches"
	rm ${SEND_FILE} ${RECEIVE_FILE}
else
	echo "Files not do match"
fi
