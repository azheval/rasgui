#!/usr/bin/env sh
mode="$1"
sub="$2"

if [ "$mode $sub" = "cluster list" ]; then
  printf '%s\n' 'cluster                          : 446eed3f-17e7-4f17-8537-8234a0f6254f'
  printf '%s\n' 'host                             : mock-host'
  printf '%s\n' 'port                             : 1541'
  printf '%s\n' 'name                             : "Local cluster"'
  exit 0
fi

if [ "$mode $sub" = "cluster info" ]; then
  printf '%s\n' "cluster                          : $3"
  printf '%s\n' 'host                             : mock-host'
  printf '%s\n' 'port                             : 1541'
  printf '%s\n' 'name                             : "Local cluster"'
  exit 0
fi

if [ "$mode $sub" = "session list" ]; then
  printf '%s\n' 'session                          : 11111111-1111-1111-1111-111111111111'
  printf '%s\n' 'infobase                         : 22222222-2222-2222-2222-222222222222'
  printf '%s\n' 'user-name                        : test-user'
  exit 0
fi

if [ "$mode $sub" = "connection list" ]; then
  printf '%s\n' 'connection                       : 33333333-3333-3333-3333-333333333333'
  printf '%s\n' 'process                          : 44444444-4444-4444-4444-444444444444'
  exit 0
fi

printf '%s\n' "mock command: $*"
exit 0
