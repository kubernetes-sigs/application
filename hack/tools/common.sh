#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

arch=amd64
os="unknown"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  os="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  os="darwin"
fi

if [[ "$os" == "unknown" ]]; then
  echo "OS '$OSTYPE' not supported. Aborting." >&2
  exit 1
fi

go_workspace=''
for p in ${GOPATH//:/ }; do
  if [[ $PWD/ = $p/* ]]; then
    go_workspace=$p
  fi
done

if [ -z $go_workspace ]; then
  echo 'Current directory is not in $GOPATH' >&2
  exit 1
fi

# Turn colors in this script off by setting the NO_COLOR variable in your
# environment to any value:
#
# $ NO_COLOR=1 test.sh
NO_COLOR=${NO_COLOR:-""}
if [ -z "$NO_COLOR" ]; then
  header=$'\e[1;33m'
  reset=$'\e[0m'
else
  header=''
  reset=''
fi

function header_text {
  echo "$header$*$reset"
}
