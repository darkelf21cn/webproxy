#!/bin/bash
set -e

####### Parameters #######
if [[ -z "${SS_SERVER_HOST}" ]]; then
  echo "environment variable SS_SERVER_HOST is not set"
  exit 1; 
fi

if [[ -z "${SS_SERVER_PORT}" ]]; then
  echo "environment variable SS_SERVER_PORT is not set"
  exit 1; 
fi

if [[ -z "${SS_SERVER_PASSWORD}" ]]; then
  echo "environment variable SS_SERVER_PASSWORD is not set"
  exit 1; 
fi

if [[ -z "${SS_SERVER_ENCRYPT_METHOD}" ]]; then
  echo "environment variable SS_SERVER_ENCRYPT_METHOD is not set"
  exit 1; 
fi

if [[ -z "${SS_SERVER_TIMEOUT}" ]]; then
  echo "environment variable SS_SERVER_TIMEOUT is not set"
  exit 1; 
fi

if [[ -z "${CHINA_IP_LIST_UPDATE_INTERVAL_HOUR}" ]]; then
  echo "environment variable CHINA_IP_LIST_UPDATE_INTERVAL_HOUR is not set, using default value: 24"
  CHINA_IP_LIST_UPDATE_INTERVAL_HOUR=24
fi

####### Variables ########
SQUID_CONF_FILE="/etc/squid/squid.conf"
TMP_SQUID_CONF_FILE="/tmp/squid.conf"

####### Functions ########
function get_cn_ip_list {
  IP_LIST_META_FILE="/tmp/IP_LIST_META_FILE.txt"
  rm -f ${IP_LIST_META_FILE}
  wget -nv --prefer-family=IPv4 http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest -O ${IP_LIST_META_FILE}
  CN_IP_LIST=$(grep "apnic|CN|ipv4|" "${IP_LIST_META_FILE}" | awk -F'|' '{print $4"/"32-log($5)/log(2)}')
  if (( ${#CN_IP_LIST} <= 0 )); then
    echo "CN IP list is empty"
    return 1
  fi
}

function gen_squid_conf {
  get_cn_ip_list
  echo "http_port 3128" > ${TMP_SQUID_CONF_FILE}
  echo "cache_peer 127.0.0.1 parent 8118 0 no-query name=privoxy" >> ${TMP_SQUID_CONF_FILE}
  
  echo "" >> ${TMP_SQUID_CONF_FILE}
  echo "acl noproxy dst 10.0.0.0/8              # Datacenter Network" >> ${TMP_SQUID_CONF_FILE}
  echo "acl nossr dst 100.64.0.0/10           # RFC 6598 shared address space (CGN)"  >> ${TMP_SQUID_CONF_FILE}
  echo "####### China IPs Start #######" >> ${TMP_SQUID_CONF_FILE}
  while read LINE
  do
    echo "acl nossr dst ${LINE}" >> ${TMP_SQUID_CONF_FILE}
  done <<< "${CN_IP_LIST}"
  echo "######## China IPs End ########" >> ${TMP_SQUID_CONF_FILE}
  echo "" >> ${TMP_SQUID_CONF_FILE}

  echo "http_access deny noproxy" >> ${TMP_SQUID_CONF_FILE}
  echo "cache_peer_access privoxy allow all" >> ${TMP_SQUID_CONF_FILE}
  echo "never_direct deny nossr" >> ${TMP_SQUID_CONF_FILE}
  echo "never_direct allow all" >> ${TMP_SQUID_CONF_FILE}
  echo "http_access allow all" >> ${TMP_SQUID_CONF_FILE}
  /usr/sbin/squid -k parse -f ${TMP_SQUID_CONF_FILE}
  mv ${TMP_SQUID_CONF_FILE} ${SQUID_CONF_FILE}
}

function reconfig_squid {
  if [[ "${CHINA_IP_LIST_UPDATE_INTERVAL_HOUR}" == "-1" ]]; then
    return
  fi
  while :
  do
    sleep "${CHINA_IP_LIST_UPDATE_INTERVAL_HOUR}h"
    gen_squid_conf
    /usr/sbin/squid -k reconfigure
  done
}

########## Main ##########
gen_squid_conf
/usr/bin/ss-local -s ${SS_SERVER_HOST} -p ${SS_SERVER_PORT} -l 1080 -k ${SS_SERVER_PASSWORD} -m ${SS_SERVER_ENCRYPT_METHOD} -t ${SS_SERVER_TIMEOUT} -f /run/ss-local.pid
/usr/sbin/privoxy --pidfile /run/privoxy.pid /etc/privoxy/config
/usr/sbin/squid -s -f ${SQUID_CONF_FILE}
reconfig_squid
/usr/bin/tail -f /dev/null
