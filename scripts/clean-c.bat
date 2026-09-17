@echo off
:: C drive cleanup launcher - double click to run (auto requests admin)
chcp 65001 >nul
title C-Drive-Cleanup
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0clean-c.ps1"
pause
