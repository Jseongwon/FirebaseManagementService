package firebase

import (
	"context"
	"log"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var firebaseAuth *auth.Client

func InitFirebase(credentialPath string) {
	opt := option.WithCredentialsFile(credentialPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Firebase init error: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Firebase auth error: %v", err)
	}

	firebaseAuth = client
}

func CreateCustomToken(uid string) (string, error) {
	return firebaseAuth.CustomToken(context.Background(), uid)
}
