package main

import (
	"flag"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Bowery/slack"
	nsq "github.com/bitly/go-nsq"
)

func main() {
	var lookupd, topic, channel, from, token string

	flag.StringVar(&lookupd, "lookupd", "http://127.0.0.1:4161", "lookupd HTTP address")
	flag.StringVar(&topic, "topic", "", "NSQD topic")
	flag.StringVar(&channel, "channel", "", "Slack channel")
	flag.StringVar(&from, "from", "nsq_to_slack", "Slack username")
	flag.StringVar(&token, "token", "", "Slack auth token")
	flag.Parse()

	if lookupd == "" || topic == "" || channel == "" || from == "" || token == "" {

		flag.PrintDefaults()
		log.Fatal("invalid options")
	}

	client := slack.NewClient(token)

	nsqChannel := "nsq_to_slack" + strconv.FormatInt(time.Now().Unix(), 10) +
		"#ephemeral"

	c, err := nsq.NewConsumer(topic, nsqChannel, nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}

	c.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		if err := client.SendMessage(channel, string(m.Body), from); err != nil {
			log.Print(err)
		}

		m.Finish()

		return nil
	}))

	if err := c.ConnectToNSQLookupd(lookupd); err != nil {
		log.Fatal(err)
	}

	if err := client.SendMessage(channel, "nsq_to_slack announcing messages from topic '"+topic+
		"' to this channel.", from); err != nil {
		log.Print(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
