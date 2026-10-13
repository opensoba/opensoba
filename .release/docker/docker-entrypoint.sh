#!/usr/bin/dumb-init /bin/sh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0
# Copyright (c) The OpenSoba Authors
#
# Container entrypoint for the OpenSoba agent (binary: soba).
#
# Every SOBA_* variable below also accepts the legacy CONSUL_* spelling, so
# existing container setups keep working:
#
#   SOBA_BIND / SOBA_BIND_INTERFACE           -> agent -bind
#   SOBA_CLIENT / SOBA_CLIENT_INTERFACE       -> agent -client
#   SOBA_DATA_DIR                             (default /soba/data)
#   SOBA_CONFIG_DIR                           (default /soba/config)
#   SOBA_LOCAL_CONFIG                         extra JSON config
#   SOBA_UID / SOBA_GID                       run-as ids
#   SOBA_DISABLE_PERM_MGMT                    skip the chown/drop-privileges step
#   SOBA_ALLOW_PRIVILEGED_PORTS               setcap on the binary before dropping privileges

set -e

# env_or SOBA_X -> $SOBA_X if set and non-empty, otherwise $CONSUL_X.
env_or() {
  eval "val=\${$1-}"
  if [ -z "$val" ]; then
    eval "val=\${CONSUL_${1#SOBA_}-}"
  fi
  printf '%s' "$val"
}

iface_ip() {
  ip -o -4 addr list "$1" | head -n1 | awk '{print $4}' | cut -d/ -f1
}

bind_opt=""
iface="$(env_or SOBA_BIND_INTERFACE)"
if [ -n "$iface" ]; then
  addr="$(iface_ip "$iface")"
  if [ -z "$addr" ]; then
    echo "Could not find IP for interface '$iface', exiting" >&2
    exit 1
  fi
  bind_opt="-bind=$addr"
  echo "==> Found address '$addr' for interface '$iface', setting bind option..."
fi

client_opt=""
iface="$(env_or SOBA_CLIENT_INTERFACE)"
if [ -n "$iface" ]; then
  addr="$(iface_ip "$iface")"
  if [ -z "$addr" ]; then
    echo "Could not find IP for interface '$iface', exiting" >&2
    exit 1
  fi
  client_opt="-client=$addr"
  echo "==> Found address '$addr' for interface '$iface', setting client option..."
fi

data_dir="$(env_or SOBA_DATA_DIR)"
[ -n "$data_dir" ] || data_dir=/soba/data
config_dir="$(env_or SOBA_CONFIG_DIR)"
[ -n "$config_dir" ] || config_dir=/soba/config

local_config="$(env_or SOBA_LOCAL_CONFIG)"
if [ -n "$local_config" ]; then
  echo "$local_config" > "$config_dir/local.json"
fi

# If the user passes flags directly, prepend the binary name.
if [ "${1:0:1}" = '-' ]; then
  set -- soba "$@"
fi

# Expand known subcommands.
if [ "$1" = 'agent' ]; then
  shift
  set -- soba agent \
    -data-dir="$data_dir" \
    -config-dir="$config_dir" \
    $bind_opt \
    $client_opt \
    "$@"
elif [ "$1" = 'version' ]; then
  # This needs a special case because there's no help output.
  set -- soba "$@"
elif soba --help "$1" 2>&1 | grep -q "soba $1"; then
  # The exit code can't tell us whether a subcommand exists, so match the help.
  set -- soba "$@"
fi

# Drop privileges when running the agent, unless explicitly disabled.
if [ "$1" = 'soba' ] && [ -z "$(env_or SOBA_DISABLE_PERM_MGMT)" ]; then
  uid="$(env_or SOBA_UID)"
  [ -n "$uid" ] || uid="$(id -u soba)"
  gid="$(env_or SOBA_GID)"
  [ -n "$gid" ] || gid="$(id -g soba)"

  # If the data/config dirs are bind mounted they may be root-owned.
  [ "$(stat -c %u "$data_dir")" = "$uid" ] || chown "$uid:$gid" "$data_dir"
  [ "$(stat -c %u "$config_dir")" = "$uid" ] || chown "$uid:$gid" "$config_dir"

  if [ -n "$(env_or SOBA_ALLOW_PRIVILEGED_PORTS)" ]; then
    setcap "cap_net_bind_service=+ep" /bin/soba
  fi

  set -- su-exec "$uid:$gid" "$@"
fi

exec "$@"
