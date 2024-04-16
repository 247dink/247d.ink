package link

import (
	"os"
	"log"
	"context"

	firebase "firebase.google.com/go/v4"
	"cloud.google.com/go/firestore"
)

var Client *firestore.Client

func init() {
	project_id := os.Getenv("FIRESTORE_PROJECT_ID")
	ctx := context.Background()
	var err error

	if project_id == "" {
		log.Fatalln("Must define FIRESTORE_PROJECT_ID")
		return
	}

	conf := &firebase.Config{ProjectID: project_id}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Printf("Error initializing app: %s", err)
	}

	Client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}
