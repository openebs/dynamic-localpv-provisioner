#!/usr/bin/env bash

set -eu

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]:-"$0"}")")"
export OPENEBS_NAMESPACE=${OPENEBS_NAMESPACE:-openebs}
TEST_DIR="$SCRIPT_DIR"/../tests

help() {
  cat <<EOF >&2
Usage: $(basename "${0}") [COMMAND] [OPTIONS]

Commands:
  run                          Run the tests.
  load                         Build and load the image into the K8s cluster.
  install                      Install helm chart and wait for it to be ready.
  clean                        Clean the leftovers.

Options:
  -h, --help                   Display this text.

Options for run:
  -r, --reset                  Clean before running the tests.
  -x, --no-cleanup             Don't cleanup after running the tests.
  -b, --build-always           Build and load the images before running the tests. [ By default image is built if not present only ]
  -t, --test-only              Don't install, test only.

Examples:
  $(basename "${0}") run -rxb
EOF
}

echo_err() {
  echo -e "ERROR: $1" >&2
}

needs_help() {
  [ -n "${1:-}" ] && echo_err "$1\n"
  help
  exit 1
}

die() {
  echo_err "FATAL: $1"
  exit 1
}

cleanup() {
  set +e

  echo "Cleaning up test resources"

  if kubectl get nodes 2>/dev/null; then
    kubectl delete deployment -lrole=test -A
    kubectl delete pod -lrole=test --force -A

    sleep 3

    if helm uninstall localpv-provisioner -n "$OPENEBS_NAMESPACE" --ignore-not-found --timeout=1m --wait; then
      kubectl delete pod --force --all -n "$OPENEBS_NAMESPACE"
    fi
  fi

  set -e
  # always return true
  return 0
}

dump_provisioner_logs() {
  NR=$1
  POD=$(kubectl get pods -l app=localpv-provisioner -o jsonpath='{.items[0].metadata.name}' -n "$OPENEBS_NAMESPACE")
  kubectl describe po "$POD" -n "$OPENEBS_NAMESPACE"
  printf "\n\n"
  kubectl logs --tail="${NR}" "$POD" -n "$OPENEBS_NAMESPACE" -c localpv-provisioner
  printf "\n\n"
}

dump_logs() {
  echo "******************** Dynamic LocalPV-Provisioner logs ***************************** "
  dump_provisioner_logs 1000

  echo "get all the pods"
  kubectl get pods -owide --all-namespaces

  echo "get pvc and pv details"
  kubectl get pvc,pv -oyaml --all-namespaces

  echo "get sc details"
  kubectl get sc --all-namespaces -oyaml
}

is_pod_ready(){
  [ "$(kubectl get po "$1" -o 'jsonpath={.status.conditions[?(@.type=="Ready")].status}' -n "$OPENEBS_NAMESPACE")" = 'True' ] &&
  [ "$(kubectl get po "$1" -o 'jsonpath={.metadata.deletionTimestamp}' -n "$OPENEBS_NAMESPACE")" = "" ]
}

is_provisioner_ready(){
  for pod in $localDriver;do
    is_pod_ready "$pod" || return 1
  done
}

wait_for_provisioner_ready() {
  period=120
  interval=1
  i=0
  while [ "$i" -le "$period" ]; do
    localDriver="$(kubectl get pods -l release=localpv-provisioner -o 'jsonpath={.items[*].metadata.name}' -n "$OPENEBS_NAMESPACE")"
    if is_provisioner_ready "$localDriver"; then
      return 0
    fi

    i=$(( i + interval ))
    echo "Waiting for localpv-provisioner to be ready..."
    sleep "$interval"
  done

  echo "Waited for $period seconds, but all pods are not ready yet."
  return 1
}

helm_install() {
  helm install localpv-provisioner ./deploy/helm/charts -n "$OPENEBS_NAMESPACE" --create-namespace --set localpv.image.pullPolicy=Never --set analytics.enabled=false

  wait_for_provisioner_ready

  kubectl get pods -n "$OPENEBS_NAMESPACE"
}

