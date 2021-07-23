#!/bin/bash
set -e

export BASE_PATH="$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  pwd -P
)"
cd $BASE_PATH/..

CHINA=${CHINA:false}
if [ "$CHINA" == "true" ]; then
  echo "use go proxy"
  export GOPROXY="https://goproxy.cn,direct"
fi

vik8s() {
  ./bin/vik8s --china=${CHINA} -f ./bin/etc $@
}

k8s_install() {
  echo "start install kubernetes clusters."
  vik8s init --ssh-pk=.vagrant/machines/master0/virtualbox/private_key --k8s-version=1.19.12 --master=10.24.0.10
  #  vik8s join --ssh-pk=.vagrant/machines/slave20/virtualbox/private_key --node=10.24.0.20
  #  vik8s join --ssh-pk=.vagrant/machines/slave21/virtualbox/private_key --node=10.24.0.21
}

k8s_clean() {
  vik8s clean --force
  vik8s reset all
}

test_plain() {
  box_name=$1
  run_user=$2

  echo "$run_user test in $box_name start <<<<<<<<"
  echo "vagrant setup"
  box=$box_name vagrant up
  k8s_install "$run_user"
  echo "do some test plain"
  #k8s_clean
  echo "$run_user test in $box_name end   >>>>>>>>"
}

test_plain "centos8" "root"
#test_plain "centos7" "root"
#
#test_plain "centos8" "vagrant"
#test_plain "centos7" "vagrant"
#
#test_plain "ubuntu" "vagrant"
