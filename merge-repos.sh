#!/usr/bin/env bash

set -euo pipefail

# create tmp folder
WORK_DIR=`mktemp -d`

function cleanup {      
  rm -rf "$WORK_DIR"
  echo "Deleted temp working directory $WORK_DIR"
}

trap cleanup EXIT

git clone git@github.com:grafana/pyroscope.git "${WORK_DIR}/og"
git clone git@github.com:grafana/phlare.git    "${WORK_DIR}/phlare"


# rewrite phlare history to maintain correct links
cd "$WORK_DIR/phlare"
git filter-repo --message-callback '
import re
return re.sub(b"#([0-9]+)", b"https://github.com/grafana/phlare/issues/\\1", message)
'

# move import path to new repo's
git ls-files '*.go' go.mod go.sum api/go.mod api/go.sum ebpf/go.mod ebpf/go.sum | xargs sed -i 's#github.com/grafana/phlare#github.com/grafana/pyroscope#g'
go mod tidy
git add -A .
git commit -m "Rename go import path"

# move og into subfolder
cd "$WORK_DIR/og"
mkdir -p ../temp
mv * ../temp
mv .* ../temp
mv ../temp/.git .
mv ../temp og/
git add -A .
git commit -m "Move og-pyroscope into subfolder"

# now merge phlare in
git remote add phlare "${WORK_DIR}/phlare"
git fetch phlare
git merge phlare/main --allow-unrelated-histories -m "The new Pyroscope"

git push git@github.com:simonswine/pyroscope HEAD:merged -f
