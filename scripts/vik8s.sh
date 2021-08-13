#!/usr/bin/env bash
set -e

SCRIPTS_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd -P)"
cd $SCRIPTS_PATH/..
./bin/vik8s -f ./bin $@
