package google

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"

	"golang.org/x/oauth2"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Requires a Google Cloud Service Account: https://support.google.com/a/answer/7378726?hl=en
// Enable the APIs in https://console.cloud.google.com/apis/library?filter=category:gsuite
// Authorize the Service Account inside your google domain for specific scopes: https://admin.google.com/ac/owl/domainwidedelegation?hl=en

// Build and returns a oauth2.TokenSource (and associated refresh context) that
// acts on behalf of the specified subject with provided scopes
func authenticate(subject string, scopes []string, viperKey string) (oauth2.TokenSource, context.Context) {
	ctx := context.Background()

	// Unfortunately google does not export the method to create a JWT from config you've already parsed
	// So we have to marshal back to JSON
	serviceCredentials := viper.Get(viperKey)
	bs, err := json.Marshal(serviceCredentials)
	if err != nil {
		log.Fatalf("unable to marshal config to JSON: %v", err)
	}

	config, err := google.JWTConfigFromJSON(bs, scopes...)
	if err != nil {
		log.Fatalf("failed to create JWTConfigFromJSON: %v", err)
	}
	config.Subject = subject

	ts := config.TokenSource(ctx)

	return ts, ctx
}

func NewCalendarService(subject string, scopes []string) *calendar.Service {
	ts, ctx := authenticate(subject, scopes, "google.calendar.service")

	srv, err := calendar.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create new service: %v", err))
	}

	return srv
}

func NewDocsService(subject string, scopes []string) *docs.Service {
	ts, ctx := authenticate(subject, scopes, "google.docs.service")

	srv, err := docs.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create new service: %v", err))
	}

	return srv
}

func NewDriveService(subject string, scopes []string) *drive.Service {
	ts, ctx := authenticate(subject, scopes, "google.drive.service")

	srv, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create new service: %v", err))
	}

	return srv
}
