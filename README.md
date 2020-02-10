# ✉️ sendmail-telegram

Sendmail-telegram is an implementation of the `sendmail(8)` interface that uses a telegram bot to deliver emails.

A mapping of recipient addresses to telegram chat ids can be configured, as well as a catch-all chat.

## Usage

telegram-sendmail is `sendmail`-compliant (to an extent), to use it simply pipe an rfc2822-compliant email through `stdin`. Recipients can either be specified as positional arguments, or in the email using the `To`, `Cc` and `Bcc` headers.

```bash
$ cat sample.eml | ./sendmail
```

```
Send an email to telegram users through a bot

Usage:
  sendmail [flags]
  sendmail [command]

Available Commands:
  aid         Print recent updates for the bot
  help        Help about any command

Flags:
  -h, --help   help for sendmail

Use "sendmail [command] --help" for more information about a command.

```

## Configuration

Sendmail-telegram needs some configuration to work. The minimum required settings are a Bot API Token, and a ChatID.

A sample configuration file can be found in `sendmail-telegram.SAMPLE.yml`, and will be searched for in:

1. CWD (`.`)
2. `$XDG_CONFIG_HOME`
3. `$HOME/.config`
2. `/etc`

This file needs your Bot API token, and at least one Chat ID to send messages to:

**Bot API Token**: Can be obtained from the botfather. See [the official docs](https://core.telegram.org/bots/api) for details. When you get one, fill it in the default config.

**Chat ID**: For security reasons, telegram bots can only send messages to existing chats, this is, they cannot message you on their own. For this reason, you need to obtain a Chat ID before the bot can send you anything.
 
Fortunately, sendmail-telegram comes with a handy helper for this built-in. To use it, first send something to your bot, and then run `./sendmail aid`. This will print the chat IDs and names of the most recent messages sent to your bot.
