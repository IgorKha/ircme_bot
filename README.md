# ircme_bot

Simple Telegram bot on Go (telego) that builds messages in this format:

`<name> thinks that: <text>`

`<text>` is sent in _italic_ formatting.

`<name>` is resolved with this priority:

1. `username`
2. `first_name`
3. `someone`

## Features

- Handles **inline query** updates.
- Returns one inline result (article) with the final message text.
- Handles `/me <text>` command in regular chat messages:
  - sends a new bot message in the same style;
  - deletes the original user command message.
- Uses dedicated handlers for inline and command flows, wired through a shared update dispatcher.
- Keeps an in-memory usage counter and logs to stdout after every chosen inline result.

## Run

```bash
export TELEGRAM_BOT_TOKEN="<your_bot_token>"
go run ./cmd/bot
```
