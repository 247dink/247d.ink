package link

import (
	"os"
	"log"
	"time"
	"errors"
	"context"
	"encoding/base64"
	"crypto/hmac"
	"crypto/sha256"

	"cloud.google.com/go/firestore"
	"github.com/teris-io/shortid"
)


const COLLECTION_NAME string = "url"


type Link struct {
	Id      string      `json:"id" firestore:"id"`
	Url     string      `json:"url" firestore:"url"`
	Created time.Time   `json:"created" firestore:"created"`
	Accessed time.Time  `json:"accessed" firestore:"accessed"`
	Updated time.Time   `json:"updated" firestore:"updated"`
	Count	int64	    `json:"count" firestore:"count"`
}

func NewLink(url string) (*Link, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	l := &Link{
		Url: url,
		Count: 0,
		Id: id,
		Created: time.Now(),
	}
	return l, nil
}

type Server struct {
	*firestore.Client
	context.Context
	PublicKey []byte
}

func NewServer(client *firestore.Client, ctx context.Context) (*Server, error) {
	pubKey := os.Getenv("SHARED_TOKEN")
	if pubKey == "" {
		return nil, errors.New("Signing key missing set SHARED_TOKEN")
	}

	s := &Server{
		Client: client,
		Context: ctx,
		PublicKey: []byte(pubKey),
	}
	return s, nil
}

func (s *Server) Save(url string) *Link {
	var link *Link = nil

	s.Client.RunTransaction(s.Context, func(ctx context.Context, tx *firestore.Transaction) error {
		q := s.Client.Collection("url").Where("url", "==", url)
		doc, err := q.Documents(s.Context).Next()
		if err == nil {
			doc.DataTo(&link)
			ref := s.Client.Collection(COLLECTION_NAME).Doc(link.Id)
			tx.Update(ref, []firestore.Update{
				{Path: "updated", Value: time.Now()},
			})
			return nil
		}
		log.Printf("Url '%s' not found: %s", url, err)

		link, err := NewLink(url)
		if err != nil {
			log.Printf("Could not create link")
			return nil
		}

		log.Printf("link.Id: %s", link.Id)
		ref := s.Client.Collection(COLLECTION_NAME).Doc(link.Id)
		if err = tx.Create(ref, link); err != nil {
			log.Printf("Error saving: %s", err)
			return nil
		}

		return nil
	})

	return link
}

func (s *Server) Get(id string) *Link {
	var link *Link = nil

	s.Client.RunTransaction(s.Context, func(ctx context.Context, tx *firestore.Transaction) error {
		ref := s.Client.Collection(COLLECTION_NAME).Doc(id)
		doc, err := tx.Get(ref)
		if err != nil {
			log.Printf("Could not find url '%s': %s", id, err)
			return nil
		}

		tx.Update(ref, []firestore.Update{
			{Path: "count", Value: firestore.Increment(1)},
			{Path: "accessed", Value: time.Now()},
		})

		doc.DataTo(&link)
		return nil
	})

	return link
}

func (s *Server) CheckSignature(url string, signature string) bool {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, s.PublicKey)
	mac.Write([]byte(url))
	xMAC := mac.Sum(nil)

	return hmac.Equal(sig, xMAC)
}