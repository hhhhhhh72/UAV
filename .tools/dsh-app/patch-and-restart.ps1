$ErrorActionPreference = 'Continue'
Start-Sleep -Seconds 8
$exe = 'C:\Users\21125\AppData\Local\Programs\dsh-desktop\DSH-Desktop.exe'
$log = 'D:\w-yao\.tools\dsh-app\patch-log.txt'
"$(Get-Date -Format o) [v2] patch start" | Out-File $log -Append -Encoding utf8

# 等 exe 文件解锁（进程停止后窗口期）
$tries = 0
while ($tries -lt 30) {
    try {
        $fs = [System.IO.File]::Open($exe, 'Open', 'ReadWrite', 'None')
        $fs.Close()
        break
    } catch {
        $tries++
        Start-Sleep -Seconds 1
    }
}
"$(Get-Date -Format o) [v2] exe unlocked after $tries tries" | Out-File $log -Append -Encoding utf8

# 打补丁：独立 JS（cwd=.tools 保证 require('rcedit') 可解析）
Push-Location 'D:\w-yao\.tools'
node 'D:\w-yao\.tools\patch-icons.js' *>> $log
Pop-Location

"$(Get-Date -Format o) [v2] patch done, relaunching" | Out-File $log -Append -Encoding utf8
Start-Process $exe
"$(Get-Date -Format o) [v2] relaunched" | Out-File $log -Append -Encoding utf8
