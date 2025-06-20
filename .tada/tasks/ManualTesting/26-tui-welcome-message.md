---
title: TUI welcome message is shown only once per project and can be disabled globally
priority: 1
status: todo
---

# TUI welcome message is shown only once per project and can be disabled globally

1. Delete `.tada/.welcome_shown` if it exists.
2. Run `tada tui` and confirm the welcome message appears.
3. Run `tada tui` again and confirm the message does NOT appear.
4. Set `show_welcome: false` in your global config (`tada config set show_welcome false --global`).
5. Delete `.tada/.welcome_shown` and run `tada tui` again; confirm the message does NOT appear.
6. Set `show_welcome: true` to re-enable.
