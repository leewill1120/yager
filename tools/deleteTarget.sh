#!/bin/sh
if [ ! $# -eq 2 ];then
	echo "args:IP Port"
	exit
fi

IP=$1
Port=$2

targetList=`curl -sL $IP:$Port/block/list | sed 's/target":"/\n/g' | awk -F '"}' '{print $1}' | grep iqn`

for t in $targetList
do
	echo -e "delete $t\c"
	curl -L -d "{\"target\":\"$t\"}" $IP:$Port/block/delete
	echo ""
done
