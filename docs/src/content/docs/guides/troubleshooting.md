---
title: "Troubleshooting"
description: "Find the failing step before changing configuration."
---

| Symptom | Check |
| --- | --- |
| 401 / Unauthorized | Reset invalid tokens; use the bot token, not the public key |
| 403 / Missing Access during sync | Application installed in the chosen guild; correct server ID |
| Command not listed | Scope, successful synchronization and command permissions |
| Gateway 4014 | Requested privileged intents enabled in the portal; reduce intents if unused |
| Application did not respond | Handler logs, matching route, prompt acknowledgement |
| Already responded | Only one initial response; use Edit or Followup afterward |
| Component rejected | MessageSpec validation; V2 cannot combine legacy embeds or polls |

Keep registration and synchronization errors in startup logs. For slow work, acknowledge with `Defer` and finish with `Edit`. Never log the full interaction object: it can contain interaction tokens and user input.

If two copies of the bot are running, stop the unintended instance before debugging responses. Reproduce the issue in a test server with one process and one command.
