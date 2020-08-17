package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
)

type Bot struct {
	Token  string
	ChatID string
}

// SendMessage sends message to bot
func (b Bot) SendMessage(msg string) (string, error) {
	sendText := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&parse_mode=Markdown&text=%s", b.Token, b.ChatID, msg)

	rsp, err := http.Get(sendText)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func main() {
	endpoint := flag.String("endpoint", "https://node.skycoin.net/api/v1/health", "health endpoint that need to check")
	botToken := flag.String("token", "", "bot token")
	botChatID := flag.String("chatid", "", "bot chat id")
	flag.Parse()

	if *botToken == "" {
		log.Println("token is not provided")
		os.Exit(1)
	}

	if *botChatID == "" {
		log.Println("chat id is not provided")
		os.Exit(1)
	}

	bot := Bot{
		Token:  *botToken,
		ChatID: *botChatID,
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		os.Exit(1)
	}()

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 3 * time.Minute
	b.MaxInterval = 30 * time.Minute
	b.MaxElapsedTime = 24 * time.Hour

	log.Printf("checking endpoint: %s\n", *endpoint)
	checkHealth(&bot, *endpoint)

	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			// check the api health
			backoff.RetryNotify(func() error {
				return checkHealth(&bot, *endpoint)
			}, b, func(_ error, d time.Duration) {
				log.Println("do next check after: ", d)
			})
		}
	}
}

func checkHealth(bot *Bot, endpoint string) error {
	rsp, err := http.Get(endpoint)
	if err != nil {
		log.Println("checking node health failed:", err)
		_, berr := bot.SendMessage(err.Error())
		if berr != nil {
			log.Println("[bot error]: ", berr.Error())
			return berr
		}
		log.Println("bot notified")
		return err
	}
	defer rsp.Body.Close()
	v, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Println("Read endpoint response failed: ", err.Error())
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		_, berr := bot.SendMessage(fmt.Sprintf("endpoint %s returns code: %d", endpoint, rsp.StatusCode))
		if berr != nil {
			log.Println("[bot error]:", berr.Error())
			return berr
		}
		log.Println("bot notified")

		return errors.New(string(v))
	}
	log.Println("health💗")
	return nil
}
