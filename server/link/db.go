package link

import (
	"os"
	"log"
	"context"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"cloud.google.com/go/firestore"
)

var Client *firestore.Client

func init() {
	auth_key := os.Getenv("AUTH_KEY")
	project_id := os.Getenv("PROJECT_ID")
	ctx := context.Background()
	var app *firebase.App = nil
	var err error

	if auth_key != "" {
		sa := option.WithCredentialsFile(auth_key)
		app, err = firebase.NewApp(ctx, nil, sa)
		if err != nil {
			log.Printf("Error loading credentials: %s", err)
		}
	}
	
	if app == nil && project_id != "" {
		conf := &firebase.Config{ProjectID: project_id}
		app, err = firebase.NewApp(ctx, conf)
		if err != nil {
			log.Printf("Error initializing app: %s", err)
		}
	}

	if app == nil {
		log.Fatalln("Must define AUTH_KEY or PROJECT_ID")
	}
	
	Client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}
