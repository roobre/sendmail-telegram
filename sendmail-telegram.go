package smtg

import (
	"bytes"
	"errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/mail"
	"strings"
	"text/template"
)

type SendmailTg struct {
	config  Config
	session *tgbotapi.BotAPI
}

type Config struct {
	Token       string
	Format      string
	AddrMapping map[string]int64
	CatchAll    int64
}

const DefaultFormat = `{{ if index . "Subject" -}}
*{{.Subject}}*

{{ end -}}
{{ .Body }}
---
` + "```" +
	`
To: {{.To}}
From: {{.From}}` + "```"

func New(config Config) (*SendmailTg, error) {
	stg := &SendmailTg{
		config: config,
	}

	if strings.TrimSpace(config.Format) == "" {
		stg.config.Format = DefaultFormat
	}

	botapi, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	_, err = botapi.GetMe()
	if err != nil {
		return nil, err
	}

	stg.session = botapi
	return stg, nil
}

func (s *SendmailTg) Updates() ([]*tgbotapi.Chat, error) {
	updates, err := s.session.GetUpdates(tgbotapi.UpdateConfig{})
	if err != nil {
		return nil, err
	}

	notified := map[int64]bool{}
	var chats []*tgbotapi.Chat
	for _, u := range updates {
		if u.Message == nil {
			continue
		}

		if u.Message.Chat == nil {
			continue
		}

		if notified[u.Message.Chat.ID] {
			chats = append(chats, u.Message.Chat)
			notified[u.Message.Chat.ID] = true
		}

		if len(chats) > 5 {
			break
		}
	}

	return chats, nil
}

func (s *SendmailTg) Sendmail(to []*mail.Address, email *mail.Message) error {
	body, err := ioutil.ReadAll(email.Body)
	if err != nil {
		return err
	}

	templateContents := map[string]string{}

	for key, val := range email.Header {
		templateContents[key] = strings.Join(val, ", ")
	}
	templateContents["Body"] = string(body)
	// TODO: Maybe I should give the To header a special treatment as well

	templ, err := template.New("emailtemplate").Parse(s.config.Format)
	if err != nil {
		return err
	}

	msgBuf := &bytes.Buffer{}
	err = templ.Execute(msgBuf, templateContents)
	if err != nil {
		return err
	}

	if len(msgBuf.String()) == 0 {
		return errors.New("refusing to send empty message")
	}

	tgmsg := tgbotapi.NewMessage(0, msgBuf.String())
	tgmsg.ParseMode = tgbotapi.ModeMarkdown

	sentTo := map[int64]bool{}
	for _, addr := range to {
		chatId := s.mapAddress(addr)
		if !sentTo[chatId] {
			tgmsg.ChatID = chatId
			_, err = s.session.Send(tgmsg)
			if err != nil {
				log.Println(err)
			}

			sentTo[chatId] = true
		}
	}

	return nil
}

func (s *SendmailTg) mapAddress(address *mail.Address) int64 {
	if chatId, found := s.config.AddrMapping[address.Address]; found {
		return chatId
	}

	return s.config.CatchAll
}
