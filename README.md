# @prantlf/esbuild

_An extremely fast JavaScript bundler (see below [about this fork](#about-this-fork))_

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
* [Plugins](https://esbuild.github.io/plugins/)

Check out the [getting started](https://esbuild.github.io/getting-started/) instructions if you want to give esbuild a try.

## About This Fork

This is a fork of the original project that demonstrates experimental features:

1. Analysis of module dependencies only, without compiling the output bundle. (see the branch [analyse](https://github.com/prantlf/esbuild/commits/analyse))
2. Support for [AMD](https://github.com/amdjs/amdjs-api/wiki/AMD) input (WIP). (see the branch [amdjs](https://github.com/prantlf/esbuild/commits/amdjs))
3. Support for [UMD](https://github.com/umdjs/umd) output. (see the branch [umdjs](https://github.com/prantlf/esbuild/commits/umdjs))

### Installation

Install the `esbuild` binary globally by either of:

    npm i -g @prantlf/esbuild
    pnpm i -g @prantlf/esbuild
    yarn global add @prantlf/esbuild

### Analysis

How to perform the source analysis only on the command line:

    esbuild --analyse entry.js --metafile=dependencies.json

in Go:

```go
result := api.Analyse(api.AnalyseOptions{
  EntryPoints: []string{"entry.js"},
  Metafile:    "dependencies.json",
  LogLevel:    api.LogLevelInfo,
})
```

and in JavaScript:

```js
const result = await service.analyse({
  entryPoints: ['entry.js'],
  metafile: 'dependencies.json',
  write: false
});
```

### AMD

How to build an AMD project on the command line:

    esbuild --bundle --amdconfig=config.json main.js --outfile=out/bundle.js

with the following `config.json`:

```json
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
```

### UMD

How to build an UMD library on the command line:

    esbuild --bundle --format=umd --global-name=mylib index.js --outfile=bundle.js
