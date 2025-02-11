@echo off
echo stopping bot...

taskkill /F /IM bot-service.exe
taskkill /F /IM task-service.exe
taskkill /F /IM user-service.exe
taskkill /F /IM scheduler-service.exe
taskkill /F /IM notification-service.exe

echo bot stopped
pause
