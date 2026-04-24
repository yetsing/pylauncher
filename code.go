package main

const (
	makePipCode = `
from pip._vendor.distlib.scripts import ScriptMaker
maker = ScriptMaker("pip_wrapper/scripts", "pip_wrapper/bin")
maker.executable = r"python.exe"
maker.make("pip.py")
`
	pipCode = `
#!/usr/bin/python
import sys
import os

if __name__ == "__main__":
    from pip._vendor.distlib.scripts import ScriptMaker
    ScriptMaker.executable = r"python.exe"

    from pip._internal.cli.main import main
    sys.exit(main())
`
	activateCmd = `
@echo off
set PATH=%~dp0pip_wrapper\bin\;%~dp0python3\Scripts\;%~dp0python3\;%PATH%
`
	activatePs1 = `
$ScriptDir = (Split-Path -Parent $MyInvocation.MyCommand.Definition)
$Env:PATH = "$ScriptDir\pip_wrapper\bin;$ScriptDir\python3\Scripts;$ScriptDir\python3;$Env:PATH"
`
)
