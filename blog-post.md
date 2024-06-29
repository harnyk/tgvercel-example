# Как по-быстрому захостить телеграм-бот на Vercel

...и автоматизировать этот процесс по-максимуму!

## Проблема

Любой туториал по программированию телеграм-ботов чаще всего рассказывает нам о том, как запускать бот в режиме поллинга на машине разработчика. Этот режим хорош, когда надо по-быстрому отладить прототип бота, и машина, на которой он запущен, не имеет внешнего IP-адреса, принимающего соединения из мира. В поллинг-режиме АПИ клиент телеграм-бота просто подключается к серверу телеграм и периодически опрашивает его на предмет наличия обновлений (новых сообщений, событий итд). Для продакшена такой режим не ревомендуется, так как он создаёт повышенную нагрузку на сервер телеграма. Кроме того, поллинг-режим просто не может работать в бессерверной среде, потому что для этого нужно держать постоянно запущенный процесс бота, что противоречит самой идее бессерверного хостинга.

Для таких вещей существует вебхук-режим телеграм-бота. Сам теелграм-бот становится HTTPS-эндпоинтом, у него должен быть специальный URL, доступный извне, и этот URL должен быть известен телеграму. Тогда телеграм начинает сам отправлять обновления на вебхук-URL такого бота.

Для этого всего нужно сделать кучу теледвижений, которые хотелось бы автоматизировать. Кроме того, хотелось бы упростить создание вебхук-эндпоинта бота под бесплатный облачный хостинг Vercel, который идеален для таких приложений. Мы его сделаем ещё идеальнее.

## Пишем эндпоинт

Кроме статического хостинга, Vercel позволяет писать бессерверные функции. Это чем-то похоже на AWS Lambda (а под капотом там и есть AWS Lambda), только сильно проще.

Всё, что надо сделать, это создать в папке `api` некий файл на любом поддерживаемом языке. Да, забыл сказать, это туториал для гоферов, поэтому мы выбираем Go. Итак, Hello World под Vercel, написанный на Go будет выглядеть примерно так:

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

Деплоится это одной командой:

```sh
vercel
```

И тестируется другой командой:

```sh
 curl https://my-project-name.vercel.app/api/hello
 #  Hello World!
```

Здесь, `my-project-name` - это имя вашего проекта

Поигрались, работает.

Теперь идём за знаменитой библиотекой [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api).

Единственный пример в их документации, посвящённый вебхукам, рассказывает, как средствами библиотеки поднять HTTP-сервер, слушающий на определённом порту.

```go
    // пропущено много кода...
	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)
    for update := range updates {
		log.Printf("%+v\n", update)
	}
```

Это всё хорошо, но облачный бессерверный хостинг работает не так. Нельзя просто так взять и запустить там долгоживущий процесс, слушающий сокет. Это всё хорошо на VDS-ке, но не на Vercel, который сам рулит своими сокетами, сам поднимает, масштабирует и убивает наши процессы и т.д.

Если спуститься на уровень ниже в библиотеке `telegram-bot-api`, то видим, что у структуры `Bot` есть метод `HandleUpdate`, которому можно передать `*http.Request`, а он сам разберётся. Это уже лучше сочетается с бессерверной природой Vercel.

Таким образом наш бот будет выглядеть примерно так:

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

Это простейший Телеграм-бот, отвечающий на любое сообщение его копией с припиской `you say: `.

Что мы тут видим? Появилась переменная окружения `TELEGRAM_TOKEN`. Её надо будет создать в админке Vercel для корректной работы.

Чего тут не хватает? Не хватает тут безопасности. Любой дурак может, зная URL нашего эндпоинта, отправить на него что угодно, имитируя, что это Телеграм отправил вебхук. Чуть позже мы рещим эту проблему.

Но пока - чего ещё не хватает? В этой схеме нет самого главного - кода, который бы конфигурировал бот на стороне Телеграм-сервера, сообщая ему URL нашего вебхука. В конфигурации с долгожижвущим процессом, запускаемом на VDS, это тривиальная задача. Но в бессерверной среде наш процесс не запустится, пока ему не будет отправлен HTTP-запрос, а HTTP-запрос не будет отправлен, пока Телеграм не будет знать URL вебхука. Замкнутый круг.

Мы могли бы взять `curl`, и просто из терминала отправить запрос на Telegram Bot API, типа такого:

```sh
curl -X POST https://api.telegram.org/bot${TOKEN}/setWebhook -d "url=${WEBHOOK_URL}"
```

Но тут надо знать `WEBHOOK_URL`, и если для продакшен-стреды на Vercel он однозначно известен, то для каждого preview-деплоймента он может генерироваться динамичеки. Кроме того, надо хранить токены для Телеграм-ботов для каждого окружения (production, preview, development.). А лучшим местом для этого служит Vercel. И кроме того, в идеале, всё, что мы знаем про деплоймент, это его ID или URL, то наш гипотетический скрипт может не знать в какую среду оно деплоится, и, соответственно, какой из токенов выбрать.

