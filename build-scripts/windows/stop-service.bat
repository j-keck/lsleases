@echo off
SET BASE_PATH=%~dp0
"%BASE_PATH%nssm.exe" stop lsleases
