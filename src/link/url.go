package link

import (
	"log"
	"time"
	"context"

	"cloud.google.com/go/firestore"
	"github.com/teris-io/shortid"
)

type Link struct {
	Id      string    `json:"id" firestore:"id"`
	Url     string    `json:"url" firestore:"url"`
	Created time.Time `json:"created" firestore:"created"`
	Count	int64	  `json:"count" firestore:"count"`
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
}

func NewServer(client *firestore.Client, ctx context.Context) *Server {
	s := &Server{
		Client: client,
		Context: ctx,
	}
	return s
}

func (s *Server) Save(url string) *Link {
	var link *Link = nil

	s.Client.RunTransaction(s.Context, func(ctx context.Context, tx *firestore.Transaction) error {
		q := s.Client.Collection("url").Where("url", "==", url)
		doc, err := q.Documents(s.Context).Next()
		if err == nil {
			doc.DataTo(&link)
			return nil
		}
		log.Printf("Url '%s' not found: %s", url, err)

		link, err := NewLink(url)
		if err != nil {
			log.Printf("Could not create link")
			return nil
		}

		ref := s.Client.Collection("url").Doc(link.Id)
		err = tx.Create(ref, link)
		if err != nil {
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
		ref := s.Client.Collection("url").Doc(id)
		doc, err := tx.Get(ref)
		if err != nil {
			log.Printf("Could not find url '%s': %s", id, err)
			return nil
		}

		tx.Update(ref, []firestore.Update{
			{Path: "count", Value: firestore.Increment(1)},
		})

		doc.DataTo(&link)
		return nil
	})

	return link
}
