package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/fsm"
)

type userData struct {
	Name string
	Age  int
}

type Application struct {
	b *bot.Bot
	f *fsm.FSM

	users map[int64]*userData
}

const (
	stateDefault fsm.StateID = "default"
	stateStart   fsm.StateID = "start"
	stateAskName fsm.StateID = "ask_name"
	stateAskAge  fsm.StateID = "ask_age"
	stateFinish  fsm.StateID = "finish"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app := &Application{
		users: make(map[int64]*userData),
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(app.handlerDefault),
		bot.WithMessageTextHandler("/form", bot.MatchTypeExact, app.handlerForm),
		bot.WithMessageTextHandler("/cancel", bot.MatchTypeExact, app.handlerCancel),
	}

	app.f = fsm.New(
		stateDefault,
		map[fsm.StateID]fsm.Callback{
			stateStart:   app.callbackStart,
			stateAskName: app.callbackAskName,
			stateAskAge:  app.callbackAskAge,
			stateFinish:  app.callbackFinish,
		},
	)

	var err error

	app.b, err = bot.New(os.Getenv("EXAMPLE_TELEGRAM_BOT_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}

	app.b.Start(ctx)
}

func (app *Application) handlerCancel(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	currentState, _ := app.f.Current(userID)
	if currentState == stateDefault {
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Canceled",
	})

	app.f.Transition(userID, stateDefault)
}

func (app *Application) handlerForm(ctx context.Context, b *bot.Bot, update *models.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	currentState, _ := app.f.Current(userID)
	if currentState != stateDefault {
		return
	}

	app.f.Transition(userID, stateStart, chatID, userID)
}

func (app *Application) handlerDefault(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	currentState, _ := app.f.Current(userID)

	switch currentState {
	case stateDefault:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Type /form to start the form",
		})
		return

	case stateAskName:
		if len(update.Message.Text) < 2 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Please enter a valid name, at least 2 characters",
			})
			return
		}

		app.users[userID].Name = update.Message.Text

		app.f.Transition(userID, stateAskAge, chatID)

	case stateAskAge:
		age, errAge := strconv.Atoi(update.Message.Text)
		if errAge != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Please enter a valid age",
			})
			return
		}

		if age < 18 || age > 100 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Please enter an age between 18 and 100",
			})
			return
		}

		app.users[userID].Age = age

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Thank you!",
		})

		app.f.Transition(userID, stateFinish, chatID, userID)

	default:
		fmt.Printf("unexpected state %s\n", currentState)
	}
}

func (app *Application) callbackStart(f *fsm.FSM, args ...any) {
	chatID := args[0]
	userID := args[1].(int64)

	app.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Let's start the form! Type /cancel to cancel",
	})

	app.users[userID] = &userData{}

	f.Transition(userID, stateAskName, chatID)
}

func (app *Application) callbackAskName(f *fsm.FSM, args ...any) {
	chatID := args[0]

	app.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "What's your name? (at least 2 characters)",
	})
}

func (app *Application) callbackAskAge(_ *fsm.FSM, args ...any) {
	chatID := args[0]

	app.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "How old are you? (between 18 and 100)",
	})
}

func (app *Application) callbackFinish(f *fsm.FSM, args ...any) {
	chatID := args[0]
	userID := args[1].(int64)

	app.b.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: chatID,
		Text: fmt.Sprintf("Name: %s\nAge: %d",
			bot.EscapeMarkdown(app.users[userID].Name), app.users[userID].Age),
	})

	f.Transition(userID, stateDefault)
}
