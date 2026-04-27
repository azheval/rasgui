@echo off
setlocal
set MODE=%~1
set SUB=%~2
set SUB2=%~3

if "%MODE% %SUB%"=="cluster list" (
  echo cluster                          : 446eed3f-17e7-4f17-8537-8234a0f6254f
  echo host                             : mock-host
  echo port                             : 1541
  echo name                             : "Local cluster"
  exit /b 0
)

if "%MODE% %SUB%"=="cluster info" (
  echo cluster                          : %~3
  echo host                             : mock-host
  echo port                             : 1541
  echo name                             : "Local cluster"
  exit /b 0
)

if "%MODE% %SUB%"=="session list" (
  echo session                          : 11111111-1111-1111-1111-111111111111
  echo infobase                         : 22222222-2222-2222-2222-222222222222
  echo user-name                        : test-user
  exit /b 0
)

if "%MODE% %SUB%"=="connection list" (
  echo connection                       : 33333333-3333-3333-3333-333333333333
  echo process                          : 44444444-4444-4444-4444-444444444444
  exit /b 0
)

echo mock command: %*
exit /b 0
