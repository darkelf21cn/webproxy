#!/bin/bash
set -e

####### Variables ########
SQUID_CONF_FILE="/etc/squid/squid.conf"
TMP_SQUID_CONF_FILE="/tmp/squid.conf"

########## Main ##########
# get china ip list
IP_LIST_META_FILE="/tmp/IP_LIST_META_FILE.txt"
rm -f ${IP_LIST_META_FILE}
wget -nv --prefer-family=IPv4 http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest -O ${IP_LIST_META_FILE}
CN_IP_LIST=$(grep "apnic|CN|ipv4|" "${IP_LIST_META_FILE}" | awk -F'|' '{print $4"/"32-log($5)/log(2)}')
if (( ${#CN_IP_LIST} <= 0 )); then
echo "CN IP list is empty"
return 1
fi

# generate 
echo "http_port 3128" > ${TMP_SQUID_CONF_FILE}
echo "cache_peer 127.0.0.1 parent 8118 0 no-query name=privoxy" >> ${TMP_SQUID_CONF_FILE}
echo "" >> ${TMP_SQUID_CONF_FILE}
echo "acl nossr dst 10.0.0.0/8              # Datacenter Network" >> ${TMP_SQUID_CONF_FILE}
echo "acl nossr dst 100.64.0.0/10           # RFC 6598 shared address space (CGN)"  >> ${TMP_SQUID_CONF_FILE}
echo "####### China IPs Start #######" >> ${TMP_SQUID_CONF_FILE}
while read LINE
do
echo "acl nossr dst ${LINE}" >> ${TMP_SQUID_CONF_FILE}
done <<< "${CN_IP_LIST}"
echo "######## China IPs End ########" >> ${TMP_SQUID_CONF_FILE}
echo "" >> ${TMP_SQUID_CONF_FILE}
echo "#http_access deny noproxy" >> ${TMP_SQUID_CONF_FILE}
echo "cache_peer_access privoxy allow all" >> ${TMP_SQUID_CONF_FILE}
echo "never_direct deny nossr" >> ${TMP_SQUID_CONF_FILE}
echo "never_direct allow all" >> ${TMP_SQUID_CONF_FILE}
echo "http_access allow all" >> ${TMP_SQUID_CONF_FILE}
/usr/sbin/squid -k parse -f ${TMP_SQUID_CONF_FILE}
mv ${TMP_SQUID_CONF_FILE} ${SQUID_CONF_FILE}
