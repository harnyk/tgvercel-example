# How to Quickly Host a Telegram Bot on Vercel

...and automate this process as much as possible!

## Problem

Any tutorial on programming Telegram bots usually tells us how to run a bot in polling mode on the developer's machine. This mode is good when you need to quickly debug a bot prototype and the machine it is running on does not have an external IP address that accepts connections from the outside world. In polling mode, the Telegram bot API client simply connects to the Telegram server and periodically polls it for updates (new messages, events, etc.). This mode is not recommended for production as it creates an increased load on the Telegram server. Additionally, polling mode simply cannot work in a serverless environment because it requires a constantly running bot process, which contradicts the very idea of serverless hosting.

For such cases, there is the webhook mode of a Telegram bot. The Telegram bot itself becomes an HTTPS endpoint; it must have a special URL accessible from the outside, and this URL must be known to Telegram. Then, Telegram starts sending updates to the bot's webhook URL itself.

For all this, a lot of steps are required that we would like to automate. Furthermore, we would like to simplify the creation of the bot's webhook endpoint for the free cloud hosting Vercel, which is ideal for such applications. We will make it even more ideal.

## Writing the Endpoint

In addition to static hosting, Vercel allows you to write serverless functions. This is somewhat similar to AWS Lambda (and under the hood, it is AWS Lambda), but much simpler.

All you need to do is create a file in the `api` folder in any supported language. Yes, I forgot to mention, this tutorial is for Gophers, so we choose Go. So, a Hello World on Vercel written in Go will look something like this:

```go
// File: api/hello.go
package handler

import (
    "fmt"
    "net/http"
)

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello World!")
}
```

Deploy it with a single command:

```sh
vercel
```

And test it with another command:

```sh
curl https://my-project-name.vercel.app/api/hello
# Hello World!
```

Here, `my-project-name` is the name of your project.

Played around, it works.

Now let's get the famous [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) library.

The only example in their documentation dedicated to webhooks explains how to use the library to raise an HTTP server listening on a specific port.

```go
// a lot of code omitted...
updates := bot.ListenForWebhook("/" + bot.Token)
go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)
for update := range updates {
	log.Printf("%+v\n", update)
}
```

This is all good, but cloud serverless hosting doesn't work that way. You can't just take and run a long-running process listening on a socket there. This is all good on a VDS, but not on Vercel, which manages its sockets itself, raises, scales, and kills our processes, etc.

If we go one level lower in the `telegram-bot-api` library, we see that the `Bot` structure has a `HandleUpdate` method that can take an `*http.Request`, and it will handle it itself. This is more compatible with Vercel's serverless nature.

So, our bot will look something like this:

```go
// File: api/webhook.go
package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	update, err := bot.HandleUpdate(r)
	if err != nil {
		log.Fatal(err)
	}

	if update.Message != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "you say: "+update.Message.Text)
		_, err := bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
```

This is a simple Telegram bot that responds to any message with a copy of it prefixed with `you say: `.

What do we see here? There is an environment variable `TELEGRAM_TOKEN`. It will need to be created in the Vercel dashboard for it to work correctly.

What's missing here? Security. Anyone who knows our endpoint URL can send anything to it, pretending that Telegram sent the webhook. We will solve this problem a bit later.

But for now, what else is missing? The most important thing is missing - the code that would configure the bot on the Telegram server side, informing it of our webhook URL. In the configuration with a long-running process launched on a VDS, this is a trivial task. But in a serverless environment, our process will not start until an HTTP request is sent to it, and an HTTP request will not be sent until Telegram knows the webhook URL. A vicious circle.

We could use `curl` and simply send a request to the Telegram Bot API from the terminal, like this:

```sh
curl -X POST https://api.telegram.org/bot${TOKEN}/setWebhook -d "url=${WEBHOOK_URL}"
```

But you need to know the `WEBHOOK_URL`, and while it is definitely known for the production environment on Vercel, it can be dynamically generated for each preview deployment. Also, you need to store Telegram bot tokens for each environment (production, preview, development). And the best place for this is Vercel. Furthermore, ideally, all we know about the deployment is its ID or URL, so our hypothetical script might not know which environment it is deploying to and, accordingly, which token to choose.

In short, I am leading the reader to the fact that I want to promote my two projects that greatly simplify the development and maintenance cycle of a Telegram bot running on Vercel.

