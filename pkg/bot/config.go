package bot

import (
	"fmt"
	httpserver "github.com/dombo/hiberBot/pkg/bot/custom-http-server"
	"github.com/go-joe/joe"
	"github.com/go-joe/slack-adapter/v2"
	"github.com/spf13/viper"
)

// Config holds all parameters to setup a new chat bot.
type Config struct {
	Slack    SlackConfig
	Google	GoogleConfig
	HTTP     HTTPConfig
}

type SlackConfig struct {
	Token string // required slack token
	BotName string // required name of the slack bot
	VerificationToken string `mapstructure:"verification_token"` // required slack EventsAPI verification token
	ListenAddr string // optional port to receive event callbacks on
	Debug bool	// optional enable to debug slack connection
}

type HTTPConfig struct {
	ListenAddr string // optional HTTP listen address to receive command callbacks
}

type GoogleConfig struct {
	Calendar CalendarConfig
	Docs     DocsConfig
	Drive    DriveConfig
} // TODO Offer global service key configuration with overrides and fallback support to nested config objects

type CalendarConfig struct {
	User           string                // required calendar user to operate as
	RotaCalendarId string                `mapstructure:"rota_calendar_id"`
	Service        GoogleCredentialsFile // required the json service account credentials file created in GCP
}

type DocsConfig struct {
	User string // required docs user to operate as
	Service        GoogleCredentialsFile
}

type DriveConfig struct {
	User string // required drive user to operate as
	PostmortemFileId string `mapstructure:"postmortem_fild_id"`
	Service        GoogleCredentialsFile
}

// Publicly exported variant of golang.org/x/oauth2/google/google.go:99 credentialsFile
// mapstructure required to work with json keys containing _
type GoogleCredentialsFile struct {
	Type string `mapstructure:"type"` // serviceAccountKey or userCredentialsKey

	// Service Account fields
	ClientEmail  string `mapstructure:"client_email"`
	PrivateKeyID string `mapstructure:"private_key_id"`
	PrivateKey   string `mapstructure:"private_key"`
	TokenURL     string `mapstructure:"token_uri"`
	ProjectID    string `mapstructure:"project_id"`

	// User Credential fields
	// (These typically come from gcloud auth.)
	ClientSecret string `mapstructure:"client_secret"`
	ClientID     string `mapstructure:"client_id"`
	RefreshToken string `mapstructure:"refresh_token"`
}


func GetConf() *Config {

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	// These values will not be set on the Config struct but when called through viper they will be provided
	viper.SetDefault("slack.debug", false)

	viper.SetDefault("http.listenaddr", ":9192")
	viper.SetDefault("slack.listenaddr", ":9191")


	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	conf := Config{}
	err = viper.Unmarshal(&conf)
	if err != nil {
		panic(fmt.Errorf("Fatal error unmarshalling config: %s \n", err))
	}

	return &conf
}

// Modules creates a list of joe.Modules that can be used with this configuration.
func (conf Config) Modules() []joe.Module {
	var modules []joe.Module

	modules = append(modules, slack.EventsAPIAdapter(viper.GetString("slack.listenaddr"),
		viper.GetString("slack.token"),
		viper.GetString("slack.verification_token"),
		slack.WithDebug(viper.GetBool("slack.debug"))))

	modules = append(modules, httpserver.Server(viper.GetString("http.listenaddr")))

	return modules
}

func (conf Config) Validate() error {
	//if conf.HTTPListen == "" {
	//	return errors.New("missing HTTP listen address")
	//}
	return nil
}
