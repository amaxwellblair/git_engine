package search

import (
	"errors"
	"fmt"
	"strconv"

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

	j := indexSettingsAndMapping()

	if _, err := s.ES.CreateIndex(token).BodyJson(j).Do(); err != nil {
		return err
	}
	return nil
}

// RepoSuggest serves as a json encoder for elastic search suggestions
type RepoSuggest struct {
	Name    string   `json:"name"`
	Active  bool     `json:"active"`
	ID      int      `json:"id"`
	Suggest *suggest `json:"suggest"`
}

type suggest struct {
	Input  []string `json:"input"`
	Output string   `json:"output"`
}

// CreateRepositoryList creates a new repository type
func (s *Store) CreateRepositoryList(token string, r *Repository) error {
	if !s.UserExist(token) {
		return errors.New("no user exists for this token")
	}

	exist, err := s.ES.TypeExists().Index(token).Type("repository").Do()
	if !exist {
		if err := s.createAutoCompleteMapping(token); err != nil {
			return err
		}
	}

	// Create repository suggestion
	rs := &RepoSuggest{
		Name:   r.Name,
		Active: r.Active,
		Suggest: &suggest{
			Input:  []string{r.Name},
			Output: r.Name,
		},
	}

	// Index repository
	doc, err := s.ES.Index().
		Id(strconv.Itoa(r.ID)).
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

// CreateRepository creates a repository for commits
func (s *Store) CreateRepository(name, owner, token string, commits []*GitCommit) error {
	if !s.UserExist(token) {
		return errors.New("no user exists for this token")
	}

	// Index commits
	for _, commit := range commits {
		row := &IndexCommit{
			Message: commit.Commit.Message,
			URL:     commit.HTML,
		}
		doc, err := s.ES.Index().
			Index(token).
			Type(name).
			BodyJson(row).
			Do()
		if err != nil {
			return err
		}
		fmt.Printf("Indexed commit %s to index %s, repository %s\n", doc.Id, doc.Index, doc.Type)
	}

	return nil
}

// ActivateRepository activates a repository
func (s *Store) ActivateRepository(token, repoName string) error {

	return nil
}

// GetCommits returns commits for a given repository
func (s *Store) GetCommits(token, repoName, substring string) ([]*GitCommit, error) {
	// if !s.UserExist(token) {
	// 	return nil, errors.New("no user exists for this token")
	// } else if !s.RepoExists(token, repoName) {
	// 	return nil, errors.New("no repository named " + repoName + " type exists for this token")
	// }
	//
	// query := elastic.NewMatchQuery("_all", substring)
	// searchResult, err := s.ES.Search(token).
	// 	Index(token).
	// 	Type(repoName).
	// 	Query(query).
	// 	Do()
	// if err != nil {
	// 	return nil, err
	// }
	//
	return nil, nil
}

// IndexCommit contains the elements of the document to be indexed
type IndexCommit struct {
	Message string `json:"commit_message"`
	URL     string `json:"html_url"`
}

// Search contains search terms for a ES query
type Search struct {
	Query *match `json:"match"`
}

type match struct {
	All string `json:"_all"`
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

func (s *Store) createAutoCompleteMapping(token string) error {
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

	return nil
}

func indexSettingsAndMapping() map[string]interface{} {
	// Build settings for n-gram tokenizer
	j := make(map[string]interface{})
	settings := make(map[string]interface{})
	mappings := make(map[string]interface{})
	// settings["number_of_shards"] = 1

	analysis := make(map[string]interface{})
	filter := make(map[string]interface{})
	ngramFilter := make(map[string]interface{})
	ngramFilter["type"] = "ngram"
	ngramFilter["min_gram"] = 3
	ngramFilter["max_gram"] = 3

	analyzer := make(map[string]interface{})
	ngramAnalyzer := make(map[string]interface{})
	ngramAnalyzer["type"] = "custom"
	ngramAnalyzer["tokenizer"] = "standard"
	ngramAnalyzer["filter"] = []string{"lowercase", "ngram_filter"}

	analyzer["ngram_analyzer"] = ngramAnalyzer
	filter["ngram_filter"] = ngramFilter
	analysis["filter"] = filter
	analysis["analyzer"] = analyzer

	settings["analysis"] = analysis

	typeName := make(map[string]interface{})
	all := make(map[string]interface{})
	all["type"] = "string"
	all["index_analyzer"] = "ngram_analyzer"
	all["search_analyzer"] = "standard"

	properties := make(map[string]interface{})
	commitMessage := make(map[string]interface{})
	commitMessage["type"] = "string"
	commitMessage["include_in_all"] = "true"
	commitMessage["index_analyzer"] = "ngram_analyzer"
	commitMessage["search_analyzer"] = "standard"

	properties["commit_message"] = commitMessage
	typeName["properties"] = properties
	typeName["_all"] = all

	mappings["_default_"] = typeName

	j["mappings"] = mappings
	j["settings"] = settings

	return j
}
