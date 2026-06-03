@echo off
setlocal
chcp 65001 >nul

echo ==========================================
echo Frontend bootstrap
echo ==========================================
echo CWD: %CD%
echo.

echo [Check] node path/version
where node
if errorlevel 1 (
    echo [ERROR] node not found in PATH.
    goto :END
)
node -v
echo.

echo [Check] npm path/version
where npm
if errorlevel 1 (
    echo [ERROR] npm not found in PATH.
    goto :END
)
npm -v
echo.

if not exist node_modules (
    echo [Install] node_modules missing, running npm install...
    call npm install
    if errorlevel 1 (
        echo [ERROR] npm install failed.
        goto :END
    )
) else (
    echo [Install] node_modules already exists, skip install.
)
echo.

echo [Run] npm run dev -- --host localhost --port 5173 --strictPort
call npm run dev -- --host localhost --port 5173 --strictPort

:END
echo.
echo Frontend script finished. ExitCode=%ERRORLEVEL%
pause
endlocal
