#!/bin/sh
# OpenEBS dynamic-localpv-provisioner quota helper.
#
# Applies or removes a filesystem project quota on a directory whose parent
# resides on an XFS or EXT2/3/4 filesystem with project quotas enabled.
#
# Usage:
#   quota.sh apply   --parent <path> --volume <path> --soft-kb <n> --hard-kb <n>
#   quota.sh cleanup --parent <path> --volume <path>

set -e

die() { echo "$@" >&2; exit 1; }

usage() {
    cat >&2 <<'EOF'
Usage:
  quota.sh apply   --parent <path> --volume <path> --soft-kb <n> --hard-kb <n>
  quota.sh cleanup --parent <path> --volume <path>
EOF
    exit 2
}

mode=
parent=
volume=
soft_kb=
hard_kb=

[ $# -ge 1 ] || usage
mode=$1; shift

while [ $# -gt 0 ]; do
    case "$1" in
        --parent)  [ $# -ge 2 ] || die "--parent needs a value";  parent=$2;  shift 2 ;;
        --volume)  [ $# -ge 2 ] || die "--volume needs a value";  volume=$2;  shift 2 ;;
        --soft-kb) [ $# -ge 2 ] || die "--soft-kb needs a value"; soft_kb=$2; shift 2 ;;
        --hard-kb) [ $# -ge 2 ] || die "--hard-kb needs a value"; hard_kb=$2; shift 2 ;;
        *) die "unknown argument: $1" ;;
    esac
done

[ -n "$parent" ] || die "missing --parent"
[ -n "$volume" ] || die "missing --volume"

fs=$(stat -f -c %T "$volume" 2>/dev/null || echo "unknown")

case "$mode" in
    apply)
        [ -n "$soft_kb" ] || die "missing --soft-kb"
        [ -n "$hard_kb" ] || die "missing --hard-kb"
        case "$fs" in
            xfs)
                pid=$(xfs_quota -x -c 'report -h' "$parent" 2>/dev/null | tail -2 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
                pid=$((pid + 1))
                xfs_quota -x -c "project -s -p $volume $pid" "$parent"
                xfs_quota -x -c "limit -p bsoft=${soft_kb}k bhard=${hard_kb}k $pid" "$parent"
                ;;
            ext2/ext3)
                pid=$(repquota -P "$parent" 2>/dev/null | tail -3 | awk 'NR==1{print substr ($1,2)}+0' || echo "0")
                pid=$((pid + 1))
                chattr +P -p "$pid" "$volume"
                # setquota -P expects block limits as plain numbers (in KB blocks).
                setquota -P "$pid" "$soft_kb" "$hard_kb" 0 0 "$parent"
                ;;
            *)
                echo "Unsupported filesystem type: $fs" >&2
                rm -rf "$volume"
                exit 1
                ;;
        esac
        ;;
    cleanup)
        case "$fs" in
            xfs)
                id=$(xfs_io -c stat "$volume" 2>/dev/null | awk '/projid/{print $3}' | head -1)
                echo "projid=$id"
                if [ -n "$id" ] && [ "$id" != "0" ]; then
                    xfs_io -c "chproj -R 0" "$volume" 2>/dev/null || true
                    xfs_quota -x -c "limit -p bsoft=0 bhard=0 $id" "$parent" 2>/dev/null || true
                fi
                ;;
            ext2/ext3)
                id=$(lsattr -pd "$volume"/ 2>/dev/null | awk '{print $1}')
                if [ -n "$id" ] && [ "$id" != "0" ]; then
                    setquota -P "$id" 0 0 0 0 "$parent" 2>/dev/null || true
                fi
                ;;
        esac
        rm -rf "$volume"
        ;;
    *)
        die "unknown mode: $mode (expected 'apply' or 'cleanup')"
        ;;
esac
