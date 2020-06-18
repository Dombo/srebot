module github.com/dombo/srebot

go 1.14

require (
	github.com/go-joe/cron v1.1.0
	github.com/go-joe/joe v0.9.0
	github.com/go-joe/slack-adapter/v2 v2.0.1-0.20200611132151-e305d7279be0
	github.com/jinzhu/now v1.1.1
	github.com/slack-go/slack v0.6.5
	github.com/spf13/viper v1.7.0
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.26.0
)

replace github.com/go-joe/slack-adapter/v2 => /home/dom/Code/go-joe/slack-adapter
