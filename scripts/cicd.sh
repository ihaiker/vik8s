#!/bin/bash
set -e

export SCRIPTS_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd -P)"
cd $SCRIPTS_PATH/..

vik8s() {
  ./bin/vik8s $@
}

machine_install(){
  box=$1 vagrant up
}
machine_destroy(){
  vagrant destroy -f
}

cluster_install() {
  echo "start install kubernetes clusters."
  vik8s init master01
  vik8s join slave20 slave21
#  vik8s cni calico
}

k8s_clean() {
  vik8s reset all
  vik8s clean all --force
}

add_host() {
  echo "add hosts config: run user $1"
}

test_plan() {
  box_name=$1
  run_user=$2
  echo " -------------- remove config root directory -------------------"
  rm -rf ~/.vik8s/default

  echo "-------- $run_user test in $box_name start --------"
  echo "vagrant setup"
  time machine_install $box_name
  time add_host $run_user
  time cluster_install
#  time k8s_clean
#  time machine_destroy
  echo "-------- $run_user test in $box_name end   --------"
}

echo " ----------- build ------------- "
. $SCRIPTS_PATH/build.sh

test_plan "centos8" "root"
#test_plan "centos7" "root"
#test_plan "centos8" "vagrant"
#test_plan "centos7" "vagrant"
#test_plain "ubuntu" "vagrant"
