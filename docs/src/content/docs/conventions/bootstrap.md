---
title: "Bootstrap"
description: "Connect configuration, routes and command synchronization."
---

Use this startup order:

1. Validate environment variables.
2. Build command declarations and handle validation errors.
3. Create the session and explicitly set intents.
4. Register all routes, checking every returned error.
5. Attach `router.Handle` and open the session.
6. Synchronize commands with the application ID and chosen guild scope.
7. Wait for shutdown and close the connection.

Use `func run() error` to own the session and `defer session.Close()` after a successful open. Return startup errors from `run`; log the final failure in `main`, so cleanup can run before process exit.

`SyncOptions{GuildID: guildID}` registers in a test server. An empty `GuildID` means global commands. `Delete: true` removes remote commands outside your declaration; enable it only when this process owns the entire command set in that scope.

Follow the [support bot cookbook](/cookbook/support-bot/) for a complete program using this lifecycle.
