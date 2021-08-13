#!/bin/bash
set -e
export HOSTS_SH_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd -P)"

pushd $HOSTS_SH_PATH/..
  source ${HOSTS_SH_PATH}/vagrant.sh
  mkdir -p $HOME/.ssh
  vagrant ssh-config | tee $HOME/.ssh/vagrant_config
  hosts=$(cat "$HOME/.ssh/vagrant_config" | grep "Host " | awk '{print $2}')
  for host in $hosts ; do
      ssh_config=$(vagrant ssh-config $host)
      hostname=$(echo "$ssh_config" | grep "HostName " | awk '{print $2}')
      port=$(echo "$ssh_config" | grep "Port " | awk '{print $2}')
      username=$(echo "$ssh_config" | grep "User " | awk '{print $2}')
      identity=$(echo "$ssh_config" | grep "IdentityFile " | awk '{print $2}')
      if [ "x$hostname" == "x127.0.0.1" ]; then
          echo "provider == private_network"
          hostname=$(vagrant ssh $host -c 'ifconfig eth1' | grep "inet " | awk '{print $2}')
          port=22
      fi
      ./bin/vik8s -f ./bin hosts --user $username --private-key $identity --port $port $hostname
  done
popd
