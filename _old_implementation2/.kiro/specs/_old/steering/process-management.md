# Process Management for son-et

## Recommended: Auto-terminating Execution

The best way to run son-et for testing is to use a command that automatically terminates after a set time:

```bash
# Run for 5 seconds then auto-terminate (recommended)
go run cmd/son-et/main.go samples/xxx > output.log 2>&1 & PID=$!; sleep 5; kill -9 -$PID 2>/dev/null; wait $PID 2>/dev/null; cat output.log

# Shorter version for quick tests (3 seconds)
go run cmd/son-et/main.go samples/xxx > output.log 2>&1 & PID=$!; sleep 3; kill -9 -$PID 2>/dev/null; wait $PID 2>/dev/null; cat output.log
```

**IMPORTANT:** Use `kill -9 -$PID` (with minus sign) to kill the entire process group, not just the `go run` parent process.

This approach:
- Captures the process ID immediately after starting
- Waits for a specified duration
- Kills the entire process group (go run + child processes)
- Displays the output log

## Manual Process Management

If you need to manually find and kill processes:

### Finding Running Processes

```bash
# Search for son-et processes
ps aux | grep "main samples" | grep -v grep
```

### Killing Processes

```bash
# Kill all son-et processes (use with caution - finds by pattern)
ps aux | grep "main samples" | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null

# Better: Kill specific process group (if you have the PID)
kill -9 -$PID 2>/dev/null
```

**Note:** When using `go run`, always use `kill -9 -$PID` (with minus sign) to kill the entire process group.

## Process Identification

When running `go run cmd/son-et/main.go samples/xxx`, the process appears as:
- `/Users/.../Library/Caches/go-build/.../main samples/xxx`
- `/var/folders/.../go-build.../exe/main samples/xxx`

The key pattern is: `main samples`