run_test_suit() {
  wait_for_provisioner_ready

  cd "$TEST_DIR"

  echo "running ginkgo test case with coverage"

  # --focus="TEST HOSTPATH EXT4 QUOTA LOCAL PV WITH UNSUPPORTED FILESYSTEM"
  if ! sudo -E env "PATH=${PATH}" ginkgo -v -coverprofile="integration_coverage.txt" -covermode=atomic; then
    dump_logs
    [ "$CLEAN_AFTER" = "true" ] && cleanup
    exit 1
  fi
}

run() {
  [ "$CLEAN_BEFORE" = "true" ] && cleanup

  if [ "$TEST_ONLY" = "false" ]; then
    maybe_load_image
    helm_install
  fi

  run_test_suit

  printf "\n\n"
  echo "######### All test cases passed #########"

  [ "$CLEAN_AFTER" = "true" ] && cleanup
}

load_k3s() {
  if [ "${CI_K3S:-}" = "true" ]; then
    local img="${1:-}"
    if [ -z "${1:-}" ]; then
      repo="$(make image-repo -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
      tag="$(make image-tag -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
      img="$repo:$tag"
    fi
    docker save "$img" | ctr images import -
  fi
}

load_image() {
  make provisioner-localpv-image
  load_k3s "${1:-}"
}

maybe_load_image() {
  local repo tag img did cid

  if [ "$BUILD_ALWAYS" = "true" ]; then
    load_image
    return 0
  fi

  repo="$(make image-repo -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
  tag="$(make image-tag -s -C "$SCRIPT_DIR"/.. 2>/dev/null)"
  img="$repo:$tag"

  did="$(docker image ls --no-trunc --format json | jq -r --arg repo "$repo" --arg tag "$tag" 'select(.Repository == $repo and .Tag == $tag)|.ID')"
  if [ -z "$did" ]; then
    make provisioner-localpv-image
  fi

  if ! [ "${CI_K3S:-}" = "true" ]; then
    return 0
  fi

  cid="$(crictl image --output json | jq -r --arg image "docker.io/$(make image-ref -s -C "$SCRIPT_DIR"/.. 2>/dev/null)" '.images[]|select(.repoTags[0] == $image)|.id')"
  # If image not present, or different to the docker source, then rebuild it!
  if [ -z "$cid" ] || [ "$cid" != "$did" ]; then
    load_image "$img"
    return 0
  fi

  return 0
}


# allow override
if [ -z "${KUBECONFIG:-}" ]
then
  export KUBECONFIG="${HOME}/.kube/config"
fi

COMMAND=
CLEAN_BEFORE="false"
CLEAN_AFTER="true"
BUILD_ALWAYS="false"
TEST_ONLY="false"

while test $# -gt 0; do
  arg="$1"
  case "$arg" in
    run | clean | load | install)
      [ -n "$COMMAND" ] && needs_help "Can't specify two commands"
      COMMAND="$1"
      ;;
    -r | --reset)
      CLEAN_BEFORE="true"
      ;;
    -x | --no-cleanup)
      CLEAN_AFTER="false"
      ;;
    -b | --build-always)
      BUILD_ALWAYS="true"
      ;;
    -t | --test-only)
      TEST_ONLY="true"
      ;;
    -h | --help)
      needs_help
      ;;
    -*)
      singleLetterOpts="${1:1}"
      while [ -n "$singleLetterOpts" ]; do
        case "${singleLetterOpts:0:1}" in
          r)
            CLEAN_BEFORE="true"
            ;;
          x)
            CLEAN_AFTER="false"
            ;;
          b)
            BUILD_ALWAYS="true"
            ;;
          t)
            TEST_ONLY="true"
            ;;
          *)
            needs_help "Unrecognized argument $singleLetterOpts"
            ;;
        esac
        singleLetterOpts="${singleLetterOpts:1}"
      done
      ;;
    *)
      needs_help "Unrecognized argument $1"
      ;;
  esac
  shift
done

case "$COMMAND" in
  clean)
    cleanup
    ;;
  load)
    load_image
    ;;
  install)
    helm_install
    ;;
  run)
    run
    ;;
  *)
    needs_help "Missing Command"
    ;;
esac
