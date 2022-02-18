#!/bin/bash

docker-compose up -d
rm tmp_expected_file 2>/dev/null

# connecting to janus and collecting all messages
mosquitto_sub -t \# -v > tmp_res_file &

# publishing
mosquitto_pub -t /silent -m "we will not get this message back"

mosquitto_pub -t /verbose -m "this message will be returned"
echo "/verbose this message will be returned" >> tmp_expected_file

mosquitto_pub -t /val_map_test -m "get_msg" # this text will be replaced with another
echo "/val_map_msg val_map works!" >> tmp_expected_file

# comparing expected and actual files
diff tmp_expected_file tmp_res_file
result=$?

# kill mosquitto_sub process and clean up
pkill -P $$
rm tmp_res_file
docker-compose stop
rm tmp_expected_file

echo
if [ $result -eq "0" ];
then
  echo "Tests ok";
else
  echo "ERROR"
fi
