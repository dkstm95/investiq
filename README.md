# gacha

Ask investment questions from your terminal.

`gacha` is a small terminal app for investment research.
You ask a plain question.
Gacha turns it into a structured AI research workflow.

Investment outcomes cannot be predicted perfectly.
Careful research can still improve the odds.
Gacha helps you check current data, compare alternatives, surface risks,
and turn unclear questions into cleaner decisions.

Korean: [docs/ko/README.md](docs/ko/README.md)

## TL;DR

Install Gacha, run `gch`, answer the one-time research profile, and ask an
investment question.

```bash
curl -fsSL https://raw.githubusercontent.com/dkstm95/gacha/main/install.sh | sh
"$HOME/.local/bin/gch"
```

Gacha is for research, not trading automation.
It helps the AI check current data, compare alternatives, surface risks, and
produce a report you can revisit.

## What It Helps With

Gacha supports different stages of investment clarity:

- Discover: you do not know what or how to invest in yet.
- Theme: you know a field, but not the specific stock, ETF, or asset.
- Entry: you know what you may buy, but not when or under what conditions.
- Holding: you own something and need hold, trim, sell, or stop-out rules.
- Portfolio: you need allocation, concentration, overlap, or risk review.

The goal is not certainty.
The goal is faster, calmer, more consistent research.

## What It Does

- Opens a simple terminal workspace for investment questions.
- Asks the AI to check current web or market data before giving an opinion.
- Starts with a short, plain-language report.
- Saves finished reports when you choose to save them.
- Falls back to a copy/paste prompt if AI setup is not ready.

Gacha does not place trades.
It does not fetch market data by itself.
It prepares the workflow and sends it to your connected AI tool.

## Quick Start

### If You Use an AI Agent

You can ask your coding assistant or terminal agent to install Gacha for you:

```text
Install Gacha by following the README:
https://github.com/dkstm95/gacha
Then run gch and help me complete the first-run setup.
```

### macOS and Linux

Install:

```bash
curl -fsSL https://raw.githubusercontent.com/dkstm95/gacha/main/install.sh | sh
```

Run:

```bash
"$HOME/.local/bin/gch"
```

The installer adds two commands:

- `gch`: short command for daily use
- `gacha`: full command name

You do not need extra programming tools to use Gacha.

On first run, Gacha asks a few questions to set your research profile.
You can skip it, or change it later with `/profile`.

Gacha may also ask to set up AI.
Follow the prompts and connect the account you want to use.

If the installer says `$HOME/.local/bin` is not on `PATH`, run the printed
`export PATH=...` command before using `gch` in that same terminal. Add that
line to your shell profile (`~/.zprofile` for zsh or `~/.bash_profile` for
bash) before opening a new terminal if you want the short command to persist.

### Windows

Download the latest Windows zip:

```text
https://github.com/dkstm95/gacha/releases/latest
```

Unzip it.
Move `gacha.exe` into a folder where you keep command-line tools.
Then open a new terminal window:

```powershell
gacha setup
```

If you want the short command too, copy `gacha.exe` as `gch.exe`:

```powershell
Copy-Item gacha.exe gch.exe
```

Windows automatic AI setup is not supported yet.
Install the extra AI tool shown by Gacha first.
Then run:

```powershell
gacha setup
```

## Your First Question

Start the app:

```bash
gch
```

You will see a simple prompt-first terminal screen.
If this is your first run, Gacha asks for default research preferences first:
markets, time horizon, risk preference, report style, and common goals.

Type a question:

```text
Should I buy NVDA now?
```

Or ask one question without opening the app:

```bash
gch "Should I buy NVDA now?"
```

You do not need to choose a model. Gacha handles model routing through the
local AI runtime and falls back to your runtime default when needed.

## Example Questions

```text
What should I invest in for the next 6 to 12 months?
I want exposure to AI infrastructure. Which stocks or ETFs should I compare?
I want to invest in semiconductors. What should I compare?
I own TSLA. When should I trim, sell, or stop out?
Review my portfolio: AAPL 35%, NVDA 30%, SGOV 35%.
```

Good questions include:

- your goal
- your time horizon
- your risk tolerance
- your current holdings, if they matter

## If Setup Is Not Ready

Check your setup:

```bash
gch doctor
```

This shows:

- whether AI setup is ready
- whether an AI account is connected
- where reports are saved

Fix setup:

```bash
gch setup
```

On macOS and Linux, `gch setup` can finish AI setup for you.
Then it starts account login.

On Windows, install the required AI tool separately first.

If AI setup still cannot run, Gacha prints a complete prompt.
You can paste that prompt into a web AI with browsing.

## App Commands

Inside the app:

```text
/profile
/settings
/theme
/help
/quit
```

`/profile` shows or edits your research profile.
`/settings` opens language and theme settings.
`/theme` jumps directly to the interactive theme selector.
Use the arrow keys and enter, or type a full command directly:

```text
/language auto
/language en
/language ko
/theme system
/theme dark
/theme light
/theme gacha
```

Setup, diagnostics, and updates stay outside the app:

```bash
gch profile
gch setup
gch doctor
gch update
```

## Privacy and Data Flow

Gacha stores your profile and saved reports on your machine.
Reports are saved only after you choose to save them.

When you ask a question, Gacha sends the research prompt to your connected
OpenCode runtime and AI provider.
Your AI provider's own data policy applies to that request.

Gacha does not include product analytics or telemetry in the current version.
`gch update` contacts GitHub only when you run the update command.

## Reports

When the AI finishes a report, Gacha asks whether to save it.

Default location:

```text
~/.local/share/gacha/reports
```

Reports are saved only when you answer yes.
Copy/paste fallback prompts are not saved as reports.

## Language

Gacha tries to match your terminal language.
If your question contains Korean text, Gacha asks the AI to answer in Korean.

To set the language inside the app:

```text
/settings
```

Then choose `Language`, or type `/language en`, `/language ko`, or
`/language auto` directly.

## Updates

macOS and Linux:

```bash
gch update
```

This downloads the right app file for your computer.
Then it replaces the old one.

Download the latest Windows zip from the releases page.
Replace your old `gacha.exe`.
Then open a new terminal.

## Uninstall

macOS and Linux:

```bash
rm -f ~/.local/bin/gacha ~/.local/bin/gch
```

If you installed to a custom `GACHA_INSTALL_DIR`, remove `gacha` and `gch`
from that directory instead.

Windows:

```powershell
Remove-Item .\gacha.exe
Remove-Item .\gch.exe
```

Run those commands in the folder where you placed the executables.

To remove Gacha's local profile and saved reports too:

```bash
rm -rf ~/.config/gacha ~/.local/share/gacha
```

This does not remove OpenCode or your AI provider credentials.
Remove those separately only if you no longer use them for anything else.

## Fresh Data

Investment information changes quickly.
Gacha always tells the AI to check current web or market data.
It does this even if your question does not say "latest".

If current data cannot be checked, the AI should not make a recommendation.

A good Gacha report should make these clear:

- what data was checked, when, and from which sources
- the plain bottom line
- the next action and review timing
- the biggest risks and opposite view
- when to buy, hold, sell, or watch

## Limits

Gacha does not:

- place trades
- promise returns
- replace professional financial, tax, or legal advice
- directly fetch market data in the current version

Gacha prepares a strict research workflow.
Your connected AI tool performs the current web or market-data research.

## Developers

Development notes: [docs/development.md](docs/development.md)

## License

MIT
