@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul

:: Version: 1.4
set "ROOT=%~dp0"

echo ==========================================
echo [0/3] Cleaning up existing processes...
echo ==========================================

:: Try killing by ports first
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8000 ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8080 ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :5173 ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1

:: Kill common process names
taskkill /F /IM node.exe /T >nul 2>&1
taskkill /F /IM uvicorn.exe /T >nul 2>&1
taskkill /F /IM python.exe /T >nul 2>&1
taskkill /F /IM main.exe /T >nul 2>&1

timeout /t 2 >nul

echo [1/3] Starting Agent Service (FastAPI) on :8000 ...
start "AI Agent" /D "%ROOT%agent_py" cmd /k "python -m uvicorn main:app --host 127.0.0.1 --port 8000"

timeout /t 2 >nul

echo [2/3] Starting Backend Service (Gin) on :8080 ...
start "Backend" /D "%ROOT%backend_go" cmd /k "go run ."

timeout /t 2 >nul

echo [3/3] Starting Frontend Service (Vue) on :5173 ...
start "Frontend" /D "%ROOT%front_vue" cmd /k "call start_frontend.bat"

echo Browser auto-open watcher started in background.
start "FrontendWatcher" /min powershell -NoProfile -Command "$ok=$false; 1..45 | ForEach-Object { try { $r=Invoke-WebRequest -Uri 'http://localhost:5173/' -UseBasicParsing -TimeoutSec 2; if($r.StatusCode -ge 200){$ok=$true; break} } catch {}; Start-Sleep -Seconds 1 }; if($ok){ Start-Process 'http://localhost:5173/' }"

echo.
echo ==========================================
echo 所有服务已尝试启动 (V1.4) / All services starting.
echo ==========================================
echo - Agent:    http://127.0.0.1:8000
echo - Backend:  http://127.0.0.1:8080
echo - Frontend: http://localhost:5173
echo ==========================================
echo.
echo 请在单独的窗口中检查控制台输出。
echo.

endlocal
