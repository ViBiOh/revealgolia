package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var (
	chapterTitleRegex = regexp.MustCompile(`^#\s+(.*)\s*`)
	imgRegex          = regexp.MustCompile(`\[]\((.*)\)`)
	strongRegex       = regexp.MustCompile(`\*\*\s*([^*]*)\s*\*\*`)
	italicRegex       = regexp.MustCompile(`\*\s*([^*]*)\s*\*`)
)

// Batch contains batchs actions
type Batch struct {
	Requests []BatchAction `json:"requests"`
}

// BatchAction contains action to perform inside a batch
type BatchAction struct {
	Action string `json:"action"`
	Body   Item   `json:"body"`
}

// Item store indexed item
type Item struct {
	URL      string   `json:"url"`
	Content  string   `json:"content"`
	Chapter  string   `json:"chapter"`
	Img      string   `json:"img"`
	Keywords []string `json:"keywords"`
	H        int      `json:"h"`
	V        int      `json:"v"`
}

func getRequest(app, key string) *request.Request {
	return request.New().Header("X-Algolia-Application-Id", app).Header("X-Algolia-API-Key", key)
}

func getURL(app, path, index string) string {
	return fmt.Sprintf(fmt.Sprintf("https://%s.algolia.net%s", app, path), index)
}

// getSearchObjects transform input reveal file to algolia object
func getSearchObjects(name, source string, sep, verticalSep *regexp.Regexp) ([]Item, error) {
	content, err := os.ReadFile(source)
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

func debugObjects(objects []Item) error {
	output, err := json.MarshalIndent(objects, "", "  ")
	if err != nil {
		return err
	}

	logger.Info("%s\n", output)
	return nil
}

func saveObjects(request *request.Request, app, index string, objects []Item) error {
	requests := make([]BatchAction, len(objects))
	for index, object := range objects {
		requests[index] = BatchAction{
			Action: "addObject",
			Body:   object,
		}
	}

	_, err := request.Post(getURL(app, "/1/indexes/%s/batch", index)).JSON(context.Background(), Batch{requests})
	return err
}

func main() {
	fs := flag.NewFlagSet("revealgolia", flag.ExitOnError)

	loggerConfig := logger.Flags(fs, "logger")

	app := flags.New("", "algolia").Name("app").Default("").Label("Application").ToString(fs)
	key := flags.New("", "algolia").Name("key").Default("").Label("Key").ToString(fs)
	index := flags.New("", "algolia").Name("index").Default("").Label("Index").ToString(fs)
	source := flags.New("", "reveal").Name("source").Default("").Label("Walked markdown directory").ToString(fs)
	prefixFromFolder := flags.New("", "reveal").Name("prefixFromFolder").Default(false).Label("Use name of folder as URL prefix").ToBool(fs)
	sep := flags.New("", "reveal").Name("sep").Default("^\n\n\n").Label("Separator").ToString(fs)
	vsep := flags.New("", "reveal").Name("verticalSep").Default("^\n\n").Label("Vertical separator").ToString(fs)

	debug := flags.New("", "app").Name("debug").Default(false).Label("Debug output instead of sending them").ToBool(fs)

	logger.Fatal(fs.Parse(os.Args[1:]))
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	sepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *sep))
	vsepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *vsep))

	if !*debug {
		logger.Fatal(clearIndex(getRequest(*app, *key), *app, *index))
		logger.Fatal(configIndex(getRequest(*app, *key), *app, *index))
	}

	logger.Fatal(filepath.Walk(*source, func(path string, info os.FileInfo, _ error) error {
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

		logger.Info("%d objects found in %s", len(objects), info.Name())
		if *debug {
			if err := debugObjects(objects); err != nil {
				return err
			}
		} else if err := saveObjects(getRequest(*app, *key), *app, *index, objects); err != nil {
			return err
		}

		return nil
	}))

	logger.Info("Done!")
}
