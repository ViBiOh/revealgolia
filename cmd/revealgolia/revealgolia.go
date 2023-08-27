package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
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

type Batch struct {
	Requests []BatchAction `json:"requests"`
}

type BatchAction struct {
	Action string `json:"action"`
	Body   Item   `json:"body"`
}

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

	slog.Info(fmt.Sprintf("%s\n", output))
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
	fs.Usage = flags.Usage(fs)

	loggerConfig := logger.Flags(fs, "logger")

	app := flags.New("app", "Application").DocPrefix("algolia").String(fs, "", nil)
	key := flags.New("key", "Key").DocPrefix("algolia").String(fs, "", nil)
	index := flags.New("index", "Index").DocPrefix("algolia").String(fs, "", nil)
	source := flags.New("source", "Walked markdown directory").DocPrefix("reveal").String(fs, "", nil)
	prefixFromFolder := flags.New("prefixFromFolder", "Use name of folder as URL prefix").DocPrefix("reveal").Bool(fs, false, nil)
	sep := flags.New("sep", "Separator").DocPrefix("reveal").String(fs, "^\n\n\n", nil)
	vsep := flags.New("verticalSep", "Vertical separator").DocPrefix("reveal").String(fs, "^\n\n", nil)

	debug := flags.New("debug", "Debug output instead of sending them").DocPrefix("app").Bool(fs, false, nil)

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	logger.Init(loggerConfig)

	ctx := context.Background()

	sepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *sep))
	vsepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *vsep))
	req := getRequest(*app, *key)

	if !*debug {
		if err := clearIndex(ctx, req, *app, *index); err != nil {
			slog.Error("clear index", "err", err)
			os.Exit(1)
		}

		if err := configIndex(ctx, req, *app, *index); err != nil {
			slog.Error("config index", "err", err)
			os.Exit(1)
		}
	}

	err := filepath.Walk(*source, func(path string, info os.FileInfo, _ error) error {
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

		slog.Info("Objects found", "count", len(objects), "name", info.Name())
		if *debug {
			if err := debugObjects(objects); err != nil {
				return err
			}
		} else if err := saveObjects(ctx, req, *app, *index, objects); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		slog.Error("walk source", "err", err)
		os.Exit(1)
	}

	slog.Info("Done!")
}
