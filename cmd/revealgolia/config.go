package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
)

type configuration struct {
	logger *logger.Config

	app              *string
	key              *string
	index            *string
	source           *string
	prefixFromFolder *bool
	sep              *string
	vsep             *string

	debug *bool
}

func newConfig() configuration {
	fs := flag.NewFlagSet("revealgolia", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger: logger.Flags(fs, "logger"),

		app:              flags.New("app", "Application").DocPrefix("algolia").String(fs, "", nil),
		key:              flags.New("key", "Key").DocPrefix("algolia").String(fs, "", nil),
		index:            flags.New("index", "Index").DocPrefix("algolia").String(fs, "", nil),
		source:           flags.New("source", "Walked markdown directory").DocPrefix("reveal").String(fs, "", nil),
		prefixFromFolder: flags.New("prefixFromFolder", "Use name of folder as URL prefix").DocPrefix("reveal").Bool(fs, false, nil),
		sep:              flags.New("sep", "Separator").DocPrefix("reveal").String(fs, "^\n\n\n", nil),
		vsep:             flags.New("verticalSep", "Vertical separator").DocPrefix("reveal").String(fs, "^\n\n", nil),

		debug: flags.New("debug", "Debug output instead of sending them").DocPrefix("app").Bool(fs, false, nil),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
