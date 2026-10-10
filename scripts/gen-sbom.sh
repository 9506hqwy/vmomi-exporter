#!/bin/bash
set -uo pipefail

SCRIPTS_PATH=$(dirname "$(readlink -f "$0")")
BASE_PATH=$(dirname "${SCRIPTS_PATH}")

mkdir -p "${BASE_PATH}/artifacts"

find ./cmd -mindepth 1 -type d -print0 | while IFS= read -d "" -r APP; do
    NAME=$(basename "${APP}")
    DATE=$(date '+%Y%m%d%H%M')

    go tool cyclonedx-gomod app -files -json -main "${APP}" -packages -paths -output "${BASE_PATH}/artifacts/${NAME}-${DATE}.sbom.json"
done
