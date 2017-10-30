#!/bin/bash

set -e -u -x

THIS_DIR=$(cd $(dirname $0) && pwd)
cd $THIS_DIR

export CONFIG=/tmp/test-config.json
export APPS_DIR=`pwd`/../src/example-apps

ENVIRONMENT_NAME="$1"
VARS_STORE="$HOME/workspace/cf-networking-deployments/environments/$ENVIRONMENT_NAME/vars-store.yml"


echo "
{
  \"nats_url\": \"api.$ENVIRONMENT_NAME.c2c.cf-app.com\",
  \"nats_username\": \"nats\",
  \"nats_password\": \"{{nats_password}}\",
  \"nats_monitoring_port\": 8222,
  \"num_messages\": 10,
  \"num_publishers\": 10
}
" > ${CONFIG}

NATS_PASSWORD=`grep nats_password ${VARS_STORE} | cut -d' ' -f2`
sed -i -- "s/{{admin-nats_password}}/${NATS_PASSWORD}/g" /tmp/test-config.json

ginkgo -v ../src/performance
