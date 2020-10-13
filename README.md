# @prantlf/esbuild

_An extremely fast JavaScript bundler_

[Website](https://esbuild.github.io/) | [Getting started](https://esbuild.github.io/getting-started/) | [Documentation](https://esbuild.github.io/api/) | [FAQ](https://esbuild.github.io/faq/)

## Why?

Our current build tools for the web are 10-100x slower than they could be:

![](images/benchmark.png)

The main goal of the esbuild bundler project is to bring about a new era of build tool performance, and create an easy-to-use modern bundler along the way.

Major features:

* Extreme speed without needing a cache
* ES6 and CommonJS modules
* Tree shaking of ES6 modules
* An [API](https://esbuild.github.io/api/) for JavaScript and Go
* [TypeScript](https://esbuild.github.io/content-types/#typescript) and [JSX](https://esbuild.github.io/content-types/#jsx) syntax
* [Source maps](https://esbuild.github.io/api/#sourcemap)
* [Minification](https://esbuild.github.io/api/#minify)

Check out the [getting started](https://esbuild.github.io/getting-started/) instructions if you want to give esbuild a try.

## About This Fork

This is a fork of the original project that demonstrates experimental features:

1. Installation by `go get`. ([enable-go-get](https://github.com/prantlf/esbuild/commits/enable-go-get))
2. Analysis of module dependencies only, without compiling the output bundle. ([analyse](https://github.com/prantlf/esbuild/commits/analyse))
3. Support for [AMD.JS](https://github.com/amdjs/amdjs-api/wiki/AMD) (WIP). ([amdjs](https://github.com/prantlf/esbuild/commits/amdjs))

How to make it installable in `Go` projects by `go get`. For example:

    GO111MODULE=off go get -u github.com/prantlf/esbuild/...@308fdf459d65

**UPDATE:** This change has been made in the [original repository](https://github.com/evanw/esbuild#esbuild) ([764109e](https://github.com/evanw/esbuild/commit/764109effefbd2302e0f2df301b82dce6ef8024e), [0d4c970](https://github.com/evanw/esbuild/commit/0d4c9705da29859f7587ab9ce3c3314c8efaa978), [a94685d](https://github.com/evanw/esbuild/commit/a94685dd3afdbb2c92de5fee1a803123db1d9222)). Using the original project:

    GO111MODULE=off go get -u github.com/evanw/esbuild/...

How to perform the source analysis only, command line, Go and JavaScript:

    esbuild --analyse entry.js --metafile=dependencies.json

    result := api.Analyse(api.AnalyseOptions{
      EntryPoints: []string{"entry.js"},
      Metafile:    "dependencies.json",
      LogLevel:    api.LogLevelInfo,
    })

    const result = await service.analyse({
      entryPoints: ['entry.js],
      metafile: 'dependencies.json',
      write: false
    });

How to build an AMD.JS project, command line and `config.json`:

    esbuild --bundle --amdconfig=config.json main.js --outfile=out/bundle.js

    {
      "baseUrl": "src",
      "paths": {
        "external": "empty:",
        "internal": "."
      },
      "map": {
        "*": {
          "external/libs/jquery": "internal/libs/jquery"
        }
      },
      "plugins": {
        "json": {
          "fileExtensions": [".json"]
        },
        "css": {
          "fileExtensions": [".css"],
          "appendFileExtension": true
        },
        "i18n": {
          "fileExtensions": [".js"],
          "loadScript": {
            "replacementPattern": "/nls/",
            "replacementValue": "/nls/root/"
          }
        }
      }
    }