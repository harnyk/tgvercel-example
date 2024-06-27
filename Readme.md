# Example of using `tgvercel` and `tgvercelbot` tools

## How to use

Clone the repository.

Create a Vercel project.

Run:

```sh
go run github.com/harnyk/tgvercel init --target=preview --telegram-token=YOUR_TELEGRAM_TOKEN
```

Deploy your project using the following command:

```sh
make deploy-preview
```

Send something to your bot in Telegram and it will respond with the same message.

## Links

-   [tgvercel](https://github.com/harnyk/tgvercel) - a command line tool to setup Telegram webhooks for Vercel
-   [tgvercelbot](https://github.com/harnyk/tgvercelbot) - a tiny wrapper around the [tgbotapi](https://github.com/go-telegram-bot-api/telegram-bot-api) library to make it easier to listen to Telegram webhooks
