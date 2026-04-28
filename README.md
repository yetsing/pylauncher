# pylauncher

Setup Python Standalone Deploy Environment on Windows

## Usage

```shell
> pylaun.exe -h
Usage of pylaun.exe:
  -debug
        debug mode
  -gui
        use gui launcher.exe
  -list-python
        show available python version
  -platform string
        platform(values: x86 x64 arm64) (default "x64")
  -python string
        python version (default "3.12")
  -verbose
        verbose mode
  -version
        show version
```

## Example

```text
> pylaun.exe -version
Version: 0.11.0 (af88c6e 2026-04-28 16:57:27)
> pylaun.exe -list-python
3.5.2
3.5.2.1
3.5.2.2
3.5.3-rc1
3.5.3
3.5.4-rc
...
> pylaun.exe
[INFO] 2026/04/29 01:33:03 ⏳ Installing Python 3.12.10
[INFO] 2026/04/29 01:33:03 📥 Downloading https://www.nuget.org/api/v2/package/python/3.12.10 ...
[INFO] 2026/04/29 01:33:12 📦 Decompressing...
[INFO] 2026/04/29 01:33:16 📋 Use copy & delete cause of move cross device
[INFO] 2026/04/29 01:33:18 🔨 Make pip wrapper
Looking in links: c:\Users\xxx\AppData\Local\Temp\tmpqfu2r3oa
Requirement already satisfied: pip in d:\my_code\another\example2\python3\lib\site-packages (25.0.1)
[INFO] 2026/04/29 01:33:24 🔨 Make activate script
[INFO] 2026/04/29 01:33:24 🔨 Make entrypoint
[ERROR] 2026/04/29 01:33:34 ⚠️ Failed to download launcher.exe from github: Get "https://github.com/yetsing/pylauncher/releases/download/LauncherV1.2.1/cli-64.exe": net/http: TLS handshake timeout
[INFO] 2026/04/29 01:33:34 🔄 Try download launcher.exe from gitee
[INFO] 2026/04/29 01:33:35 ✅️ Done
[INFO] 2026/04/29 01:33:35 🎯 entrypoint main.py
```

运行完成后，当前目录会新增如下文件

```text
.
├── pip_wrapper     改造过的 pip.exe
├── python3         Python 解释器和 Python 包
├── activate.cmd    cmd 激活环境
├── activate.ps1    powershell 激活环境
├── launcher.exe    启动 exe
└── main.py         入口脚本文件
```

运行 `launcher.exe` ，启动方式如下（按优先级排序）

- 运行 `main.py` (non-gui) / `main.pyw` (gui)

- 读取 `main.mod` 记录的模块名，使用 `python -m xxx` 启动

使用 `activate.cmd` / `activate.ps1` 激活环境。

注意：目前 `activate.cmd` / `activate.ps1` 只是简单地设置 PATH 环境变量。

```text
> . .\activate.ps1
> Get-Command pip

CommandType     Name            Version    Source
-----------     ----            -------    ------
Application     pip.exe         0.0.0.0    D:\my_code\another\example2\pip_wrapper\bin\pip.exe
> Get-Command python

CommandType     Name            Version    Source
-----------     ----            -------    ------
Application     python.exe      3.12.10... D:\my_code\another\example2\python3\python.exe
```

## 参考

[Portable Python Bundles on Windows](https://dev.to/treehouse/portable-python-bundles-on-windows-41ac)
