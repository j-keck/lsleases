@echo off
NET SESSION >nul 2>&1
IF NOT %ERRORLEVEL% == 0 (
  echo 'execute per right mouse click "Run as Administrator"'
  PAUSE
  EXIT
)
SET BASE_PATH=%~dp0
"%BASE_PATH%nssm.exe" stop lsleases
