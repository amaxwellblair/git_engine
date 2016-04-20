package search

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/olivere/elastic.v3"
)

// Store holds the Elastic Search client
type Store struct {
	ES *elastic.Client
}

// NewStore returns a new instance of store
func NewStore() *Store {
	return &Store{
		ES: MustOpenConnection(),
	}
}

// MustOpenConnection will create a client for Elastic Search
func MustOpenConnection() *elastic.Client {
	c, err := elastic.NewClient()
	if err != nil {
		panic(err)
	}
	return c
}

// UserExist checks if a user has already been created
func (s *Store) UserExist(token string) bool {
	_, err := s.ES.IndexExists(token).Do()
	if err != nil {
		return false
	}
	return true
}

// CreateUserIndex creates a new index
func (s *Store) CreateUserIndex(token string) error {
	if _, err := s.ES.CreateIndex(token).Do(); err != nil {
		return err
	}
	return nil
}

// CreateRepository creates a new repository type
func (s *Store) CreateRepository(token string, r *Repository) error {
	if !s.UserExist(token) {
		return errors.New("no user exists for this token")
	}

	doc, err := s.ES.Index().
		Index(token).
		Type("repository").
		BodyJson(r).
		Do()
	if err != nil {
		return err
	}
	fmt.Printf("Indexed repository %s to index %s, type %s\n", doc.Id, doc.Index, doc.Type)
	return nil
}

// GetRepositories retrieves a repository from the index
func (s *Store) GetRepositories(token string) ([]*Repository, error) {
	searchResult, err := s.ES.Search().
		Index(token).
		Type("repository").
		From(0).Size(1000).
		Sort("name", true).
		Do()
	if err != nil {
		return nil, err
	}

	var repos []*Repository
	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			var r Repository
			err := json.Unmarshal(*hit.Source, &r)
			if err != nil {
				return nil, err
			}
			repos = append(repos, &r)
		}
	} else {
		return nil, errors.New("no repositories found")
	}

	return repos, nil
}
