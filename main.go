package main

import (
	"encoding/json"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/mackerel-client-go"
	"github.com/slack-go/slack"
	"log"
	"os"
)

var (
	mackerelClient *mackerel.Client
	slackClient    *slack.Client
	version        string
)

var args struct {
	MackerelAPIKey string `long:"mackerel-api-key" required:"true" env:"MFOOL_MACKEREL_API_KEY" description:"Mackerel API Key"`
	SlackToken     string `long:"slack-token" required:"true" env:"MFOOL_SLACK_TOKEN" description:"Slack Bot Token"`
	SlackChannelID string `long:"slack-channel-id" required:"true" env:"MFOOL_SLACK_CHANNEL_ID" description:"Slack Channel ID"`
	Version        bool   `short:"v" long:"version" description:"Show version"`
}

func main() {
	_, _ = flags.Parse(&args)

	if args.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	mackerelClient = mackerel.NewClient(args.MackerelAPIKey)
	org, err := mackerelClient.GetOrg()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Organization:", org.Name)

	slackClient = slack.New(args.SlackToken)
	_, err = slackClient.AuthTest()
	if err != nil {
		log.Fatalln(err)
	}

	notWorkingHosts, err := getHosts("not_working")
	if err != nil {
		log.Fatalln(err)
	}
	if len(notWorkingHosts) != 0 {
		attachment := slack.Attachment{
			Title:     "There hosts are not working status!",
			TitleLink: fmt.Sprintf("https://mackerel.io/orgs/%s/hosts?status=standby&status=maintenance", org.Name),
		}

		for _, notWorkingHost := range notWorkingHosts {
			fields := []slack.AttachmentField{
				{
					Title: "Name",
					Value: notWorkingHost.Name,
					Short: false,
				},
				{
					Title: "ID",
					Value: notWorkingHost.ID,
					Short: true,
				},
				{
					Title: "Status",
					Value: notWorkingHost.Status,
					Short: true,
				},
			}
			attachment.Fields = append(attachment.Fields, fields...)
		}
		err = sendSlack(attachment)
		if err != nil {
			log.Fatalln(err)
		}
	}

	mutedMonitors, err := getMonitors("muted")
	if err != nil {
		log.Fatalln(err)
	}
	if len(mutedMonitors) != 0 {
		attachment := slack.Attachment{
			Title:     "There monitors are muted!",
			TitleLink: fmt.Sprintf("https://mackerel.io/orgs/%s/monitors", org.Name),
		}

		for _, mutedMonitor := range mutedMonitors {
			fields := []slack.AttachmentField{
				{
					Title: "Name",
					Value: mutedMonitor.Name,
					Short: false,
				},
				{
					Title: "ID",
					Value: mutedMonitor.ID,
					Short: true,
				},
				{
					Title: "Is Muted?",
					Value: "yes",
					Short: true,
				},
			}
			attachment.Fields = append(attachment.Fields, fields...)
		}
		err = sendSlack(attachment)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Todo check the notification is disabled https://f.easyuploader.app/eu-prd/upload/20200922211424_3270464255613853487a.png
}

func getHosts(status string) ([]*mackerel.Host, error) {
	param := &mackerel.FindHostsParam{
		Statuses: nil,
	}

	switch status {
	case "not_working":
		param.Statuses = []string{"standby", "maintenance"}

	}

	hosts, err := mackerelClient.FindHosts(param)
	if err != nil {
		return nil, err
	}

	return hosts, nil
}

func getMonitors(status string) ([]*mackerel.MonitorConnectivity, error) {
	monitors, err := mackerelClient.FindMonitors()
	if err != nil {
		return nil, err
	}

	var isMute bool
	if status == "muted" {
		isMute = true
	}

	var filteredMonitors []*mackerel.MonitorConnectivity
	for _, monitor := range monitors {
		var monitorConnectivity mackerel.MonitorConnectivity

		b, err := json.Marshal(monitor)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &monitorConnectivity)
		if err != nil {
			return nil, err
		}

		if monitorConnectivity.IsMute == isMute {
			filteredMonitors = append(filteredMonitors, &monitorConnectivity)
		}
	}

	return filteredMonitors, nil
}

func sendSlack(attachment slack.Attachment) error {
	attachment.Color = "warning"

	_, _, err := slackClient.PostMessage(
		args.SlackChannelID,
		slack.MsgOptionAttachments(attachment),
	)
	return err
}
