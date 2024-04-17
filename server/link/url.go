package link

import (
	"log"
	"time"
	"net/http"
	"context"

	"cloud.google.com/go/firestore"
	"github.com/teris-io/shortid"
)


const COLLECTION_NAME string = "url"


type Link struct {
	Id          string      `json:"id" firestore:"-"`
	Url         string      `json:"url" firestore:"url"`
	TTL			time.Time	`json:"ttl" firestore:"ttl"`
	Created     time.Time   `json:"created" firestore:"-"`
	Accessed    time.Time   `json:"accessed" firestore:"-"`
	Updated     time.Time   `json:"updated" firestore:"-"`
	AccessCount	int64	    `json:"accessCount" firestore:"accessCount"`
	UpdateCount	int64	    `json:"updateCount" firestore:"updateCount"`
}

func NewLink(url string) (*Link, error) {
	id, err := shortid.Generate()
	if err != nil {
		return nil, err
	}

	l := &Link{
		Url: url,
		TTL: time.Time{},
		AccessCount: 0,
		UpdateCount: 0,
		Id: id,
		Created: time.Now(),
		Accessed: time.Time{},
		Updated: time.Time{},
	}
	return l, nil
}

type Server struct {
}

func NewServer() (*Server, error) {
	s := &Server{}
	return s, nil
}

func (s *Server) Save(url string, ttl int, r *http.Request) (*Link, error) {
	var link *Link = nil

	var ttlValue time.Time = time.Time{}
	if ttl != 0 {
		ttlValue = time.Now().Add(time.Hour * time.Duration(ttl * 24))
	}

	err := Client.RunTransaction(r.Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		q := Client.Collection(COLLECTION_NAME).Where("url", "==", url).Limit(1)
		doc, err := q.Documents(ctx).Next()
		if err == nil {
			log.Printf("Url exists")
			doc.DataTo(&link)
			tx.Update(doc.Ref, []firestore.Update{
				{Path: "updateCount", Value: firestore.Increment(1)},
				{Path: "ttl", Value: ttlValue},
			})
			link.Id = doc.Ref.ID
			link.Created = doc.CreateTime
			if !doc.ReadTime.IsZero() {
				link.Accessed = doc.ReadTime
			}
			if !doc.UpdateTime.IsZero() {
				link.Updated = doc.UpdateTime
			}
			return nil
		}

		log.Printf("Adding new url %s", url)
		link, err = NewLink(url)
		if err != nil {
			log.Printf("Could not create link")
			return err
		}

		link.TTL = ttlValue

		log.Printf("link.Id: %s", link.Id)
		ref := Client.Collection(COLLECTION_NAME).Doc(link.Id)
		if err = tx.Create(ref, link); err != nil {
			log.Printf("Error saving: %s", err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("Error running transaction: %s", err)
		return nil, err
	}

	return link, nil
}

func (s *Server) Get(id string, r *http.Request) (*Link, error) {
	var link *Link = nil

	err := Client.RunTransaction(r.Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		ref := Client.Collection(COLLECTION_NAME).Doc(id)
		doc, err := tx.Get(ref)
		if err != nil {
			log.Printf("Could not find url '%s': %s", id, err)
			return err
		}

		tx.Update(ref, []firestore.Update{
			{Path: "accessCount", Value: firestore.Increment(1)},
		})

		doc.DataTo(&link)
		link.Id = doc.Ref.ID
		link.Created = doc.CreateTime
		if !doc.ReadTime.IsZero() {
			link.Accessed = doc.ReadTime
		}
		if !doc.UpdateTime.IsZero() {
			link.Updated = doc.UpdateTime
		}
		return nil
	})

	if err != nil {
		log.Printf("Error running transaction: %s", err)
		return nil, err
	}

	return link, nil
}

func (s *Server) Check(r *http.Request) error {
	q := Client.Collection(COLLECTION_NAME).Where("url", "==", "https://www.google.com/").Limit(1)
	_, err := q.Documents(r.Context()).Next()
	return err
}
