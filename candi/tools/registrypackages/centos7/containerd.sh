# Copyright 2021 Flant CJSC
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

#!/usr/bin/env bash

set -Eeo pipefail

if [[ -z ${REGISTRY} ]]; then
    >&2 echo "ERROR: REGISTRY is not set"
    exit 1
fi

REGISTRY_PATH="${REGISTRY}/deckhouse/binaries"

for VERSION in $(yq r /deckhouse/candi/version_map.yml -j | jq -r '.k8s | .[].bashible.centos."7"' | grep desiredVersion | grep containerd | awk '{print $2}' | tr -d '"' | sort | uniq); do
  PACKAGE="$(sed "s/containerd.io-/containerd.io:/" <<< "${VERSION}")"
  mkdir package
  pushd package
  # Centos
  # get url with yumdownloader --urls
  RPM_PACKAGE="https://download.docker.com/linux/centos/7/x86_64/stable/Packages/${VERSION}.rpm"
  RPM_PACKAGE_SELINUX="http://mirror.centos.org/centos/7/extras/x86_64/Packages/container-selinux-2.119.2-1.911c772.el7_8.noarch.rpm"
  wget ${RPM_PACKAGE} ${RPM_PACKAGE_SELINUX}
  RPM_PACKAGE_FILE="$(ls containerd.io*)"
  RPM_PACKAGE_SELINUX_FILE="$(ls container-selinux*)"
  popd

  cat <<EOF > package/install
#!/bin/bash
set -Eeo pipefail
rpm -U ${RPM_PACKAGE_SELINUX_FILE} ${RPM_PACKAGE_FILE}
yum versionlock add containerd.io container-selinux
EOF
  chmod +x package/install

  cat <<EOF > package/uninstall
#!/bin/bash
set -Eeo pipefail
yum versionlock delete containerd.io container-selinux
rpm -e ${RPM_PACKAGE_FILE%.rpm} ${RPM_PACKAGE_SELINUX_FILE%.rpm}
EOF
  chmod +x package/uninstall

  cat <<EOF > Dockerfile
FROM scratch
COPY ./package/* /
EOF

  docker build -t ${REGISTRY_PATH}/${PACKAGE} .
  docker push ${REGISTRY_PATH}/${PACKAGE}
  rm -rf package Dockerfile
done
