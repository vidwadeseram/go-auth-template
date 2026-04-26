#!/bin/sh

set -eu

host="${1:-db}"
port="${2:-5432}"

until nc -z "$host" "$port"; do
  sleep 1
done
