package search

import (
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
	exists, err := s.ES.IndexExists(token).Do()
	if err != nil || !exists {
		return false
	}
	return true
}

// RepoExists checks if a repo has already been created
func (s *Store) RepoExists(token, name string) bool {
	exists, err := s.ES.TypeExists().Index(token).Type(name).Do()
	if err != nil || !exists {
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

// RepoSuggest serves as a json encoder for elastic search suggestions
type RepoSuggest struct {
	Name    string   `json:"name"`
	Suggest *suggest `json:"suggest"`
}

type suggest struct {
	Input  []string `json:"input"`
	Output string   `json:"output"`
}

// CreateRepository creates a new repository type
func (s *Store) CreateRepository(token string, r *Repository) error {
	if !s.UserExist(token) {
		return errors.New("no user exists for this token")
	}

	exist, err := s.ES.TypeExists().Index(token).Type("repository").Do()
	if !exist {
		// Build mapping for auto completion
		j := make(map[string]interface{})
		properties := make(map[string]interface{})
		name := make(map[string]string)
		suggest := make(map[string]string)
		name["type"] = "string"
		suggest["type"] = "completion"
		suggest["analyzer"] = "simple"
		suggest["search_analyzer"] = "simple"
		suggest["payloads"] = "true"
		properties["name"] = name
		properties["suggest"] = suggest
		j["properties"] = properties

		// Set mapping for autocompletion
		mapping, err := s.ES.PutMapping().
			Index(token).
			Type("repository").
			BodyJson(j).
			Do()
		if err != nil {
			return err
		} else if !mapping.Acknowledged {
			return errors.New("mapping not acknowledged")
		}

	}

	// Create repository suggestion
	rs := &RepoSuggest{
		Name: r.Name,
		Suggest: &suggest{
			Input:  []string{r.Name},
			Output: r.Name,
		},
	}

	// Index repository
	doc, err := s.ES.Index().
		Index(token).
		Type("repository").
		BodyJson(rs).
		Do()
	if err != nil {
		return err
	}
	fmt.Printf("Indexed repository %s to index %s, type %s\n", doc.Id, doc.Index, doc.Type)
	return nil
}

// GetRepositories retrieves a repository from the index
func (s *Store) GetRepositories(token, search string) ([]*Repository, error) {
	if !s.UserExist(token) {
		return nil, errors.New("no user exists for this token")
	} else if !s.RepoExists(token, "repository") {
		return nil, errors.New("no repository type exists for this token")
	}

	comp := elastic.NewCompletionSuggester("repository-suggest")
	comp.Field("suggest").Text(search)

	searchResult, err := s.ES.Suggest(token).Suggester(comp).Do()
	if err != nil {
		return nil, err
	}

	sug := searchResult["repository-suggest"][0].Options

	var repos []*Repository
	for _, value := range sug {
		repos = append(repos, &Repository{Name: value.Text})
	}

	return repos, nil
}
