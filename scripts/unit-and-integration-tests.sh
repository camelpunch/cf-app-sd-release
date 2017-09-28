#!/bin/bash

set -e -u -x

cd service-discovery-release
export GOPATH=$PWD

# declare -a packages=(
#   "src/cf-pusher"
#   "src/example-apps"
#   "src/iptables-logger"
#   "src/lib"
#   "src/netmon"
#   "src/policy-server"
#   )

declare -a serial_packages=(
  "src/service-discovery-controller/integration"
  )

function loadIFB {
  set +e
    depmod "$(uname -r)"
    modprobe ifb
  set -e
}

function bootDB {
  db=$1

  if [ "$db" = "postgres" ]; then
    launchDB="(/docker-entrypoint.sh postgres &> /var/log/postgres-boot.log) &"
    testConnection="psql -h localhost -U postgres -c '\conninfo' &>/dev/null"
  elif [ "$db" = "mysql" ]  || [ "$db" = "mysql-5.6" ]; then
    launchDB="(MYSQL_ROOT_PASSWORD=password /entrypoint.sh mysqld &> /var/log/mysql-boot.log) &"
    testConnection="echo '\s;' | mysql -h 127.0.0.1 -u root --password='password' &>/dev/null"
  else
    echo "skipping database"
    return 0
  fi

  echo -n "booting $db"
  eval "$launchDB"
  for _ in $(seq 1 60); do
    set +e
    eval "$testConnection"
    exitcode=$?
    set -e
    if [ $exitcode -eq 0 ]; then
      echo "connection established to $db"
      return 0
    fi
    echo -n "."
    sleep 1
  done
  echo "unable to connect to $db"
  exit 1
}

loadIFB
bootDB "${DB:-"notset"}"

if [ "${1:-""}" = "" ]; then
  # for dir in "${packages[@]}"; do
  #   pushd "$dir"
  #     ginkgo -r -p --race -randomizeAllSpecs -randomizeSuites -failFast "${@:2}" --skipPackage=timeouts
  #   popd
  # done
  for dir in "${serial_packages[@]}"; do
    pushd "$dir"
      ginkgo -r -randomizeAllSpecs -randomizeSuites -failFast "${@:2}"
    popd
  done
else
  dir=${@: -1}
  for package in "${serial_packages[@]}"; do
    if [ "$dir" = "$package" ] || [ "$dir" = "$package"/ ]; then
      ginkgo -r -randomizeAllSpecs -randomizeSuites "${@}"
      exit $?
    fi
  done
  ginkgo -r -p -randomizeAllSpecs -randomizeSuites "${@}"
fi

