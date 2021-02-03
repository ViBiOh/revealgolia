# revealgolia

[![Build](https://github.com/ViBiOh/revealgolia/workflows/Build/badge.svg)](https://github.com/ViBiOh/revealgolia/actions)
[![codecov](https://codecov.io/gh/ViBiOh/revealgolia/branch/main/graph/badge.svg)](https://codecov.io/gh/ViBiOh/revealgolia)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/revealgolia)](https://goreportcard.com/report/github.com/ViBiOh/revealgolia)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ViBiOh_revealgolia&metric=alert_status)](https://sonarcloud.io/dashboard?id=ViBiOh_revealgolia)

## Usage

```bash
Usage of revealgolia:
  -app string
        [algolia] Application {REVEALGOLIA_APP}
  -debug
        [app] Debug output instead of sending them {REVEALGOLIA_DEBUG}
  -index string
        [algolia] Index {REVEALGOLIA_INDEX}
  -key string
        [algolia] Key {REVEALGOLIA_KEY}
  -loggerJson
        [logger] Log format as JSON {REVEALGOLIA_LOGGER_JSON}
  -loggerLevel string
        [logger] Logger level {REVEALGOLIA_LOGGER_LEVEL} (default "INFO")
  -loggerLevelKey string
        [logger] Key for level in JSON {REVEALGOLIA_LOGGER_LEVEL_KEY} (default "level")
  -loggerMessageKey string
        [logger] Key for message in JSON {REVEALGOLIA_LOGGER_MESSAGE_KEY} (default "message")
  -loggerTimeKey string
        [logger] Key for timestamp in JSON {REVEALGOLIA_LOGGER_TIME_KEY} (default "time")
  -prefixFromFolder
        [reveal] Use name of folder as URL prefix {REVEALGOLIA_PREFIX_FROM_FOLDER}
  -sep string
        [reveal] Separator {REVEALGOLIA_SEP} (default "^\n\n\n")
  -source string
        [reveal] Walked markdown directory {REVEALGOLIA_SOURCE}
  -verticalSep string
        [reveal] Vertical separator {REVEALGOLIA_VERTICAL_SEP} (default "^\n\n")
```
