package link

import (
	"os"
	"log"
	"time"
	"net/http"
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
	Id      string      `json:"id" firestore:"-"`
	Url     string      `json:"url" firestore:"url"`
	Created time.Time   `json:"created" firestore:"-"`
	Accessed time.Time  `json:"accessed" firestore:"-"`
	Updated time.Time   `json:"updated" firestore:"-"`
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
		Accessed: time.Time{},
		Updated: time.Time{},
	}
	return l, nil
}

type Server struct {
	PublicKey []byte
}

func NewServer() (*Server, error) {
	pubKey := os.Getenv("SHARED_TOKEN")
	if pubKey == "" {
		return nil, errors.New("Signing key missing set SHARED_TOKEN")
	}

	s := &Server{
		PublicKey: []byte(pubKey),
	}
	return s, nil
}

func (s *Server) Save(url string, r *http.Request) *Link {
	var link *Link = nil

	Client.RunTransaction(r.Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		q := Client.Collection("url").Where("url", "==", url)
		doc, err := q.Documents(r.Context()).Next()
		if err == nil {
			doc.DataTo(&link)
			tx.Update(doc.Ref, []firestore.Update{
				{Path: "updated", Value: time.Now()},
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
		log.Printf("Url '%s' not found: %s", url, err)

		link, err = NewLink(url)
		if err != nil {
			log.Printf("Could not create link")
			return nil
		}

		log.Printf("link.Id: %s", link.Id)
		ref := Client.Collection(COLLECTION_NAME).Doc(link.Id)
		if err = tx.Create(ref, link); err != nil {
			log.Printf("Error saving: %s", err)
			return nil
		}

		return nil
	})

	return link
}

func (s *Server) Get(id string, r *http.Request) *Link {
	var link *Link = nil

	Client.RunTransaction(r.Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		ref := Client.Collection(COLLECTION_NAME).Doc(id)
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