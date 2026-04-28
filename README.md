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

运行完成后，当前目录会新增如下文件

```text
.
├── pip_wrapper
├── python3
├── activate.cmd
├── activate.ps1
├── launcher.exe
└── main.py
```

运行 `launcher.exe` ，启动方式如下（按优先级排序）

- 运行 `main.py` (non-gui) / `main.pyw` (gui)

- 读取 `main.mod` 记录的模块名，使用 `python -m xxx` 启动

使用 `activate.cmd` / `activate.ps1` 激活虚拟环境。

## 参考

[Portable Python Bundles on Windows](https://dev.to/treehouse/portable-python-bundles-on-windows-41ac)
