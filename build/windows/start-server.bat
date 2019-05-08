@echo off
SET BASE_PATH=%~dp0
START /MIN CMD /C "%BASE_PATH%lsleases.exe" -s
