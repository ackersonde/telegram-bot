![Deploy telegram bot](https://github.com/ackersonde/telegram-bot/workflows/Deploy%20telegram%20bot/badge.svg)

# telegram-bot
This is my personal Telegram bot. I primarily use it to [upload PDFs to my reMarkable tablet](https://github.com/ackersonde/telegram-bot/blob/main/telegram.go#L86).

<a href="https://core.telegram.org/bots"><img src="https://core.telegram.org/file/811140763/1/PihKNbjT8UE/03b57814e13713da37"></a>

After [SalesForce acquired Slack](https://www.fool.com/investing/2021/01/28/heres-why-this-277-billion-acquisition-by-salesfor/), I decided to get my house in order to ensure I'm not left [bot-less](https://github.com/ackersonde/bender-slackbot/blob/master/README.md) :(

# Installation and Development
It's written in Golang and uses the [Golang bindings for the Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api/blob/master/README.md).

# Building & Running
Every push will redeploy the bot to my DigitalOcean droplet. See the [github action workflow](.github/workflows/build.yml)
