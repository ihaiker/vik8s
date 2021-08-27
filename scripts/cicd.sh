#!/bin/bash
set -e

export SCRIPTS_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd -P)"
cd $SCRIPTS_PATH/..

CHINA=${CHINA:false}
if [ "$CHINA" == "true" ]; then
  echo "use go proxy"
  export GOPROXY="https://goproxy.cn,direct"
fi

vik8s() {
  ./bin/vik8s --china=${CHINA} -f ./bin $@
}

k8s_install() {
  echo "start install kubernetes clusters."
  vik8s init master01
  vik8s join slave20 slave21
  vik8s cni calico
}

k8s_clean() {
  vik8s reset all
  vik8s clean all --force
}

vagrant_cmd() {
  box_name=$1
  shift
  box=$box_name vagrant $@
}

add_host() {
  echo "add hosts config: run user $1"
  . $SCRIPTS_PATH/hosts.sh "$1"
}

test_plan() {
  box_name=$1
  run_user=$2
  echo " -------------- remove config root directory -------------------"
  rm -rf ./bin/default
  rm -f ./scripts/.vagrant.env

  echo "-------- $run_user test in $box_name start --------"
  echo "vagrant setup"
  vagrant_cmd $box_name up --provision
  time add_host $run_user
  time k8s_install
  time k8s_clean
  vagrant_cmd $box_name destroy -f
  echo "-------- $run_user test in $box_name end   --------"
}

echo " ----------- build ------------- "
. $SCRIPTS_PATH/build.sh

test_plan "centos8" "root"
test_plan "centos7" "root"
test_plan "centos8" "vagrant"
test_plan "centos7" "vagrant"
#test_plain "ubuntu" "vagrant"
