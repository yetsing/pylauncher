# pylauncher

Python Standalone Deploy Environment

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

运行 `launcher.exe` ：他会使用 `python3\python.exe` 运行同目录下的 `main.py` 。

使用 `activate.cmd` / `activate.ps1` 激活虚拟环境。

## 参考

[Portable Python Bundles on Windows](https://dev.to/treehouse/portable-python-bundles-on-windows-41ac)
