# revealgolia

[![Build](https://github.com/ViBiOh/revealgolia/workflows/Build/badge.svg)](https://github.com/ViBiOh/revealgolia/actions)

## Usage

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of revealgolia:
  --app               string  [algolia] Application ${REVEALGOLIA_APP}
  --debug                     [app] Debug output instead of sending them ${REVEALGOLIA_DEBUG} (default false)
  --index             string  [algolia] Index ${REVEALGOLIA_INDEX}
  --key               string  [algolia] Key ${REVEALGOLIA_KEY}
  --loggerJson                [logger] Log format as JSON ${REVEALGOLIA_LOGGER_JSON} (default false)
  --loggerLevel       string  [logger] Logger level ${REVEALGOLIA_LOGGER_LEVEL} (default "INFO")
  --loggerLevelKey    string  [logger] Key for level in JSON ${REVEALGOLIA_LOGGER_LEVEL_KEY} (default "level")
  --loggerMessageKey  string  [logger] Key for message in JSON ${REVEALGOLIA_LOGGER_MESSAGE_KEY} (default "msg")
  --loggerTimeKey     string  [logger] Key for timestamp in JSON ${REVEALGOLIA_LOGGER_TIME_KEY} (default "time")
  --prefixFromFolder          [reveal] Use name of folder as URL prefix ${REVEALGOLIA_PREFIX_FROM_FOLDER} (default false)
  --sep               string  [reveal] Separator ${REVEALGOLIA_SEP} (default "^\n\n\n")
  --source            string  [reveal] Walked markdown directory ${REVEALGOLIA_SOURCE}
  --verticalSep       string  [reveal] Vertical separator ${REVEALGOLIA_VERTICAL_SEP} (default "^\n\n")
```
