#!/bin/bash
# Reclaim leaked netavark IPAM leases on a dedicated podman runner.
#
# netavark 1.14.1 on RHEL 9.6 never releases IPAM leases on `podman rm`, so the
# pool fills up and jobs fail with "failed to find free IP in range". netavark
# recreates /run/containers/networks/ipam.db empty on the next container start,
# so deleting the file is the only way to reclaim leases today.
#
# The delete runs while holding podman's own rootful networking lock
# (/run/lock/netavark.lock, an fcntl lock taken by containers/common before it
# opens this database), so an allocator either finished before us and shows up
# in the fingerprint, or waits and opens the recreated file afterwards.
#
# Temporary mitigation: remove this file, its go:embed declaration, its units
# and the cloud-config block once a RHEL 9.6 errata ships the netavark fix, and
# run `systemctl disable --now reclaim-ipam.timer` on existing runners.

set -uo pipefail

IPAM_DB="${IPAM_DB:-/run/containers/networks/ipam.db}"
NETAVARK_LOCK="${NETAVARK_LOCK:-/run/lock/netavark.lock}"
BUSY_WINDOW="${BUSY_WINDOW:-120}"
LOCK_TIMEOUT="${LOCK_TIMEOUT:-10}"
# Empty means "derive it from the podman networks below".
IPAM_LEASE_THRESHOLD="${IPAM_LEASE_THRESHOLD:-}"

log() {
  logger -t reclaim-ipam "$*" 2>/dev/null || true
}

# lease_pressure prints the number of distinct IPv4-looking strings in the db.
#
# This is a pressure heuristic, not a count of live leases: bbolt keeps freed
# pages, so released addresses still appear. It over-reports and can reclaim
# early, which is the safe direction, but it must not be read as "N leases in
# use". Deduplication is also global, so addresses shared between networks are
# counted once.
lease_pressure() {
  local out rc
  out=$(grep -aoE '([0-9]{1,3}\.){3}[0-9]{1,3}' "$IPAM_DB" 2>/dev/null)
  rc=$?
  if [ "$rc" -gt 1 ]; then
    return 1
  fi
  printf '%s' "$out" | sort -u | wc -l | tr -d ' '
}

ip_to_int() {
  local a b c d
  IFS=. read -r a b c d <<<"$1"
  printf '%s' $(( (a << 24) | (b << 16) | (c << 8) | d ))
}

# pool_capacity prints the usable address count of the smallest IPv4 pool across
# the podman networks, honouring a configured lease range, or returns 1 when it
# cannot be determined. 85% of a /16 is not 85% of a /24, and a small lease
# range inside a large subnet exhausts long before the subnet does.
pool_capacity() {
  local names name entries entry cidr ranges prefix usable smallest=""
  names=$(timeout 10 podman network ls --format '{{.Name}}' 2>/dev/null) || return 1
  for name in $names; do
    entries=$(timeout 10 podman network inspect "$name" \
      --format '{{range .Subnets}}{{.Subnet}}|{{with .LeaseRange}}{{.StartIP}}-{{.EndIP}}{{end}} {{end}}' 2>/dev/null) || return 1
    for entry in $entries; do
      cidr=${entry%%|*}
      ranges=${entry#*|}
      case "$cidr" in
        *:*) continue ;; # IPv6 subnet, not the pool counted here
      esac
      prefix=${cidr##*/}
      case "$prefix" in
        ''|*[!0-9]*) continue ;;
      esac
      [ "$prefix" -ge 1 ] && [ "$prefix" -le 30 ] || continue
      if [ -n "$ranges" ]; then
        usable=$(( $(ip_to_int "${ranges#*-}") - $(ip_to_int "${ranges%-*}") + 1 ))
      else
        usable=$(( (1 << (32 - prefix)) - 2 ))
      fi
      [ "$usable" -gt 0 ] || continue
      if [ -z "$smallest" ] || [ "$usable" -lt "$smallest" ]; then
        smallest=$usable
      fi
    done
  done
  [ -n "$smallest" ] || return 1
  printf '%s' "$smallest"
}

# db_fingerprint prints size and nanosecond mtime of the db. Both change when a
# lease is committed, because boltdb writes the file rather than only mapping it.
db_fingerprint() {
  python3 -c 'import os,sys; st=os.stat(sys.argv[1]); print("%d:%d" % (st.st_size, st.st_mtime_ns))' \
    "$IPAM_DB" 2>/dev/null
}

