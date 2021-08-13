#!/usr/bin/env bash

VAGRANT_SH_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd -P)"
ENV_FILE=$VAGRANT_SH_PATH/.vagrant.env

export set_provider=${provider}
export set_box=${box}

if [[ -f "${ENV_FILE}" ]]; then
  source $ENV_FILE
fi

export provider=${set_provider:-"${env_provider:-"virtualbox"}"}
export box=${set_box:-"${env_box:-"centos8"}"}

cat <<EOF | tee $ENV_FILE
export env_provider=${provider}
export env_box=${box}
EOF

if [ ! "$#" == "0" ]; then
  cd $VAGRANT_SH_PATH/..
  vagrant $@
fi
