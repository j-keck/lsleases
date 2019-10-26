@echo off
SET BASE_PATH=%~dp0
START /MIN CMD /C "%BASE_PATH%lsleasesd.exe -webui -webui-addr 127.0.0.1:9999"
