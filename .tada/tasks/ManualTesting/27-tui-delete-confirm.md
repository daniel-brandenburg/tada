---
title: TUI delete confirmation dialog prevents accidental deletion
priority: 1
status: todo
---

# TUI delete confirmation dialog prevents accidental deletion

1. Select a task in the TUI and press `d`.
2. Confirm that a dialog appears asking to confirm deletion.
3. Press `n` or `esc` to cancel; the task should remain.
4. Press `d` again, then `y` to confirm; the task should be deleted.
5. Verify that accidental keypresses do not delete tasks without confirmation.