Короче, я подвожу читателя к тому, что я хочу попиарить две своих поделки, сильно упрощабщие цикл разработки и поддержки Телеграм-бота, крутящегося на Vercel.

Итак, встречаем, [tgvercel](https://github.com/harnyk/tgvercel). Это простая утилита позволяет:

-   сконфигурировать на Vercel токен для Телеграм-бота и секрет для защиты вебхука;
-   создать вебхук для Телеграм-бота;

Сначала сконфигурируем переменные окружения для preview и production.

```sh
tgvercel init --target=preview --telegram-token=YOUR_PREVIEW_TELEGRAM_TOKEN
tgvercel init --target=production --telegram-token=YOUR_PRODUCTION_TELEGRAM_TOKEN
```

В каждом окружении появилась пара переменных окружения:

-   `TELEGRAM_TOKEN` - токен для Телеграм-бота
-   `TELEGRAM_WEBHOOK_SECRET` - секрет для защиты вебхука, о нём поговорим позже

Затем, когда мы деплоим проект на Vercel, нужно сохранить в переменную stdout команды `vercel` - это и будет URL деплоймента.

```sh
DEPLOYMENT_URL=$(vercel)
```

И наконец, нам нужно создать вебхук для Телеграм-бота. Вот как это сделать:

```sh
tgvercel hook ${DEPLOYMENT_URL} /api/webhook
```

Теперь поговорим о безопасности нашего webhook-эндпоинта.

URL вебхука, который мы отправляем в Телеграм, содержит query-параметр `secret`. Его сгенерировала команда `tgvercel init` и записала в переменную окружения `TELEGRAM_WEBHOOK_SECRET`.

То есть наш обработчик веб-хука должен проверять какой secret пришел в запросе. Если он неверный (то есть не равен тому, что лежит в переменной окружения `TELEGRAM_WEBHOOK_SECRET`), то возвращать ошибку `401 Unauthorized`.

Чтобы не писать это всё вручную, предлагаю использовать библиотеку [tgvercelbot](https://github.com/harnyk/tgvercelbot). Это обёртка над `telegram-bot-api`, которая позволяет без лишних усилий интегрировать бот с Vercel. Вместо всего нашего кода, мы можем написать просто:

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

Что делает метод `tgv.HandleWebhook`? Он читает переменные окружения `TELEGRAM_WEBHOOK_SECRET` и `TELEGRAM_TOKEN` и создает `tgbotapi.BotAPI`. Также он проверяет, что query-параметр запроса `secret` равен тому, что лежит в переменной окружения `TELEGRAM_WEBHOOK_SECRET`. И, если всё нормально, то вызывает пользовательскую коллбек-функцию `func (bot *tgbotapi.BotAPI, update *tgbotapi.Update)`, где мы уже можем обрабатывать входящие сообщения и вообще делать всё, что захотим с инстансом API-клиента (`bot`) и обновлением (`update`).

То есть, библиотека `tgvercelbot` и утилита `tgvercel` имеют общие соглашения об именах переменных окружения и отлично упрощают жизнь разработчикам ботов.

## Локальный режим

"Но постойте-ка, а как же запустить такой бот локально?" спросите вы.

Для этого в `tgvercelbot` есть функция RunLocal, которой передаётся Телеграм-токен и всё та же функция-коллбек `func (bot *tgbotapi.BotAPI, update *tgbotapi.Update)`, реализующая пользовательскуб логику бота.

Давайте для начала вынесем функцию в отдельный пакет:

```go
// File: pkg/botlogic/onupdate.go
package botlogic

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

И теперь используем её в нашем обработчике:

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

Тогда, если вам нужно будет запустить бота локально, можно написать следующий код:

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

    // На самом деле не важно,
    // откуда вы будете брать токен для локальной разработки
	token := os.Getenv("TELEGRAM_TOKEN")

	err = tgvercelbot.RunLocal(token, botlogic.OnUpdate)
	if err != nil {
		log.Fatalf("failed to run locally: %v", err)
	}
}
```

Запускается эта штука обычной командой (не `vercel dev`, обратите внимание!):

```sh
go run main.go
```

## Заключение

Теперь вы знаете, как без особого головняка захостить Телеграм-бота, написанного на Go, на Vercel.

## Ссылки

1.  [tgvercelbot](https://github.com/harnyk/tgvercelbot)
2.  [tgvercel](https://github.com/harnyk/tgvercel)
3.  [репозиторий с примерами из этой статьи](https://github.com/harnyk/tgvercel-example)
