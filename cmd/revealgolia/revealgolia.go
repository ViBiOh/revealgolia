package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"

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

func main() {
	config := newConfig()

	ctx := context.Background()

	newClients(ctx, config)

	sepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *config.sep))
	vsepRegex := regexp.MustCompile(fmt.Sprintf("(?m)%s", *config.vsep))
	req := getRequest(*config.app, *config.key)

	if !*config.debug {
		err := clearIndex(ctx, req, *config.app, *config.index)
		logger.FatalfOnErr(ctx, err, "clear index", slog.String("index", *config.index))

		err = configIndex(ctx, req, *config.app, *config.index)
		logger.FatalfOnErr(ctx, err, "config index", slog.String("index", *config.index))
	}

	err := filepath.Walk(*config.source, func(path string, info os.FileInfo, _ error) error {
		return processFile(ctx, path, config, sepRegex, vsepRegex, req)
	})
	logger.FatalfOnErr(ctx, err, "walk source")

	slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("Index `%s` refreshed!", *config.index), slog.String("path", *config.source))
}

func processFile(ctx context.Context, path string, config configuration, sepRegex, vsepRegex *regexp.Regexp, req request.Request) error {
	if filepath.Ext(path) != ".md" {
		return nil
	}

	var name string
	if *config.prefixFromFolder {
		name = filepath.Base(filepath.Dir(path))
	}

	objects, err := getSearchObjects(name, path, sepRegex, vsepRegex)
	if err != nil {
		return err
	}

	slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("%d objects found in `%s`", len(objects), path), slog.String("index", *config.index))
	if *config.debug {
		if err := debugObjects(ctx, objects); err != nil {
			return err
		}
	} else if err := saveObjects(ctx, req, *config.app, *config.index, objects); err != nil {
		return err
	}

	return nil
}

func getRequest(app, key string) request.Request {
	return request.New().Header("X-Algolia-Application-Id", app).Header("X-Algolia-API-Key", key)
}

func clearIndex(ctx context.Context, request request.Request, app, index string) error {
	_, err := request.Post(getURL(app, "/1/indexes/%s/clear", index)).Send(ctx, nil)
	return err
}

func getURL(app, path, index string) string {
	return fmt.Sprintf(fmt.Sprintf("https://%s.algolia.net%s", app, path), index)
}

func configIndex(ctx context.Context, request request.Request, app, index string) error {
	settings := map[string]any{
		"searchableAttributes": []string{"keywords", "img", "content"},
	}

	_, err := request.Put(getURL(app, "/1/indexes/%s/settings", index)).JSON(ctx, settings)
	return err
}

func getSearchObjects(name, source string, sep, verticalSep *regexp.Regexp) ([]Item, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}

	var objects []Item
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

			keywords = keywords[:0]
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

func debugObjects(ctx context.Context, objects []Item) error {
	output, err := json.MarshalIndent(objects, "", "  ")
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("%s\n", output))
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
