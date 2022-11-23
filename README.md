# tcmd

A utility for running commands on other computers via TRIGGERcmd.com

Download the binary for your OS from [here](https://www.triggercmd.com/forum/topic/196/tcmd-go-command-line-tool-is-available-now).

## Quick start

```
tcmd --pair

open C:\Users\fred\.TRIGGERcmdData\token.tkn: The system cannot find the file specified.
No token found.
Within 10 minutes, log into your account at triggercmd.com, click your name in the upper-right, click Pair, and type in this pair code:
X1XBX

Waiting..........Token saved.
Go ahead and run something like:  tcmd --list
```

## Usage

```
tcmd.exe -h
NAME:
    tcmd - Run commands on computers in your TRIGGERcmd account
  USAGE:
    tcmd.exe [options]

  OPTIONS:
    --trigger value, -t value   Trigger name of the command you want to run
    --computer value, -c value  Name of the computer (leave blank for your default computer)
    --params value, -p value    Any parameters you want to add to the remote command
    --panel value, -P value     Name of the panel you want to use
    --button value, -b value    Name of the panel button to "press"
    --list, -l                  List your commands
    --listpanels, -L            List your panels
    --pair                      Login using a pair code
    --help, -h                  show help
    --version, -v               print the version
```
