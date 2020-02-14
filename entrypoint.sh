#!/bin/sh

set -e

if [ "$1" = 'conmon' ]; then
    shift
    exec /root/conmon "$@"
fi

exec "$@"