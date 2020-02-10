package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/mail"
	"os"
	smtg "roob.re/sendmail-telegram"
)

func main() {
	app := cobra.Command{
		Use:   "sendmail",
		Short: "Send an email to telegram users through a bot",
		Run:   sendmail,
		Args:  cobra.ArbitraryArgs,
	}
	app.Flags().Bool("t", true, "Extract recipients from message headers. These are added to any recipients specified on the command line.")

	app.AddCommand(&cobra.Command{
		Use:   "aid",
		Short: "Print recent updates for the bot",
		Run:   aid,
	})

	err := app.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func newSmgtg() (*smtg.SendmailTg, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	v.SetConfigFile("sendmail-telegram.yml")

	v.AddConfigPath(".")

	if os.Getenv("XDG_CONFIG_HOME") != "" {
		v.AddConfigPath("$XDG_CONFIG_HOME")
	}

	if os.Getenv("HOME") != "" {
		v.AddConfigPath("$HOME/.config")
	}

	v.AddConfigPath("/etc")

	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	config := smtg.Config{}
	err = v.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %s", err)
	}

	return smtg.New(config)
}

func aid(cmd *cobra.Command, args []string) {
	mailer, err := newSmgtg()
	if err != nil {
		log.Fatal(err)
	}

	updates, err := mailer.Updates()
	if err != nil {
		log.Fatal(err)
	}

	if len(updates) == 0 {
		fmt.Println("No updates found. Send a message to your bot first.")
	}

	for _, u := range updates {
		fmt.Printf("%s (%s): %d\n", u.FirstName, u.Title, u.ID)
	}
}

func sendmail(cmd *cobra.Command, args []string) {
	mailer, err := newSmgtg()
	if err != nil {
		log.Fatal(err)
	}

	msg, err := mail.ReadMessage(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var to []*mail.Address
	for _, argAddr := range args {
		addr, err := mail.ParseAddress(argAddr)
		if err != nil {
			log.Printf("error parsing address '%s', ignronig: %s", argAddr, err.Error())
			continue
		}
		to = append(to, addr)
	}

	if parseTo, _ := cmd.Flags().GetBool("t"); parseTo {
		for _, header := range []string{"To", "Cc", "Bcc"} {
			bodyRecipients, err := msg.Header.AddressList(header)
			if err != nil && !errors.Is(err, mail.ErrHeaderNotPresent) {
				log.Printf("could not parse address: " + err.Error())
				return
			}

			to = append(to, bodyRecipients...)
		}
	}

	if len(to) == 0 {
		log.Printf("recipient list is empty")
		return
	}

	err = mailer.Sendmail(to, msg)
	if err != nil {
		log.Print(err)
		return
	}
}