So, meet [tgvercel](https://github.com/harnyk/tgvercel). This simple utility allows you to:

-   Configure the token for the Telegram bot and the secret for protecting the webhook on Vercel;
-   Create a webhook for the Telegram bot;

First, configure the environment variables for preview and production.

```sh
tgvercel init --target=preview --telegram-token=YOUR_PREVIEW_TELEGRAM_TOKEN
tgvercel init --target=production --telegram-token=YOUR_PRODUCTION_TELEGRAM_TOKEN
```

Each environment will have a pair of environment variables:

-   `TELEGRAM_TOKEN` - the token for the Telegram bot
-   `TELEGRAM_WEBHOOK_SECRET` - the secret for protecting the webhook, which we will talk about later

Then, when we deploy the project to Vercel, we need to save the `vercel` command's stdout to a variable - this will be the deployment URL.

```sh
DEPLOYMENT_URL=$(vercel)
```

And finally, we need to create a webhook for the Telegram bot. Here's how to do it:

```sh
tgvercel hook ${DEPLOYMENT_URL} /api/webhook
```

Now let's talk about the security of our webhook endpoint.

The webhook URL that we send to Telegram contains the `secret` query parameter. This was generated by the `tgvercel init` command and recorded in the `TELEGRAM_WEBHOOK_SECRET` environment variable.

So our webhook handler should check which secret came in the request. If it is incorrect (i.e., not equal to what is in the `TELEGRAM_WEBHOOK_SECRET` environment variable), then it should return a `401 Unauthorized` error.

To avoid writing all this manually, I suggest using the [tgvercelbot](https://github.com/harnyk/tgvercelbot) library. This is a wrapper around `telegram-bot-api` that allows you to easily integrate the bot with Vercel. Instead of all our code, we can simply write:

```go
import (
    "github.com/harnyk/tgvercelbot"
    "http"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var tgv = tgvercelbot.New(tgvercelbot.DefaultOptions())

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	tgv.HandleWebhook(r, func (bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "you say: "+update.Message.Text)
		_, err := bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
    })
}
```

What does the `tgv.HandleWebhook` method do? It reads the `TELEGRAM_WEBHOOK_SECRET` and `TELEGRAM_TOKEN` environment variables and creates `tgbotapi.BotAPI`. It also checks that the request's `secret` query parameter matches what is in the `TELEGRAM_WEBHOOK_SECRET` environment variable. And, if everything is fine, it calls the user callback function `func (bot *tgbotapi.BotAPI, update *tgbotapi.Update)`, where we can process incoming messages and generally do whatever we want with the API client instance (`bot`) and the update (`update`).

Thus, the `tgvercelbot` library and the `tgvercel` utility have common conventions for environment variable names and greatly simplify the lives of bot developers.

## Local Mode

"But wait, how do you run such a bot locally?" you ask.

For this, `tgvercelbot` has a `RunLocal` function, which takes the Telegram token and the same callback function `func (bot *tgbotapi.BotAPI, update *tgbotapi.Update)`, implementing the bot's custom logic.

Let's start by moving the function to a separate package:

```go
// File: pkg/botlogic/onupdate.go
package botlogic

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func OnUpdate(
	bot *tgbotapi.BotAPI,
	update *tgbotapi.Update) {
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "you say: "+update.Message.Text)
		_, err := bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
```

And now use it in our handler:

```go
// File: api/webhook.go
package handler

import (
	"net/http"

	"my-module-name/pkg/botlogic"
	"github.com/harnyk/tgvercelbot"
)

var tgv = tgvercelbot.New(tgvercelbot.DefaultOptions())

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	tgv.HandleWebhook(r, botlogic.OnUpdate)
}
```

Then, if you need to run the bot locally, you can write the following code:

```go
// File: main.go
package main

import (
	"log"
	"os"

	"my-module-name/pkg/botlogic"
	"github.com/harnyk/tgvercelbot"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

    // It doesn't really matter
    // where you get the token for local development
	token := os.Getenv("TELEGRAM_TOKEN")

	err = tgvercelbot.RunLocal(token, botlogic.OnUpdate)
	if err != nil {
		log.Fatalf("failed to run locally: %v", err)
	}
}
```

Run this with the usual command (not `vercel dev`, note!):

```sh
go run main.go
```

## Conclusion

Now you know how to easily host a Telegram bot written in Go on Vercel.

## Links

1. [tgvercelbot](https://github.com/harnyk/tgvercelbot)
2. [tgvercel](https://github.com/harnyk/tgvercel)
3. [repository with examples from this article](https://github.com/harnyk/tgvercel-example)
