@echo off
echo Building shotgun-cli for Windows and WSL...

REM Check if running in WSL
set "IS_WSL="
if "%WSL_DISTRO_NAME%" NEQ "" set IS_WSL=true
if "%WSL_INTEROP%" NEQ "" set IS_WSL=true

echo Building main Windows executable...
call npm run build:windows
if %errorlevel% neq 0 (
    echo Build failed. Exiting.
    exit /b %errorlevel%
)

echo Creating wrapper script...
call node create-wrapper.js
if %errorlevel% neq 0 (
    echo Wrapper creation failed. Exiting.
    exit /b %errorlevel%
)

echo Installing globally...
call npm install -g .
if %errorlevel% neq 0 (
    echo Global installation failed. Exiting.
    exit /b %errorlevel%
)

if defined IS_WSL (
    echo.
    echo WSL detected. Attempting to make the binary accessible from WSL.
    REM Further WSL-specific steps might be needed here,
    REM such as ensuring the .exe is in a Windows PATH directory accessible from WSL,
    REM or creating a symlink/alias within WSL.
    REM For now, we assume the global npm install makes it available.
    echo Make sure your Windows PATH is configured in WSL's /etc/wsl.conf (usually it is by default).
)

echo.
echo Done! You can now use 'shotgun' or 'shotgun-cli' commands.
echo.
echo For best compatibility:
echo   - Windows: Use Windows Terminal or PowerShell.
echo   - WSL: Use your WSL terminal.
