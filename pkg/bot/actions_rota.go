package bot

import (
	"bytes"
	"fmt"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"strings"
	"text/template"
	"time"

	"github.com/go-joe/joe"
	"github.com/jinzhu/now"
	slackAPI "github.com/slack-go/slack"
	"github.com/spf13/viper"
)

func (b *Bot) Postmortem(message joe.Message) error {
	requestedPostmortemTitle := strings.TrimSpace(strings.TrimPrefix(message.Text, "postmortem"))
	if requestedPostmortemTitle == "" {
		message.Respond("You must provide a title for your postmortem: @%s postmortem Service outage", b.Bot.Name)
		return nil
	}

	var postmortemNameTemplate = template.Must(
		template.New("").
			Parse(`{{.PostmortemDate}}.{{.PostmortemTitle}}.Postmortem`))

	var tpl bytes.Buffer
	err := postmortemNameTemplate.Execute(&tpl, struct {
		PostmortemDate  string
		PostmortemTitle string
	}{
		PostmortemDate:  time.Now().Format("2006-01-02"),
		PostmortemTitle: strings.Replace(requestedPostmortemTitle," ", "-", -1),
	})
	if err != nil {
		b.Logger.Error(fmt.Sprintf("failed to template postmortem title %v", err))
	}

	driveService := drive.NewFilesService(b.Drive)

	tmpl, err := driveService.Get(viper.GetString("google.drive.postmortem_file_id")).Do()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("failed to get postmortem template %v", err))
	}

	md := drive.File{
		Name: tpl.String(),
	}

	newPostmortem, err := driveService.Copy(tmpl.Id, &md).Do()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("failed to create file %v", err))
	}

	docsService := docs.NewDocumentsService(b.Docs)
	docsRequests := make([]*docs.Request, 0)
	docsRequests = append(docsRequests, &docs.Request{
		ReplaceAllText: &docs.ReplaceAllTextRequest{
			ContainsText: &docs.SubstringMatchCriteria{
				MatchCase: false,
				Text:      "{{title}}",
			},
			ReplaceText: requestedPostmortemTitle,
		},
	})
	docsRequests = append(docsRequests, &docs.Request{
		ReplaceAllText: &docs.ReplaceAllTextRequest{
			ContainsText: &docs.SubstringMatchCriteria{
				MatchCase: false,
				Text:      "{{date}}",
			},
			ReplaceText: time.Now().Format(time.RFC1123),
		},
	})
	docsRequests = append(docsRequests, &docs.Request{
		ReplaceAllText: &docs.ReplaceAllTextRequest{
			ContainsText: &docs.SubstringMatchCriteria{
				MatchCase: false,
				Text:      "{{status}}",
			},
			ReplaceText: "In Progress",
		},
	})
	req := docs.BatchUpdateDocumentRequest{
		Requests: docsRequests,
	}
	updatedRequest, err := docsService.BatchUpdate(newPostmortem.Id, &req).Do()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("Failed to update the postmortem text %v", err))
	}

	if updatedRequest.HTTPStatusCode == 200 {
		b.Logger.Info("Successfully created postmortem")
		_, _, err := b.Slack.PostMessage(message.Channel,
			slackAPI.MsgOptionText(
				fmt.Sprintf("I've created a postmortem <https://docs.google.com/document/d/%s/edit#|here>",
					updatedRequest.DocumentId), false),
		)
		if err != nil {
			fmt.Printf("%s\n", err)
			return nil
		}
	}

	return nil
}

func (b *Bot) DailySendLevel1TheRunbook() {
	level1, err := b.getRotaLevel1()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("level 1 rota user retrieval error %v", err))
	}

	channelID, timestamp, err := b.Slack.PostMessage(level1.ID, slackAPI.MsgOptionText("", false)) // TODO Format this with some relevant text
	if err != nil {
		b.Logger.Error(fmt.Sprintf("error sending daily runbook message %v", err))
	}
	b.Logger.Info(fmt.Sprintf("Message successfully sent to channel %s at %s", channelID, timestamp))
}

