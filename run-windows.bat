@echo off
echo Testing Windows terminal compatibility...
echo.

REM Set TERM environment variable for better compatibility
set TERM=xterm-256color

echo Running shotgun CLI...
echo.

REM Try to run the application
if exist "bin\shotgun.exe" (
    bin\shotgun.exe
) else if exist "bin\shotgun" (
    bin\shotgun
) else (
    echo Binary not found. Please run: npm run build:windows
    pause
    exit /b 1
)

pause