# sample_go_code

This is a small Go sample code

It keeps the idea simple:
- one file: `main.go`
- one module file: `go.mod`
- standard library only
- one local CLI
- one HTTP server
- one Telnet server
- Linux information collected directly in Go

## What it demonstrates
- Go can start multiple services at the same time using goroutines.
- The same Go functions can serve local CLI, HTTP APIs, and Telnet commands.
- Go can collect Linux system information without a heavy framework.
- Go code can stay small and readable for management-plane tasks.

## Features
Local CLI:
- `help`
- `show version`
- `show interface`
- `exit`

HTTP:
- `GET /api/version`
- `GET /api/interfaces`

Telnet:
- login: `admin / admin`
- commands:
  - `help`
  - `show version`
  - `show interface`
  - `exit`

## Run
```bash
go run .
```

Run from inside `sample_go_code/`.

Default ports:
- HTTP: `8081`
- Telnet: `2324`

Browser:
- `http://127.0.0.1:8081/api/version`
- `http://127.0.0.1:8081/api/interfaces`

Telnet:
```bash
telnet 127.0.0.1 2324
```

## Why this sample exists
The full project is the real architecture sample.
This smaller sample exists only to explain the Go language direction quickly, without the rest of the product structure.
