#!/bin/bash -e
#
# Copyright 2022 SkyAPM org
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

GO111MODULE=on
PLUGINS_HOME=$(cd "$(dirname "$0")";cd ..;pwd)

LINTER=${PLUGINS_HOME}/bin/golangci-lint
LINTER_CONFIG=${PLUGINS_HOME}/golangci.yml

function test() {
  cd $1
  echo "🟢 testing `sed -n 1p go.mod|cut -d ' ' -f2`"
  go test -v ./...
}

function deps() {
  cd $1
  echo "🟢 download `sed -n 1p go.mod|cut -d ' ' -f2`"
  go get -v -d ./...
}

function lint() {
  cd $1
  echo "🟢 golangci lint `sed -n 1p go.mod|cut -d ' ' -f2`"
  eval '${LINTER} run --timeout=10m --exclude-use-default=false --config=${LINTER_CONFIG}'
}

function fix() {
  cd $1
  echo "🟢 golangci fix `sed -n 1p go.mod|cut -d ' ' -f2`"
  eval '${LINTER} run -v --fix ./...'
}

function print_help(){
    echo "options: deps, test, lint, fix"
}

case $1 in
  deps)
    shift
    deps $@
    ;;
  test)
    shift
    test $@
    ;;
  lint)
    shift
    lint $@
    ;;
  fix)
    shift
    fix $@
    ;;
  *)
    print_help
    ;;
esac