func (b *Bot) DailySendLevel1TheSignoffReminder() {
	level1, err := b.getRotaLevel1()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("level 1 rota user retrieval error %v", err))
	}

	channelID, timestamp, err := b.Slack.PostMessage(level1.ID, slackAPI.MsgOptionText("", false)) // TODO Format this with some relevant text
	if err != nil {
		b.Logger.Error(fmt.Sprintf("error sending daily signoff reminder %v", err))
	}
	b.Logger.Info(fmt.Sprintf("Message successfully sent to channel %s at %s", channelID, timestamp))
}

func (b *Bot) DailySendLevel1TheCongratulationsMessage() {
	level1, err := b.getRotaLevel1()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("level 1 rota user retrieval error %v", err))
	}

	channelID, timestamp, err := b.Slack.PostMessage(level1.ID, slackAPI.MsgOptionText("", false)) // TODO Format this with some relevant text
	if err != nil {
		b.Logger.Error(fmt.Sprintf("error sending daily congratulations message %v", err))
	}
	b.Logger.Info(fmt.Sprintf("Message successfully sent to channel %s at %s", channelID, timestamp))
}

func (b *Bot) GetTodaysRota(message joe.Message) error {
	level1, err := b.getRotaLevel1()
	if err != nil {
		return fmt.Errorf("level 1 rota user retrieval error %v", err)
	}
	level2, err := b.getRotaLevel2()
	if err != nil {
		return fmt.Errorf("level 2 rota user retrieval error %v", err)
	}

	message.Respond("%s is on Level 1 today and %s is on Level 2 this week!", level1.Name, level2.Name)

	return nil
}

// Set the questions channel with level 1 and level 2 on-call user details
func (b *Bot) DailySetQuestionsChannelTopic() {
	level1, err := b.getRotaLevel1()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("level 1 rota user retrieval error %v", err))
	}
	level2, err := b.getRotaLevel2()
	if err != nil {
		b.Logger.Error(fmt.Sprintf("level 2 rota user retrieval error %v", err))
	}

	topicText := fmt.Sprintf(
		`L1: %s
		L2: %s`,
		level1.Name, level2.Name)

	topic, err := b.Slack.SetChannelTopic("C0159JKU1NW", topicText) // TODO Pull channel ID out to a conf
	if err != nil {
		b.Logger.Error(fmt.Sprintf("topic setting error %v", err))
	}
	b.Logger.Info(fmt.Sprintf("topic set to: %s", topic))
}

func (b *Bot) getRotaLevel2() (*slackAPI.User, error) {
	weeksRota, err := b.Calendar.Events.
		List(viper.GetString("calendar.rota_calendar_id")).
		TimeMin(now.BeginningOfWeek().Format(time.RFC3339)).
		TimeMax(now.EndOfWeek().Format(time.RFC3339)).
		Do()

	if err != nil {
		return nil, err
	}

	var level2 string
	for _, e := range weeksRota.Items {
		if strings.HasPrefix(e.Summary, "L2:") {
			level2 = e.Attendees[0].Email
		}
	}

	level2user, err := b.Slack.GetUserByEmail(level2)
	if err != nil {
		return nil, err
	}

	return level2user, nil
}

func (b *Bot) getRotaLevel1() (*slackAPI.User, error) {
	weeksRota, err := b.Calendar.Events.
		List(viper.GetString("calendar.rota_calendar_id")).
		TimeMin(now.BeginningOfDay().Format(time.RFC3339)).
		TimeMax(now.EndOfDay().Format(time.RFC3339)).
		Do()

	if err != nil {
		return nil, err
	}

	var level1 string
	for _, e := range weeksRota.Items {
		if strings.HasPrefix(e.Summary, "L1:") {
			level1 = e.Attendees[0].Email
		}
	}

	level1user, err := b.Slack.GetUserByEmail(level1)
	if err != nil {
		return nil, err
	}

	return level1user, nil
}
