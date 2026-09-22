# Facts: Claude Code Message-Level Deduplication

1. **Transcript Log Mechanics**:
   - Claude Code session logs are stored at `~/.claude/projects/*/*.jsonl`.
   - Each line represents an event (`type: "assistant"`, `type: "user"`, `type: "tool_use"`, etc.).
   - An assistant message event contains:
     ```json
     {
       "type": "assistant",
       "timestamp": "2026-09-21T18:00:00Z",
       "message": {
         "id": "msg_01X...",
         "model": "claude-sonnet-4-20250514",
         "usage": {
           "input_tokens": 1200,
           "output_tokens": 450,
           "cache_read_input_tokens": 45000,
           "cache_creation_input_tokens": 1500
         }
       }
     }
     ```
2. **Double-Counting Vectors**:
   - Streaming flush lines or tool retry attempts can record identical `message.id` entries with duplicated usage.
   - Resumed or compacted sessions can re-serialize prior turns.
3. **Thermal's Existing Dedupe Patterns**:
   - Codex (`codex.go`): Takes the last `token_count` snapshot per thread.
   - Grok (`grok.go`): Filters strictly for `turn_completed`.
   - Claude Code (`claude.go`): Currently parses every `rec.Type == "assistant"` without checking `message.id`.
