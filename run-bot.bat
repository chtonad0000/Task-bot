@echo off
echo starting bot...

start /b "" "task-service\bin\task-service.exe"
start /b "" "user-service\bin\user-service.exe"
start /b "" "scheduler-service\bin\scheduler-service.exe"
start /b "" "notification-service\bin\notification-service.exe"
start /b "" "bot-service\bin\bot-service.exe" -env bot-service\\config\\config_test.env

echo bot started.
