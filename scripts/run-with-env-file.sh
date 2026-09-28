#!/bin/sh
set -eu

if [ "$#" -ne 1 ]; then
    echo "usage: run-with-env-file.sh /path/to/avito-mcp" >&2
    exit 2
fi

env_file=${AVITO_MCP_ENV_FILE:-"$HOME/.config/avito-mcp/env"}
if [ ! -r "$env_file" ]; then
    echo "Avito MCP credential file is missing or unreadable: $env_file" >&2
    exit 1
fi

set -a
. "$env_file"
set +a

exec "$1"
