#!/bin/sh
SERVER=$1
URL="$SERVER/nexus/service/local/artifact/maven/content"
REPO="thirdparty"
USER=$2
group=bamboo
artifact=bamboo
version=1.0.0
classifier=
ext=rpm
filename=bamboo-1.0.0_1-1.noarch.rpm
curl --write-out "\nStatus: %{http_code}\n" \
--request POST \
--user $USER \
-F "r=$REPO" \
-F "g=$group" \
-F "a=$artifact" \
-F "v=$version" \
-F "c=$classifier" \
-F "p=$ext" \
-F "file=@$filename" \
"$URL"