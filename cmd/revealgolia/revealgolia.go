package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

var (
	// ErrIndexNotFound occurs when index is not found in List
	ErrIndexNotFound = errors.New("index not found")

	chapterTitleRegex = regexp.MustCompile(`^#\s+(.*)\s*`)
	imgRegex          = regexp.MustCompile(`\[\]\((.*)\)`)
	strongRegex       = regexp.MustCompile(`\*\*\s*([^*]*)\s*\*\*`)
	italicRegex       = regexp.MustCompile(`\*\s*([^*]*)\s*\*`)
)

// Item store indexed item
type Item struct {
	ObjectID string   `json:"objectID"`
	URL      string   `json:"url"`
	H        int      `json:"h"`
	V        int      `json:"v"`
	Content  string   `json:"content"`
	Chapter  string   `json:"chapter"`
	Keywords []string `json:"keywords"`
	Img      string   `json:"img"`
}

// getSearchObjects transform input reveal file to algolia object
func getSearchObjects(name, source string, sep, verticalSep *regexp.Regexp) ([]Item, error) {
	content, err := ioutil.ReadFile(source)
	if err != nil {
		return nil, err
	}

	objects := make([]Item, 0)
	index := 1

	var chapterName string
	var slideImg string
	var keywords []string

	contentStr := string(content)
	for chapterNum, chapter := range sep.Split(contentStr, -1) {
		if title := chapterTitleRegex.FindStringSubmatch(chapter); len(title) > 1 {
			chapterName = title[1]
		}

		for slideNum, slide := range verticalSep.Split(chapter, -1) {
			slideImg = ""
			if matches := imgRegex.FindStringSubmatch(slide); len(matches) > 1 {
				slideImg = matches[1]
			}

			keywords = make([]string, 0)
			if matches := strongRegex.FindStringSubmatch(slide); len(matches) > 1 {
				keywords = append(keywords, matches[1:]...)
			}
			if matches := italicRegex.FindStringSubmatch(slide); len(matches) > 1 {
				keywords = append(keywords, matches[1:]...)
			}

			objects = append(objects, Item{
				ObjectID: fmt.Sprintf("%s_%d", name, index),
				URL:      path.Join("/", name, fmt.Sprintf("/#/%d/%d", chapterNum, slideNum)),
				H:        chapterNum,
				V:        slideNum,
				Content:  slide,
				Chapter:  chapterName,
				Keywords: keywords,
				Img:      slideImg,
			})
			index++
		}
	}

	return objects, nil
}

func configIndex(index *search.Index) error {
	_, err := index.SetSettings(search.Settings{
		SearchableAttributes: opt.SearchableAttributes("keywords", "img", "content"),
	})
	return err
}

func saveObjects(objects []Item, debug bool, index *search.Index) error {
	if debug {
		output, err := json.MarshalIndent(objects, "", "  ")
		if err != nil {
			return err
		}

		log.Printf("%s\n", output)
		return nil
	}

	_, err := index.SaveObjects(objects)
	return err
}

func main() {
	fs := flag.NewFlagSet("revealgolia", flag.ExitOnError)

	app := fs.String("app", "", "[algolia] App")
	key := fs.String("key", "", "[algolia] Key")
	indexName := fs.String("index", "", "[algolia] Index")
	source := fs.String("source", "", "[reveal] Walked markdown directory")
	prefixFromFolder := fs.Bool("prefixFromFolder", false, "[reveal] Use name of folder as URL prefix")
	sep := fs.String("sep", "^\n\n\n", "[reveal] Separator")
	vsep := fs.String("verticalSep", "^\n\n", "[reveal] Vertical separator")

	debug := fs.Bool("debug", false, "Debug output instead of sending them")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	sepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *sep))
	vsepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *vsep))

	client := search.NewClient(*app, *key)
	index := client.InitIndex(*indexName)

	if _, err := index.Delete(); err != nil {
		log.Fatal(err)
	}

	if err := configIndex(index); err != nil {
		log.Fatal(err)
	}

	err := filepath.Walk(*source, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}

		name := ""
		if *prefixFromFolder {
			name = filepath.Base(filepath.Dir(path))
		}

		objects, err := getSearchObjects(name, path, sepRegex, vsepRegex)
		if err != nil {
			return err
		}

		log.Printf("%d objects found in %s\n", len(objects), info.Name())
		return saveObjects(objects, *debug, index)
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
