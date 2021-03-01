SIZE=$((2**16+1))

# define log file
LOG_FILE=test.log
rm -f ${LOG_FILE}

# temp files for send and receive
SEND_FILE=$(mktemp)
RECEIVE_FILE=$(mktemp)

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

# compare send and receive file
ls -l ${SEND_FILE} ${RECEIVE_FILE}
cmp ${SEND_FILE} ${RECEIVE_FILE}

# show log
cat ${LOG_FILE}

# clean up
rm ${SEND_FILE} ${RECEIVE_FILE}
