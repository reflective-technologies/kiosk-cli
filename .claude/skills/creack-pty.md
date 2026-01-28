# creack/pty

Go package that provides helpers for working with Unix pseudo-terminals (PTYs). Use it when a subprocess needs a real TTY (shells, TUIs, password prompts, color output, etc.).

## When To Use

- Running interactive commands that behave differently without a TTY
- Driving shells or REPLs from Go
- Capturing TTY-specific output (ANSI color, cursor control)

## Install

```bash
go get github.com/creack/pty
```

## Core API

- `pty.Start(cmd)` - Start `*exec.Cmd` with a controlling PTY and return the PTY file
- `pty.StartWithSize(cmd, ws)` - Start with a specific terminal size
- `pty.StartWithAttrs(cmd, ws, attrs)` - Start with size and custom `SysProcAttr`
- `pty.Open()` - Open a PTY/TTY pair
- `pty.Setsize(pty, ws)` / `pty.Getsize(pty)` - Set or read terminal size
- `pty.InheritSize(pty, tty)` - Apply PTY size to TTY (use on `SIGWINCH`)

## Common Pattern

```go
cmd := exec.Command("bash")
ptmx, err := pty.Start(cmd)
if err != nil { /* handle */ }
defer ptmx.Close()

// Optional: put local terminal in raw mode
oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
if err != nil { /* handle */ }
defer term.Restore(int(os.Stdin.Fd()), oldState)

// Copy input/output
// go io.Copy(ptmx, os.Stdin)
// go io.Copy(os.Stdout, ptmx)
```

## Resize Handling

- Watch `SIGWINCH` and call `pty.InheritSize(ptmx, os.Stdout)` or `pty.Setsize` with a computed `Winsize`
- Initialize size once before starting, then update on resize

## Pitfalls

- Always restore terminal state on exit when using raw mode
- PTY operations are Unix-focused; unsupported platforms return `ErrUnsupported`
- If you need `Close()` to interrupt `Read()` with deadlines, you may need to set non-blocking mode manually
