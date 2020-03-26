package main

import (
	"context"
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

	"github.com/ViBiOh/httputils/v3/pkg/request"
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
	URL      string   `json:"url"`
	H        int      `json:"h"`
	V        int      `json:"v"`
	Content  string   `json:"content"`
	Chapter  string   `json:"chapter"`
	Keywords []string `json:"keywords"`
	Img      string   `json:"img"`
}

func getRequest(app, key string) *request.Request {
	return request.New().Header("X-Algolia-Application-Id", app).Header("X-Algolia-API-Key", key)
}

func getURL(app, path, index string) string {
	return fmt.Sprintf(fmt.Sprintf("https://%s.algolia.net%s", app, path), index)
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

func configIndex(request *request.Request, app, index string) error {
	settings := map[string]interface{}{
		"searchableAttributes": []string{"keywords", "img", "content"},
	}

	_, err := request.Put(getURL(app, "/1/indexes/%s/settings", index)).JSON(context.Background(), settings)
	return err
}

func clearIndex(request *request.Request, app, index string) error {
	_, err := request.Post(getURL(app, "/1/indexes/%s/clear", index)).Send(context.Background(), nil)
	return err
}

func debugObject(objects Item) error {
	output, err := json.MarshalIndent(objects, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("%s\n", output)
	return nil
}

func saveObject(request *request.Request, app, index string, object Item) error {
	_, err := request.Post(getURL(app, "/1/indexes/%s", index)).JSON(context.Background(), object)
	return err
}

func main() {
	fs := flag.NewFlagSet("revealgolia", flag.ExitOnError)

	app := fs.String("app", "", "[algolia] App")
	key := fs.String("key", "", "[algolia] Key")
	index := fs.String("index", "", "[algolia] Index")
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

	if !*debug {
		if err := clearIndex(getRequest(*app, *key), *app, *index); err != nil {
			log.Fatal(err)
		}

		if err := configIndex(getRequest(*app, *key), *app, *index); err != nil {
			log.Fatal(err)
		}
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
		for _, object := range objects {
			if *debug {
				if err := debugObject(object); err != nil {
					return err
				}
			} else if err := saveObject(getRequest(*app, *key), *app, *index, object); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
