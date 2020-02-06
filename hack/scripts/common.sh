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

k8s_version=1.16.4
goarch=amd64
goos="unknown"
tmp_root=/tmp

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  goos="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  goos="darwin"
fi

if [[ "$goos" == "unknown" ]]; then
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

# fetch k8s API gen tools and make it available under kb_root_dir/bin.
function fetch_kb_tools {
  header_text "fetching kb tools"
  kb_tools_archive_name="kubebuilder-tools-$k8s_version-$goos-$goarch.tar.gz"
  kb_tools_download_url="https://storage.googleapis.com/kubebuilder-tools/$kb_tools_archive_name"

  kb_tools_archive_path="$tmp_root/$kb_tools_archive_name"
  if [ ! -f $kb_tools_archive_path ]; then
    curl -sL ${kb_tools_download_url} -o "$kb_tools_archive_path"
  fi
  tar -zvxf "$kb_tools_archive_path" -C "$tmp_root/"
}


function setup_envs {
  header_text "setting up env vars"

  # Setup env vars
  export PATH=$tmp_root/kubebuilder/bin:$PATH
  export TEST_ASSET_KUBECTL=$tmp_root/kubebuilder/bin/kubectl
  export TEST_ASSET_KUBE_APISERVER=$tmp_root/kubebuilder/bin/kube-apiserver
  export TEST_ASSET_ETCD=$tmp_root/kubebuilder/bin/etcd
  export TEST_DEP=$tmp_root/kubebuilder/init_project
}

function install_kind {
  header_text "Checking for kind"
  if ! is_installed kind ; then
    header_text "Installing kind"
    KIND_DIR=$(mktemp -d)
    pushd $KIND_DIR
    GO111MODULE=on go get sigs.k8s.io/kind@v0.6.0
    popd
  fi
}

function is_installed {
  if command -v $1 &>/dev/null; then
    return 0
  fi
  return 1
}
