package bot

import (
	"fmt"
	httpserver "github.com/dombo/hiberBot/pkg/bot/custom-http-server"

	"github.com/dombo/hiberBot/pkg/bot/services/google"
	"github.com/go-joe/cron"
	"github.com/go-joe/joe"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"

	slackAPI "github.com/slack-go/slack"
)

type Bot struct {
	*joe.Bot        // Anonymously embed the joe.Bot type so we can use its functions easily.
	conf     Config // You can keep other fields here as well.
	Slack    *slackAPI.Client
	Calendar *calendar.Service
	Docs     *docs.Service
	Drive    *drive.Service
	Actions	 string
}

type StartOfDayEvent struct{}
type BeforeEndOfDayEvent struct{}
type EndOfDayEvent struct{}

func NewBot(conf *Config) (*Bot, error) {
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration %w", err)
	}

	modules := append(conf.Modules(), // TODO Shift these for local time
		cron.ScheduleEvent("30 6 * * 1-5", StartOfDayEvent{}),
		cron.ScheduleEvent("30 14 * * 1-5", BeforeEndOfDayEvent{}),
		cron.ScheduleEvent("30 15 * * 1-5", EndOfDayEvent{}),
	)

	b := &Bot{
		Bot: joe.New(conf.Slack.BotName,
			modules...),
		Slack: slackAPI.New(conf.Slack.Token, slackAPI.OptionDebug(conf.Slack.Debug)),
		Calendar: google.NewCalendarService(
			conf.Google.Calendar.User,
			[]string{
				calendar.CalendarReadonlyScope,
			}),
		Docs: google.NewDocsService(
			conf.Google.Docs.User,
			[]string{
				docs.DocumentsScope, // TODO Can we reduce these permissions somehow
			}),
		Drive: google.NewDriveService(
			conf.Google.Drive.User,
			[]string{
				drive.DriveScope, // TODO Can we reduce these permissions somehow
			}),
	}

	// Events API authentication handled in custom server.go implementation
	//b.Brain.RegisterHandler(b.MessageRouter)

	b.Brain.RegisterHandler(b.StartupHook)
	b.Brain.RegisterHandler(b.ShutdownHook)
	b.Brain.RegisterHandler(b.CommandsRouter)

	b.Respond("postmortem(.+)?", b.Postmortem)
	b.Respond("rota", b.GetTodaysRota)

	return b, nil
}

func (b *Bot) StartupHook(evt joe.InitEvent) error {
	return nil
}

func (b *Bot) ShutdownHook(evt joe.ShutdownEvent) error {
	return nil
}

func (b *Bot) CommandsRouter(evt httpserver.RequestEvent) error {
	switch evt.URL.Path {
	case "/test":
		b.Say("#welcome", "Received test command!")
	}
	return nil
}

func (b *Bot) AtStartOfDay(StartOfDayEvent) {
	b.DailySetQuestionsChannelTopic()
	b.DailySendLevel1TheRunbook()
}

func (b *Bot) BeforeEndOfDay(BeforeEndOfDayEvent) {
	b.DailySendLevel1TheSignoffReminder()
}

func (b *Bot) AtEndOfDay(EndOfDayEvent) {
	b.DailySendLevel1TheCongratulationsMessage()
}
