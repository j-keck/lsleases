@echo off
TITLE lsleases
SET BASE_PATH=%~dp0
ECHO start sniffer to capture ip addresses ...
"%BASE_PATH%lsleases.exe" -s
