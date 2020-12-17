package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/google/go-github/github"
	"io"
	"net/http"
	"path"
	"strings"
)

type githubFetcher struct {
	client *github.Client
	owner  string
	repo   string
	branch string
}

func (gh *githubFetcher) fetchGithubFileContents(path string) (io.ReadCloser, error) {
	content, _, _, err := gh.client.Repositories.GetContents(context.Background(), gh.owner, gh.repo, path, nil)
	if err != nil {
		return nil, err
	}
	name := *content.Name
	if !strings.HasSuffix(name, ".csv") {
		return nil, fmt.Errorf("file %s is not a csv file", name)
	} else {
		resp, err := http.Get(*content.DownloadURL)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}
}

func (gh *githubFetcher) fetchGithubDirContents(path string) ([]string, error) {
	_, dirContent, _, err := gh.client.Repositories.GetContents(context.Background(), gh.owner, gh.repo, path, nil)
	if err != nil {
		return nil, err
	}

	fileNames := make([]string, 5)
	for _, content := range dirContent {
		fileNames = append(fileNames, *content.Name)
	}

	return fileNames, nil
}

func (gh *githubFetcher) fetchGithubFileContentsViaDataApi(path string) (string, error) {
	branch, _, err := gh.client.Repositories.GetBranch(context.Background(), gh.owner, gh.repo, gh.branch)
	if err != nil {
		return "", err
	}
	commit := branch.GetCommit()
	tree, _, err := gh.client.Git.GetTree(context.Background(), gh.owner, gh.repo, *commit.SHA, true)
	if err != nil {
		return "", err
	}

	var file github.TreeEntry
	found := false
	for _, f := range tree.Entries {
		if *f.Path == path {
			file = f
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("the path %s is not found in the tree", path)
	}

	blob, _, err := gh.client.Git.GetBlob(context.Background(), gh.owner, gh.repo, *file.SHA)
	if err != nil {
		return "", err
	}

	if blob.GetEncoding() == "base64" {
		b, err := base64.StdEncoding.DecodeString(blob.GetContent())
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return blob.GetContent(), nil
}

func (gh *githubFetcher) getPlaces() ([]*Place, error) {
	r, err := gh.fetchGithubFileContents("csse_covid_19_data/UID_ISO_FIPS_LookUp_Table.csv")
	if err != nil {
		return nil, err
	}
	var places []*Place
	if err := gocsv.Unmarshal(r, &places); err != nil {
		return nil, err
	}
	return places, nil
}

type NumberType string

const (
	CONFIRMED  NumberType = "confirmed"
	DEATHS     NumberType = "deaths"
	RECOVERIES NumberType = "recoveries"
)

type PlaceType string

const (
	US     PlaceType = "US"
	GLOBAL PlaceType = "global"
)

func (gh *githubFetcher) getNumbers(numberType NumberType, placeType PlaceType) error {
	if placeType == US && numberType == RECOVERIES {
		return errors.New("united states does not report recovery counts")
	}
	fileName := fmt.Sprintf("time_series_covid19_%s_%s.csv", numberType, placeType)
	filePath := path.Join("csse_covid_19_data", "csse_covid_19_time_series", fileName)

	contents, err := gh.fetchGithubFileContentsViaDataApi(filePath)
	if err != nil {
		return err
	}

	fmt.Print(contents)

	return nil
}
