#!/bin/bash -e
# Licensed to SkyAPM org under one or more contributor
# license agreements. See the NOTICE file distributed with
# this work for additional information regarding copyright
# ownership. SkyAPM org licenses this file to you under
# the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

PLUGINS_HOME=$(cd "$(dirname "$0")";cd ..;pwd)
LINTER=${PLUGINS_HOME}/bin/golangci-lint
LINTER_CONFIG=${PLUGINS_HOME}/golangci.yml

PKGS=""

function lint() {
    for d in $(find * -name 'go.mod'); do
        pushd $(dirname $d)
        echo "🟢 golangci lint `sed -n 1p go.mod|cut -d ' ' -f2`"
        ${LINTER} run --timeout=10m --exclude-use-default=false --config=${LINTER_CONFIG}
        popd
    done
}

function fix() {
    for d in $(find * -name 'go.mod'); do
        pushd $(dirname $d)
        echo "🟢 golangci fix `sed -n 1p go.mod|cut -d ' ' -f2`"
        ${LINTER} run -v --fix ./...
        popd
    done
}

case $1 in
  fix)
    fix
    ;;
  *)
    lint
    ;;
esac





