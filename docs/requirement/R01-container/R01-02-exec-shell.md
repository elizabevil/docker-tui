# R01-02: Exec Shell Optimization

## Status: Draft

## Feature Summary
Improve container exec shell performance and rendering quality to match binary CLI experience.

## Problem Statement

### Current Issues

1. **Performance Issues**
   - API-based exec is slower than direct binary (docker/podman exec)
   - Each data chunk triggers full UI re-render
   - term.Buffer processes character-by-character inefficiently

2. **Rendering Issues**
   - TTY rendering via term.Buffer has edge cases
   - Cursor position tracking may be inaccurate
   - Terminal resize handling has lag

3. **User Experience**
   - Response time noticeably slower than `docker exec -it`
   - Some escape sequences not handled properly
   - Interactive applications may flicker

## Optimization Strategy

### Phase 1: Binary-First Execution (In Progress)

Use CLI binary directly when available, fallback to API only when necessary.

- Check for matching CLI (docker/podman) based on current Engine
- Execute CLI with `exec -it` for direct TTY
- Fallback to API if binary unavailable

**Status**: Basic infrastructure done, needs suspend/resume implementation

### Phase 2: Rendering Optimization

#### 2.1 Batch Rendering
- Accumulate output chunks before rendering (e.g., 16ms window)
- Use requestAnimationFrame-style throttling
- Only trigger re-render when batch is ready

#### 2.2 term.Buffer Improvements
- Process data in larger chunks, not character-by-character
- Pre-allocate line buffers
- Use strings.Builder for line construction

#### 2.3 Cursor Optimization
- Track cursor position separately from rendering
- Only update cursor on user input or explicit cursor movement sequences

### Phase 3: Direct TTY Bypass (k9s-style)

For maximum performance, bypass the TUI rendering entirely:

1. Suspend TUI (save terminal state)
2. Run CLI exec in foreground
3. Resume TUI on exit

This provides native terminal experience with zero overhead.

**Reference**: k9s uses tcell/tview's Suspend() mechanism

## Implementation Plan

### Priority 1: Binary Execution (Binary Approach)
- [x] CLI detector (detector.go)
- [x] Binary executor (cli.go)
- [x] Config option (ExecCLI)
- [ ] TUI suspend/resume integration
- [ ] Proper signal handling

### Priority 2: Rendering Optimization
- [x] Chunk-based rendering with buffering
- [x] term.Buffer performance improvements
- [ ] Cursor position caching

#### Implemented (Phase 2)
1. **Batch buffer in ExecState** (`internal/tui/state/exec.go`)
   - Added `batchBuf` (strings.Builder) and `batchSize` counter
   - Flushes when batch >= 8192 bytes OR contains newline OR small chunk (<512 bytes)
   - Reduces number of term.Buffer.Write calls

2. **Reader goroutine batching** (`internal/tui/keyboard/container_action.go`)
   - Reader accumulates up to 8192 bytes before sending to channel
   - Reduces channel communication overhead

3. **term.Buffer optimization** (`internal/tui/term/buffer.go`)
   - Added `extractPrintable()` to batch-extract non-control characters
   - Added `writeString()` for efficient string writing
   - Added `handleBackspace()` helper
   - Avoids byte-by-byte processing

### Priority 3: Fallback Improvements
- [ ] API session connection pooling
- [ ] Streaming buffer optimization
- [ ] Escape sequence handling fixes

## Technical Notes

### Current Data Flow

```
ContainerExecCreate API
       ↓
ContainerExecAttach (stream)
       ↓
ExecCh (channel) - one chunk at a time
       ↓
Append() → term.Buffer.Write()
       ↓
renderExecPassthroughPanel()
       ↓
Full UI re-render
```

### Proposed Optimization: Chunked Rendering

```
ContainerExecCreate API
       ↓
ContainerExecAttach (stream)
       ↓
Output buffer (accumulates ~16ms)
       ↓
Single Append() with batched data
       ↓
renderExecPassthroughPanel()
       ↓
Single UI re-render
```

### Proposed: Binary with Suspend

```
User presses 'e'
       ↓
Save terminal state (tcell/tview Suspend)
       ↓
exec.Command("docker", "exec", "-it", id, shell)
       ↓
Direct stdin/stdout/stderr to terminal
       ↓
User interacts directly (no TUI)
       ↓
Exit → Restore terminal state
       ↓
Resume TUI
```

## Configuration

```yaml
docker:
  execCLI: true  # Use CLI binary when available (default: false)
```

## Acceptance Criteria

- [ ] Binary exec works when docker/podman CLI is available
- [ ] Falls back to API when CLI unavailable
- [ ] Rendering performance < 16ms per frame
- [ ] No visible flicker during normal use
- [ ] Cursor position accurate in common applications (vim, less)
- [ ] Terminal resize handled within 100ms
