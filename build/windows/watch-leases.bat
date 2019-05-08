@echo off
TITLE watch for new leases (LSLEASES_VERSION)  
SET BASE_PATH=%~dp0
"%BASE_PATH%lsleases.exe" -w
