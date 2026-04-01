# Program Flow

This note explains how `main.go` runs and how requests flow through the sample.

Main source file:
- `main.go`

## 1. Program startup

When we run:

```bash
go run .
```

Go starts:
- `main()` in `main.go:34`

Inside `main()` the code does this in order:

1. Read HTTP and Telnet port values
   - `main.go:35`
   - `main.go:36`

2. Start the HTTP server in a goroutine
   - `main.go:38`

3. Start the Telnet server in another goroutine
   - `main.go:50`

4. Start the local CLI in the main thread
   - `main.go:57`

Simple startup flow:

`main() -> start HTTP -> start Telnet -> start local CLI`

## 2. Local CLI flow

The local CLI starts here:
- `runLocalCLI()` in `main.go:126`

Then:
1. Print local CLI banner
2. Call `runSession()` in `main.go:131`
3. Show prompt
4. Read the user command
5. Call `handleCommand()` in `main.go:150`

Example: `show version`

Flow:
- `runLocalCLI()`
- `runSession()`
- `handleCommand()`
- `versionInfo()` in `main.go:174`
- `formatVersion()` in `main.go:215`

Example: `show interface`

Flow:
- `runLocalCLI()`
- `runSession()`
- `handleCommand()`
- `interfaceInfo()` in `main.go:185`
- `formatInterfaces()` in `main.go:227`

## 3. HTTP / Web flow

The HTTP server is started inside the first goroutine:
- `main.go:38`

Registered handlers:
- `/` -> `handleRoot()` in `main.go:60`
- `/api/version` -> `handleVersion()` in `main.go:68`
- `/api/interfaces` -> `handleInterfaces()` in `main.go:72`

Example: browser opens `/api/version`

Flow:
- HTTP request reaches `handleVersion()`
- `handleVersion()` calls `versionInfo()`
- result is returned by `writeJSON()` in `main.go:76`

Example: browser opens `/api/interfaces`

Flow:
- HTTP request reaches `handleInterfaces()`
- `handleInterfaces()` calls `interfaceInfo()`
- result is returned by `writeJSON()`

Simple HTTP flow:

`HTTP request -> handler -> shared data function -> JSON response`

## 4. Telnet flow

The Telnet server starts here:
- `serveTelnet()` in `main.go:85`

Flow:
1. Open TCP listener on the Telnet port
2. Wait for incoming connection
3. Accept a connection
4. Start `handleTelnet()` in `main.go:102` in a goroutine

Inside `handleTelnet()`:
1. Ask for username
2. Ask for password
3. Validate `admin / admin`
4. Start the same session engine used by the local CLI:
   - `runSession()` in `main.go:131`

Example: Telnet user enters `show version`

Flow:
- `serveTelnet()`
- `handleTelnet()`
- `runSession()`
- `handleCommand()`
- `versionInfo()`
- `formatVersion()`

Example: Telnet user enters `show interface`

Flow:
- `serveTelnet()`
- `handleTelnet()`
- `runSession()`
- `handleCommand()`
- `interfaceInfo()`
- `formatInterfaces()`

## 5. Shared logic idea

The most important design point in this sample is:

- Local CLI and Telnet share the same command/session engine
- HTTP uses different handlers
- All of them reuse the same data collection functions

Shared functions:
- `versionInfo()` in `main.go:174`
- `interfaceInfo()` in `main.go:185`

This is the core message of the sample:

`one Go program -> multiple interfaces -> shared backend logic`

## 6. Very short summary

- `main()` starts HTTP, Telnet, and local CLI
- Local CLI and Telnet share the same command loop
- HTTP uses API handlers
- All interfaces reuse the same Linux data functions
