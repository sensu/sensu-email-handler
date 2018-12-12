# Sensu Go CHANGEME Plugin
TravisCI: [![TravisCI Build Status](https://travis-ci.org/CHANGEME/sensu-CHANGEME.svg?branch=master)](https://travis-ci.org/CHANGEME/sensu-CHANGEME)

TODO: Description.

## Installation

Download the latest version of the sensu-CHANGEME from [releases][1],
or create an executable script from this source.

From the local path of the sensu-CHANGEME repository:

```
go build -o /usr/local/bin/sensu-CHANGEME main.go
```

## Configuration

Example Sensu Go definition:

```json
{
    "api_version": "core/v2",
    "type": "CHANGEME",
    "metadata": {
        "namespace": "default",
        "name": "CHANGEME"
    },
    "spec": {
        "...": "..."
    }
}
```

## Usage Examples

Help:

```
The Sensu Go CHANGEME for x

Usage:
  sensu-CHANGEME [flags]

Flags:
  -f, --foo string   example
  -h, --help         help for sensu-CHANGEME
```

## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/CHANGEME/sensu-CHANGEME/releases
