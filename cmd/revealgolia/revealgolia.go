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

	"github.com/ViBiOh/flags"
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

func getRequest(app, key string) request.Request {
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

func configIndex(ctx context.Context, request request.Request, app, index string) error {
	settings := map[string]any{
		"searchableAttributes": []string{"keywords", "img", "content"},
	}

	_, err := request.Put(getURL(app, "/1/indexes/%s/settings", index)).JSON(ctx, settings)
	return err
}

func clearIndex(ctx context.Context, request request.Request, app, index string) error {
	_, err := request.Post(getURL(app, "/1/indexes/%s/clear", index)).Send(ctx, nil)
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

func saveObjects(ctx context.Context, request request.Request, app, index string, objects []Item) error {
	requests := make([]BatchAction, len(objects))
	for index, object := range objects {
		requests[index] = BatchAction{
			Action: "addObject",
			Body:   object,
		}
	}

	_, err := request.Post(getURL(app, "/1/indexes/%s/batch", index)).JSON(ctx, Batch{requests})
	return err
}

func main() {
	fs := flag.NewFlagSet("revealgolia", flag.ExitOnError)

	loggerConfig := logger.Flags(fs, "logger")

	app := flags.String(fs, "", "algolia", "app", "Application", "", nil)
	key := flags.String(fs, "", "algolia", "key", "Key", "", nil)
	index := flags.String(fs, "", "algolia", "index", "Index", "", nil)
	source := flags.String(fs, "", "reveal", "source", "Walked markdown directory", "", nil)
	prefixFromFolder := flags.Bool(fs, "", "reveal", "prefixFromFolder", "Use name of folder as URL prefix", false, nil)
	sep := flags.String(fs, "", "reveal", "sep", "Separator", "^\n\n\n", nil)
	vsep := flags.String(fs, "", "reveal", "verticalSep", "Vertical separator", "^\n\n", nil)

	debug := flags.Bool(fs, "", "app", "debug", "Debug output instead of sending them", false, nil)

	logger.Fatal(fs.Parse(os.Args[1:]))
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	ctx := context.Background()

	sepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *sep))
	vsepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *vsep))
	req := getRequest(*app, *key)

	if !*debug {
		logger.Fatal(clearIndex(ctx, req, *app, *index))
		logger.Fatal(configIndex(ctx, req, *app, *index))
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
		} else if err := saveObjects(ctx, req, *app, *index, objects); err != nil {
			return err
		}

		return nil
	}))

	logger.Info("Done!")
}