# clear_under_lock takes podman's networking lock, re-validates the fingerprint
# and deletes the db while still holding it. It prints one of cleared, changed,
# locked or unreadable; the exit status is 0 only for cleared.
clear_under_lock() {
  local expected=$1
  python3 - "$NETAVARK_LOCK" "$LOCK_TIMEOUT" "$IPAM_DB" "$expected" <<'PY'
import fcntl
import os
import sys
import time

lock_path, timeout, db, expected = sys.argv[1], float(sys.argv[2]), sys.argv[3], sys.argv[4]

fd = os.open(lock_path, os.O_RDWR | os.O_CREAT, 0o600)
deadline = time.monotonic() + timeout
while True:
    try:
        fcntl.lockf(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
        break
    except OSError:
        if time.monotonic() >= deadline:
            print("locked")
            sys.exit(3)
        time.sleep(0.05)

try:
    try:
        st = os.stat(db)
    except OSError:
        print("unreadable")
        sys.exit(4)
    if "%d:%d" % (st.st_size, st.st_mtime_ns) != expected:
        print("changed")
        sys.exit(5)
    try:
        os.unlink(db)
    except OSError:
        print("unreadable")
        sys.exit(4)
    print("cleared")
finally:
    os.close(fd)
PY
}

# busy returns 0 when the runner may be running a job. Anything the script
# cannot prove is treated as busy: deleting the db under a live container kills
# that job and wastes the build. Only exited containers count as harmless.
busy() {
  local states recent errfile
  states=$(podman ps -a --format '{{.State}}' 2>/dev/null) || return 0
  if [ -n "$states" ] && grep -qvx 'exited' <<<"$states"; then
    return 0
  fi
  # `podman ps` is empty between a job being accepted and its container being
  # created; events still report containers that have already been removed.
  # --stream=false is required: podman follows events forever by default.
  # stderr is captured because podman 5.4 exits 0 after a journald read error,
  # only logging "Failed to read event", and an empty list must not be read as
  # "no container started" when the read itself failed.
  errfile=$(mktemp) || return 0
  recent=$(timeout 10 podman events --since "${BUSY_WINDOW}s" --stream=false \
    --filter event=create --format '{{.ID}}' 2>"$errfile") || { rm -f "$errfile"; return 0; }
  if [ -s "$errfile" ]; then
    rm -f "$errfile"
    return 0
  fi
  rm -f "$errfile"
  [ -n "$recent" ] && return 0
  return 1
}

[ -f "$IPAM_DB" ] || exit 0

if ! command -v python3 >/dev/null 2>&1; then
  log "python3 is unavailable, cannot take ${NETAVARK_LOCK}, skipping"
  exit 0
fi

leases=$(lease_pressure) || {
  log "cannot read ${IPAM_DB}, skipping"
  exit 0
}
if [ "$leases" -eq 0 ]; then
  log "no leases readable in ${IPAM_DB}, skipping"
  exit 0
fi

if [ -z "$IPAM_LEASE_THRESHOLD" ]; then
  capacity=$(pool_capacity) || {
    log "cannot determine the podman subnet, skipping"
    exit 0
  }
  IPAM_LEASE_THRESHOLD=$(( capacity * 85 / 100 ))
fi
[ "$leases" -ge "$IPAM_LEASE_THRESHOLD" ] || exit 0

before=$(db_fingerprint)
if [ -z "$before" ]; then
  log "cannot fingerprint ${IPAM_DB}, skipping"
  exit 0
fi

if busy; then
  log "skipped: runner busy, ${leases}/${IPAM_LEASE_THRESHOLD} lease strings in ${IPAM_DB}"
  exit 0
fi
sleep 5
if busy; then
  log "skipped: runner became busy, ${leases}/${IPAM_LEASE_THRESHOLD} lease strings in ${IPAM_DB}"
  exit 0
fi

case "$(clear_under_lock "$before")" in
  cleared)
    log "cleared ${IPAM_DB}: ${leases}/${IPAM_LEASE_THRESHOLD} lease strings, no container created in ${BUSY_WINDOW}s and two idle checks 5s apart"
    ;;
  changed)
    log "skipped: the db changed during the idle checks, ${leases}/${IPAM_LEASE_THRESHOLD} lease strings in ${IPAM_DB}"
    ;;
  locked)
    log "skipped: could not take ${NETAVARK_LOCK} within ${LOCK_TIMEOUT}s, ${leases}/${IPAM_LEASE_THRESHOLD} lease strings in ${IPAM_DB}"
    ;;
  *)
    log "skipped: could not read or remove ${IPAM_DB}"
    ;;
esac
