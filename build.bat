@echo off
echo Building shotgun-cli for Windows...
call npm run build:windows
call node create-wrapper.js
echo Installing globally...
call npm install -g .
echo.
echo Done! You can now use 'shotgun' or 'shotgun-cli' commands.
echo.
echo For best compatibility, use Windows Terminal or PowerShell.
