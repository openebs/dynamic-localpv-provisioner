#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]:-"$0"}")")"
ROOT_DIR="$SCRIPT_DIR/.."
CHART_DIR="$ROOT_DIR/deploy/helm/charts"
VALUES_YAML="$CHART_DIR/values.yaml"

NEW_REGISTRY="ghcr.io"
NEW_REPOSITORY="openebs/dev"
COMPONENT="provisioner-localpv"

source "$SCRIPT_DIR/yq_utils.sh"
source "$SCRIPT_DIR/log.sh"

help() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  --registry                                The registry to be updated to.
  --repository                              The repository to be updated to.
  --component                               The component to be updated (provisioner-localpv or pvc-manager).

Examples:
  $(basename "$0") --registry ghcr.io --repository openebs/dev
  $(basename "$0") --registry ghcr.io --repository openebs/dev --component pvc-manager
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case $1 in
    -h|--help)
      help
      exit 0
      ;;
    --registry)
      shift
      NEW_REGISTRY=$1
      shift
      ;;
    --repository)
      shift
      NEW_REPOSITORY=$1
      shift
      ;;
    --component)
      shift
      COMPONENT=$1
      shift
      ;;
    *)
      help
      log_fatal "Unknown option: $1"
      ;;
  esac
done

if [ -z "${NEW_REGISTRY:-}" ]; then
  log_fatal "Missing required flag: --registry"
fi

if [ -z "${NEW_REPOSITORY:-}" ]; then
  log_fatal "Missing required flag: --repository"
fi

if [ "$COMPONENT" = "provisioner-localpv" ]; then
  yq_ibl ".localpv.image.registry = \"$NEW_REGISTRY\"" "$VALUES_YAML"
  yq_ibl ".localpv.image.repository = \"$NEW_REPOSITORY\"" "$VALUES_YAML"
elif [ "$COMPONENT" = "pvc-manager" ]; then
  yq_ibl ".pvcManager.image.registry = \"$NEW_REGISTRY\"" "$VALUES_YAML"
  yq_ibl ".pvcManager.image.repository = \"$NEW_REPOSITORY\"" "$VALUES_YAML"
else
  log_fatal "Unknown component: $COMPONENT. Supported components: provisioner-localpv, pvc-manager"
fi