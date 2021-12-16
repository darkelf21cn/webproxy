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
function reconfig_squid {
  if [[ "${CHINA_IP_LIST_UPDATE_INTERVAL_HOUR}" == "-1" ]]; then
    return
  fi
  while :
  do
    sleep "${CHINA_IP_LIST_UPDATE_INTERVAL_HOUR}h"
    /usr/local/sbin/gen_squid_conf.sh
    /usr/sbin/squid -k reconfigure
  done
}

########## Main ##########
/usr/bin/ss-local -s ${SS_SERVER_HOST} -p ${SS_SERVER_PORT} -l 1080 -k ${SS_SERVER_PASSWORD} -m ${SS_SERVER_ENCRYPT_METHOD} -t ${SS_SERVER_TIMEOUT} -f /run/ss-local.pid
/usr/sbin/privoxy --pidfile /run/privoxy.pid /etc/privoxy/config
/usr/sbin/squid -s -f ${SQUID_CONF_FILE}
reconfig_squid
/usr/bin/tail -f /dev/null
