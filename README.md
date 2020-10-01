[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/local-functions)](https://goreportcard.com/report/github.com/Luzifer/local-functions)
![](https://badges.fyi/github/license/Luzifer/local-functions)
![](https://badges.fyi/github/downloads/Luzifer/local-functions)
![](https://badges.fyi/github/latest-release/Luzifer/local-functions)
![](https://knut.in/project-status/local-functions)

# Luzifer / local-functions

`local-functions` is intended as the opposite of Cloud-Functions: Run scripts on the local machine through HTTP calls.

**Be aware:** This will expose scripts in a certain folder on your machine. This might cause trouble for you! So you really should only expose the server on **localhost** and ensure nobody else is able to access the API. And **never ever** run this as root! (Or say good bye to your system!)

## Examples

```console
# curl -d '{"test": "foo"}' -H 'Content-Type: application/json' -X POST localhost:3000/echo
PWD=/home/luzifer/workspaces/private/go/src/github.com/Luzifer/local-functions
ACCEPT=*/*
SHLVL=1
CONTENT_TYPE=application/json
METHOD=POST
_=/usr/bin/env

=====

{"test": "foo"}
```
