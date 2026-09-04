#!/usr/bin/env bash
# Network impairment helpers for local resilience testing (Linux tc/netem).
# Requires root (or CAP_NET_ADMIN) and an interface name.
#
# Usage:
#   ./network_impairment.sh apply eth0
#   ./network_impairment.sh clear eth0

set -euo pipefail

ACTION="${1:-}"
IFACE="${2:-}"

if [[ -z "${ACTION}" || -z "${IFACE}" ]]; then
  echo "Usage: $0 {apply|clear} <iface>" >&2
  exit 1
fi

case "${ACTION}" in
  apply)
    # 100ms delay ±20ms, 2% packet loss — simulates intermittent uplink.
    tc qdisc add dev "${IFACE}" root netem delay 100ms 20ms loss 2%
    echo "Applied netem on ${IFACE}: delay 100ms±20ms, loss 2%"
    ;;
  clear)
    tc qdisc del dev "${IFACE}" root 2>/dev/null || true
    echo "Cleared qdisc on ${IFACE}"
    ;;
  *)
    echo "Unknown action: ${ACTION}" >&2
    exit 1
    ;;
esac
