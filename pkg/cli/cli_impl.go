package cli

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/evanw/esbuild/internal/cli_helpers"
	"github.com/evanw/esbuild/internal/fs"
	"github.com/evanw/esbuild/internal/logger"
	"github.com/evanw/esbuild/pkg/api"
)

func newBuildOptions() api.BuildOptions {
	return api.BuildOptions{
		Loader: make(map[string]api.Loader),
		Define: make(map[string]string),
	}
}

func newTransformOptions() api.TransformOptions {
	return api.TransformOptions{
		Define: make(map[string]string),
	}
}

func newAnalyseOptions() api.AnalyseOptions {
	return api.AnalyseOptions{
		Loader: make(map[string]api.Loader),
		Define: make(map[string]string),
	}
}

func parseOptionsImpl(osArgs []string, buildOpts *api.BuildOptions, transformOpts *api.TransformOptions, analyseOpts *api.AnalyseOptions) error {
	hasBareSourceMapFlag := false
	analyse := false

	// Parse the arguments now that we know what we're parsing
	for _, arg := range osArgs {
		switch {
		case arg == "--bundle" && buildOpts != nil:
			buildOpts.Bundle = true

		case arg == "--analyse" && analyseOpts != nil:
			analyseOpts.Bundle = true
			analyse = true

		case arg == "--preserve-symlinks" && buildOpts != nil:
			buildOpts.PreserveSymlinks = true

		case arg == "--splitting":
			if buildOpts != nil {
				buildOpts.Splitting = true
			} else {
				analyseOpts.Splitting = true
			}

		case arg == "--watch" && buildOpts != nil:
			buildOpts.Watch = &api.WatchMode{}

		case arg == "--minify":
			if buildOpts != nil {
				buildOpts.MinifySyntax = true
				buildOpts.MinifyWhitespace = true
				buildOpts.MinifyIdentifiers = true
			} else if transformOpts != nil {
				transformOpts.MinifySyntax = true
				transformOpts.MinifyWhitespace = true
				transformOpts.MinifyIdentifiers = true
			}

		case arg == "--minify-syntax":
			if buildOpts != nil {
				buildOpts.MinifySyntax = true
			} else if transformOpts != nil {
				transformOpts.MinifySyntax = true
			}

		case arg == "--minify-whitespace":
			if buildOpts != nil {
				buildOpts.MinifyWhitespace = true
			} else if transformOpts != nil {
				transformOpts.MinifyWhitespace = true
			}

		case arg == "--minify-identifiers":
			if buildOpts != nil {
				buildOpts.MinifyIdentifiers = true
			} else if transformOpts != nil {
				transformOpts.MinifyIdentifiers = true
			}

		case strings.HasPrefix(arg, "--charset="):
			var value *api.Charset
			if buildOpts != nil {
				value = &buildOpts.Charset
			} else {
				value = &transformOpts.Charset
			}
			name := arg[len("--charset="):]
			switch name {
			case "ascii":
				*value = api.CharsetASCII
			case "utf8":
				*value = api.CharsetUTF8
			default:
				return fmt.Errorf("Invalid charset value: %q (valid: ascii, utf8)", name)
			}

		case strings.HasPrefix(arg, "--tree-shaking="):
			var value *api.TreeShaking
			if buildOpts != nil {
				value = &buildOpts.TreeShaking
			} else {
				value = &transformOpts.TreeShaking
			}
			name := arg[len("--tree-shaking="):]
			switch name {
			case "ignore-annotations":
				*value = api.TreeShakingIgnoreAnnotations
			default:
				return fmt.Errorf("Invalid tree shaking value: %q (valid: ignore-annotations)", name)
			}

		case arg == "--avoid-tdz":
			if buildOpts != nil {
				buildOpts.AvoidTDZ = true
			} else {
				transformOpts.AvoidTDZ = true
			}

		case arg == "--keep-names":
			if buildOpts != nil {
				buildOpts.KeepNames = true
			} else {
				transformOpts.KeepNames = true
			}

		case arg == "--sourcemap":
			if buildOpts != nil {
				buildOpts.Sourcemap = api.SourceMapLinked
			} else if transformOpts != nil {
				transformOpts.Sourcemap = api.SourceMapInline
			}
			hasBareSourceMapFlag = true

		case strings.HasPrefix(arg, "--sourcemap="):
			value := arg[len("--sourcemap="):]
			var sourcemap api.SourceMap
			switch value {
			case "inline":
				sourcemap = api.SourceMapInline
			case "external":
				sourcemap = api.SourceMapExternal
			case "both":
				sourcemap = api.SourceMapInlineAndExternal
			default:
				return fmt.Errorf("Invalid sourcemap: %q (valid: inline, external, both)", value)
			}
			if buildOpts != nil {
				buildOpts.Sourcemap = sourcemap
			} else {
				transformOpts.Sourcemap = sourcemap
			}
			hasBareSourceMapFlag = false

		case strings.HasPrefix(arg, "--sources-content="):
			value := arg[len("--sources-content="):]
			var sourcesContent api.SourcesContent
			switch value {
			case "false":
				sourcesContent = api.SourcesContentExclude
			case "true":
				sourcesContent = api.SourcesContentInclude
			default:
				return fmt.Errorf("Invalid sources content: %q (valid: false, true)", value)
			}
			if buildOpts != nil {
				buildOpts.SourcesContent = sourcesContent
			} else {
				transformOpts.SourcesContent = sourcesContent
			}

		case strings.HasPrefix(arg, "--sourcefile="):
			if buildOpts != nil {
				if buildOpts.Stdin == nil {
					buildOpts.Stdin = &api.StdinOptions{}
				}
				buildOpts.Stdin.Sourcefile = arg[len("--sourcefile="):]
			} else if transformOpts != nil {
				transformOpts.Sourcefile = arg[len("--sourcefile="):]
			} else {
				if analyseOpts.Stdin == nil {
					analyseOpts.Stdin = &api.StdinOptions{}
				}
				analyseOpts.Stdin.Sourcefile = arg[len("--sourcefile="):]
			}

		case strings.HasPrefix(arg, "--resolve-extensions="):
			if buildOpts != nil {
				buildOpts.ResolveExtensions = strings.Split(arg[len("--resolve-extensions="):], ",")
			} else {
				analyseOpts.ResolveExtensions = strings.Split(arg[len("--resolve-extensions="):], ",")
			}

		case strings.HasPrefix(arg, "--main-fields="):
			if buildOpts != nil {
				buildOpts.MainFields = strings.Split(arg[len("--main-fields="):], ",")
			} else {
				analyseOpts.MainFields = strings.Split(arg[len("--main-fields="):], ",")
			}

		case strings.HasPrefix(arg, "--public-path=") && buildOpts != nil:
			buildOpts.PublicPath = arg[len("--public-path="):]

		case strings.HasPrefix(arg, "--global-name="):
			if buildOpts != nil {
				buildOpts.GlobalName = arg[len("--global-name="):]
			} else if transformOpts != nil {
				transformOpts.GlobalName = arg[len("--global-name="):]
			} else {
				analyseOpts.GlobalName = arg[len("--global-name="):]
			}

		case strings.HasPrefix(arg, "--metafile="):
			if buildOpts != nil {
				buildOpts.Metafile = arg[len("--metafile="):]
			} else {
				analyseOpts.Metafile = arg[len("--metafile="):]
			}

		case strings.HasPrefix(arg, "--outfile=") && buildOpts != nil:
			buildOpts.Outfile = arg[len("--outfile="):]

		case strings.HasPrefix(arg, "--outdir=") && buildOpts != nil:
			buildOpts.Outdir = arg[len("--outdir="):]

		case strings.HasPrefix(arg, "--outbase=") && buildOpts != nil:
			buildOpts.Outbase = arg[len("--outbase="):]

		case strings.HasPrefix(arg, "--tsconfig="):
			if buildOpts != nil {
				buildOpts.Tsconfig = arg[len("--tsconfig="):]
			} else {
				analyseOpts.Tsconfig = arg[len("--tsconfig="):]
			}

		case strings.HasPrefix(arg, "--amdconfig="):
			if buildOpts != nil {
				buildOpts.AMDConfig = arg[len("--amdconfig="):]
			} else {
				analyseOpts.AMDConfig = arg[len("--amdconfig="):]
			}

		case strings.HasPrefix(arg, "--tsconfig-raw=") && transformOpts != nil:
			transformOpts.TsconfigRaw = arg[len("--tsconfig-raw="):]

		case strings.HasPrefix(arg, "--define:"):
			value := arg[len("--define:"):]
			equals := strings.IndexByte(value, '=')
			if equals == -1 {
				return fmt.Errorf("Missing \"=\": %q", value)
			}
			if buildOpts != nil {
				buildOpts.Define[value[:equals]] = value[equals+1:]
			} else if transformOpts != nil {
				transformOpts.Define[value[:equals]] = value[equals+1:]
			} else {
				analyseOpts.Define[value[:equals]] = value[equals+1:]
			}

		case strings.HasPrefix(arg, "--pure:"):
			value := arg[len("--pure:"):]
			if buildOpts != nil {
				buildOpts.Pure = append(buildOpts.Pure, value)
			} else if transformOpts != nil {
				transformOpts.Pure = append(transformOpts.Pure, value)
			} else {
				analyseOpts.Pure = append(analyseOpts.Pure, value)
			}

		case strings.HasPrefix(arg, "--loader:"):
			value := arg[len("--loader:"):]
			equals := strings.IndexByte(value, '=')
			if equals == -1 {
				return fmt.Errorf("Missing \"=\": %q", value)
			}
			ext, text := value[:equals], value[equals+1:]
			loader, err := cli_helpers.ParseLoader(text)
			if err != nil {
				return err
			}
			if buildOpts != nil {
				buildOpts.Loader[ext] = loader
			} else {
				analyseOpts.Loader[ext] = loader
			}

		case strings.HasPrefix(arg, "--loader="):
			value := arg[len("--loader="):]
			loader, err := cli_helpers.ParseLoader(value)
			if err != nil {
				return err
			}
			if loader == api.LoaderFile {
				return fmt.Errorf("Cannot transform using the \"file\" loader")
			}
			if buildOpts != nil {
				if buildOpts.Stdin == nil {
					buildOpts.Stdin = &api.StdinOptions{}
				}
				buildOpts.Stdin.Loader = loader
			} else if transformOpts != nil {
				transformOpts.Loader = loader
			} else {
				if analyseOpts.Stdin == nil {
					analyseOpts.Stdin = &api.StdinOptions{}
				}
				analyseOpts.Stdin.Loader = loader
			}

		case strings.HasPrefix(arg, "--target="):
			target, engines, err := parseTargets(strings.Split(arg[len("--target="):], ","))
			if err != nil {
				return err
			}
			if buildOpts != nil {
				buildOpts.Target = target
				buildOpts.Engines = engines
			} else if transformOpts != nil {
				transformOpts.Target = target
				transformOpts.Engines = engines
			} else {
				analyseOpts.Target = target
				analyseOpts.Engines = engines
			}

		case strings.HasPrefix(arg, "--out-extension:") && buildOpts != nil:
			value := arg[len("--out-extension:"):]
			equals := strings.IndexByte(value, '=')
			if equals == -1 {
				return fmt.Errorf("Missing \"=\": %q", value)
			}
			if buildOpts.OutExtensions == nil {
				buildOpts.OutExtensions = make(map[string]string)
			}
			buildOpts.OutExtensions[value[:equals]] = value[equals+1:]

		case strings.HasPrefix(arg, "--platform="):
			value := arg[len("--platform="):]
			switch value {
			case "browser":
				if buildOpts != nil {
					buildOpts.Platform = api.PlatformBrowser
				} else {
					analyseOpts.Platform = api.PlatformBrowser
				}
			case "node":
				if buildOpts != nil {
					buildOpts.Platform = api.PlatformNode
				} else {
					analyseOpts.Platform = api.PlatformNode
				}
			case "neutral":
				if buildOpts != nil {
					buildOpts.Platform = api.PlatformNeutral
				} else {
					analyseOpts.Platform = api.PlatformNeutral
				}
			default:
				return fmt.Errorf("Invalid platform: %q (valid: browser, node, neutral)", value)
			}

		case strings.HasPrefix(arg, "--format="):
			value := arg[len("--format="):]
			switch value {
			case "iife":
				if buildOpts != nil {
					buildOpts.Format = api.FormatIIFE
				} else {
					transformOpts.Format = api.FormatIIFE
				}
			case "cjs":
				if buildOpts != nil {
					buildOpts.Format = api.FormatCommonJS
				} else {
					transformOpts.Format = api.FormatCommonJS
				}
			case "umd":
				if buildOpts != nil {
					buildOpts.Format = api.FormatUMD
				} else {
					transformOpts.Format = api.FormatUMD
				}
			case "esm":
				if buildOpts != nil {
					buildOpts.Format = api.FormatESModule
				} else {
					transformOpts.Format = api.FormatESModule
				}
			default:
				return fmt.Errorf("Invalid format: %q (valid: iife, cjs, umd, esm)", value)
			}

		case strings.HasPrefix(arg, "--external:"):
			if buildOpts != nil {
				buildOpts.External = append(buildOpts.External, arg[len("--external:"):])
			} else {
				analyseOpts.External = append(analyseOpts.External, arg[len("--external:"):])
			}

		case strings.HasPrefix(arg, "--inject:") && buildOpts != nil:
			buildOpts.Inject = append(buildOpts.Inject, arg[len("--inject:"):])

		case strings.HasPrefix(arg, "--jsx-factory="):
			value := arg[len("--jsx-factory="):]
			if buildOpts != nil {
				buildOpts.JSXFactory = value
			} else if transformOpts != nil {
				transformOpts.JSXFactory = value
			} else {
				analyseOpts.JSXFactory = value
			}

		case strings.HasPrefix(arg, "--jsx-fragment="):
			value := arg[len("--jsx-fragment="):]
			if buildOpts != nil {
				buildOpts.JSXFragment = value
			} else if transformOpts != nil {
				transformOpts.JSXFragment = value
			} else {
				analyseOpts.JSXFragment = value
			}

		case strings.HasPrefix(arg, "--banner="):
			value := arg[len("--banner="):]
			if buildOpts != nil {
				buildOpts.Banner = value
			} else {
				transformOpts.Banner = value
			}

		case strings.HasPrefix(arg, "--footer="):
			value := arg[len("--footer="):]
			if buildOpts != nil {
				buildOpts.Footer = value
			} else {
				transformOpts.Footer = value
			}

		case strings.HasPrefix(arg, "--error-limit="):
			value := arg[len("--error-limit="):]
			limit, err := strconv.Atoi(value)
			if err != nil || limit < 0 {
				return fmt.Errorf("Invalid error limit: %q", value)
			}
			if buildOpts != nil {
				buildOpts.ErrorLimit = limit
			} else if transformOpts != nil {
				transformOpts.ErrorLimit = limit
			} else {
				analyseOpts.ErrorLimit = limit
			}

			// Make sure this stays in sync with "PrintErrorToStderr"
		case strings.HasPrefix(arg, "--color="):
			value := arg[len("--color="):]
			var color api.StderrColor
			switch value {
			case "false":
				color = api.ColorNever
			case "true":
				color = api.ColorAlways
			default:
				return fmt.Errorf("Invalid color: %q (valid: false, true)", value)
			}
			if buildOpts != nil {
				buildOpts.Color = color
			} else if transformOpts != nil {
				transformOpts.Color = color
			} else {
				analyseOpts.Color = color
			}

		// Make sure this stays in sync with "PrintErrorToStderr"
		case strings.HasPrefix(arg, "--log-level="):
			value := arg[len("--log-level="):]
			var logLevel api.LogLevel
			switch value {
			case "info":
				logLevel = api.LogLevelInfo
			case "warning":
				logLevel = api.LogLevelWarning
			case "error":
				logLevel = api.LogLevelError
			case "silent":
				logLevel = api.LogLevelSilent
			default:
				return fmt.Errorf("Invalid log level: %q (valid: info, warning, error, silent)", arg)
			}
			if buildOpts != nil {
				buildOpts.LogLevel = logLevel
			} else if transformOpts != nil {
				transformOpts.LogLevel = logLevel
			} else {
				analyseOpts.LogLevel = logLevel
			}

		case strings.HasPrefix(arg, "'--"):
			return fmt.Errorf("Unexpected single quote character before flag (use \\\" to escape double quotes): %s", arg)

		case !strings.HasPrefix(arg, "-"):
			if buildOpts != nil {
				buildOpts.EntryPoints = append(buildOpts.EntryPoints, arg)
			} else {
				analyseOpts.EntryPoints = append(analyseOpts.EntryPoints, arg)
			}
		case !strings.HasPrefix(arg, "-") && buildOpts != nil:
			buildOpts.EntryPoints = append(buildOpts.EntryPoints, arg)

		default:
			if buildOpts != nil {
				return fmt.Errorf("Invalid build flag: %q", arg)
			} else if transformOpts != nil {
				return fmt.Errorf("Invalid transform flag: %q", arg)
			} else {
				return fmt.Errorf("Invalid analyse flag: %q", arg)
			}
		}
	}

	// If we're building, the last source map flag is "--sourcemap", and there
	// is no output path, change the source map option to "inline" because we're
	// going to be writing to stdout which can only represent a single file.
	if buildOpts != nil && hasBareSourceMapFlag && buildOpts.Outfile == "" && buildOpts.Outdir == "" {
		buildOpts.Sourcemap = api.SourceMapInline
	}

	if analyseOpts != nil && !analyse {
		return fmt.Errorf("Missing --analyse flag")
	}

	return nil
}

func parseTargets(targets []string) (target api.Target, engines []api.Engine, err error) {
	validTargets := map[string]api.Target{
		"esnext": api.ESNext,
		"es5":    api.ES5,
		"es6":    api.ES2015,
		"es2015": api.ES2015,
		"es2016": api.ES2016,
		"es2017": api.ES2017,
		"es2018": api.ES2018,
		"es2019": api.ES2019,
		"es2020": api.ES2020,
	}

	validEngines := map[string]api.EngineName{
		"chrome":  api.EngineChrome,
		"firefox": api.EngineFirefox,
		"safari":  api.EngineSafari,
		"edge":    api.EngineEdge,
		"node":    api.EngineNode,
		"ios":     api.EngineIOS,
	}

outer:
	for _, value := range targets {
		if valid, ok := validTargets[value]; ok {
			target = valid
			continue
		}

		for engine, name := range validEngines {
			if strings.HasPrefix(value, engine) {
				version := value[len(engine):]
				if version == "" {
					return 0, nil, fmt.Errorf("Target missing version number: %q", value)
				}
				engines = append(engines, api.Engine{Name: name, Version: version})
				continue outer
			}
		}

		var engines []string
		for key := range validEngines {
			engines = append(engines, key+"N")
		}
		sort.Strings(engines)
		return 0, nil, fmt.Errorf(
			"Invalid target: %q (valid: esN, "+strings.Join(engines, ", ")+")", value)
	}
	return
}

// This returns either BuildOptions, TransformOptions, or an error
func parseOptionsForRun(osArgs []string) (*api.BuildOptions, *api.TransformOptions, *api.AnalyseOptions, error) {
	// If there's an entry point or we're bundling, then we're building
	// If there's the --analyse flag set, then we're analysing
	for _, arg := range osArgs {
		if !strings.HasPrefix(arg, "-") || arg == "--bundle" {
			options := newBuildOptions()

			// Apply defaults appropriate for the CLI
			options.ErrorLimit = 10
			options.LogLevel = api.LogLevelInfo
			options.Write = true

			err := parseOptionsImpl(osArgs, &options, nil, nil)
			if err != nil {
				return nil, nil, nil, err
			}
			return &options, nil, nil, nil
		} else if !strings.HasPrefix(arg, "-") || arg == "--analyse" {
			options := newAnalyseOptions()

			// Apply defaults appropriate for the CLI
			options.ErrorLimit = 10
			options.LogLevel = api.LogLevelInfo
			options.Write = true

			err := parseOptionsImpl(osArgs, nil, nil, &options)
			if err != nil {
				return nil, nil, nil, err
			}
			return nil, nil, &options, nil
		}
	}

	// Otherwise, we're transforming
	options := newTransformOptions()

	// Apply defaults appropriate for the CLI
	options.ErrorLimit = 10
	options.LogLevel = api.LogLevelInfo

	err := parseOptionsImpl(osArgs, nil, &options, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	if options.Sourcemap != api.SourceMapNone && options.Sourcemap != api.SourceMapInline {
		return nil, nil, nil, fmt.Errorf("Must use \"inline\" source map when transforming stdin")
	}
	return nil, &options, nil, nil
}

func runImpl(osArgs []string) int {
	shouldPrintSummary := false
	start := time.Now()
	end := 0

	for _, arg := range osArgs {
		// Special-case running a server
		if arg == "--serve" || strings.HasPrefix(arg, "--serve=") || strings.HasPrefix(arg, "--servedir=") {
			if err := serveImpl(osArgs); err != nil {
				logger.PrintErrorToStderr(osArgs, err.Error())
				return 1
			}
			return 0
		}

		// Filter out the "--summary" flag
		if arg == "--summary" {
			shouldPrintSummary = true
			continue
		}

		osArgs[end] = arg
		end++
	}
	osArgs = osArgs[:end]

	buildOptions, transformOptions, analyseOptions, err := parseOptionsForRun(osArgs)

	switch {
	case buildOptions != nil:
		// Read the "NODE_PATH" from the environment. This is part of node's
		// module resolution algorithm. Documentation for this can be found here:
		// https://nodejs.org/api/modules.html#modules_loading_from_the_global_folders
		for _, key := range os.Environ() {
			if strings.HasPrefix(key, "NODE_PATH=") {
				value := key[len("NODE_PATH="):]
				separator := ":"
				if fs.CheckIfWindows() {
					// On Windows, NODE_PATH is delimited by semicolons instead of colons
					separator = ";"
				}
				buildOptions.NodePaths = strings.Split(value, separator)
				break
			}
		}

		// Read from stdin when there are no entry points
		if len(buildOptions.EntryPoints) == 0 {
			if buildOptions.Stdin == nil {
				buildOptions.Stdin = &api.StdinOptions{}
			}
			bytes, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				logger.PrintErrorToStderr(osArgs, fmt.Sprintf(
					"Could not read from stdin: %s", err.Error()))
				return 1
			}
			buildOptions.Stdin.Contents = string(bytes)
			buildOptions.Stdin.ResolveDir, _ = os.Getwd()
		} else if buildOptions.Stdin != nil {
			if buildOptions.Stdin.Sourcefile != "" {
				logger.PrintErrorToStderr(osArgs,
					"\"sourcefile\" only applies when reading from stdin")
			} else {
				logger.PrintErrorToStderr(osArgs,
					"\"loader\" without extension only applies when reading from stdin")
			}
			return 1
		}

		// Run the build
		result := api.Build(*buildOptions)

		// Do not exit if we're in watch mode
		if buildOptions.Watch != nil {
			<-make(chan bool)
		}

		// Stop if there were errors
		if len(result.Errors) > 0 {
			return 1
		}

		// Print a summary to stderr
		if shouldPrintSummary {
			printSummary(osArgs, result.OutputFiles, start)
		}

	case transformOptions != nil:
		// Read the input from stdin
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			logger.PrintErrorToStderr(osArgs, fmt.Sprintf(
				"Could not read from stdin: %s", err.Error()))
			return 1
		}

		// Run the transform and stop if there were errors
		result := api.Transform(string(bytes), *transformOptions)
		if len(result.Errors) > 0 {
			return 1
		}

		// Write the output to stdout
		os.Stdout.Write(result.Code)

		// Print a summary to stderr
		if shouldPrintSummary {
			printSummary(osArgs, nil, start)
		}

	case analyseOptions != nil:
		// Read the "NODE_PATH" from the environment. This is part of node's
		// module resolution algorithm. Documentation for this can be found here:
		// https://nodejs.org/api/modules.html#modules_loading_from_the_global_folders
		for _, key := range os.Environ() {
			if strings.HasPrefix(key, "NODE_PATH=") {
				value := key[len("NODE_PATH="):]
				separator := ":"
				if fs.CheckIfWindows() {
					// On Windows, NODE_PATH is delimited by semicolons instead of colons
					separator = ";"
				}
				buildOptions.NodePaths = strings.Split(value, separator)
				break
			}
		}

		// Read from stdin when there are no entry points
		if len(analyseOptions.EntryPoints) == 0 {
			if analyseOptions.Stdin == nil {
				analyseOptions.Stdin = &api.StdinOptions{}
			}
			bytes, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				logger.PrintErrorToStderr(osArgs, fmt.Sprintf(
					"Could not read from stdin: %s", err.Error()))
				return 1
			}
			analyseOptions.Stdin.Contents = string(bytes)
			analyseOptions.Stdin.ResolveDir, _ = os.Getwd()
		} else if analyseOptions.Stdin != nil {
			if analyseOptions.Stdin.Sourcefile != "" {
				logger.PrintErrorToStderr(osArgs,
					"\"sourcefile\" only applies when reading from stdin")
			} else {
				logger.PrintErrorToStderr(osArgs,
					"\"loader\" without extension only applies when reading from stdin")
			}
			return 1
		}

		// Run the build and stop if there were errors
		result := api.Analyse(*analyseOptions)
		if len(result.Errors) > 0 {
			return 1
		}

	case err != nil:
		logger.PrintErrorToStderr(osArgs, err.Error())
		return 1
	}

	return 0
}

func printSummary(osArgs []string, outputFiles []api.OutputFile, start time.Time) {
	var table logger.SummaryTable = make([]logger.SummaryTableEntry, len(outputFiles))

	if len(outputFiles) > 0 {
		if cwd, err := os.Getwd(); err == nil {
			if realFS, err := fs.RealFS(fs.RealFSOptions{AbsWorkingDir: cwd}); err == nil {
				for i, file := range outputFiles {
					path, ok := realFS.Rel(realFS.Cwd(), file.Path)
					if !ok {
						path = file.Path
					}
					base := realFS.Base(path)
					n := len(file.Contents)
					var size string
					if n < 1024 {
						size = fmt.Sprintf("%db ", n)
					} else if n < 1024*1024 {
						size = fmt.Sprintf("%.1fkb", float64(n)/(1024))
					} else if n < 1024*1024*1024 {
						size = fmt.Sprintf("%.1fmb", float64(n)/(1024*1024))
					} else {
						size = fmt.Sprintf("%.1fgb", float64(n)/(1024*1024*1024))
					}
					table[i] = logger.SummaryTableEntry{
						Dir:         path[:len(path)-len(base)],
						Base:        base,
						Size:        size,
						Bytes:       n,
						IsSourceMap: strings.HasSuffix(base, ".map"),
					}
				}
			}
		}
	}

	logger.PrintSummary(osArgs, table, start)
}

func serveImpl(osArgs []string) error {
	host := ""
	portText := "0"
	servedir := ""

	// Filter out server-specific flags
	filteredArgs := make([]string, 0, len(osArgs))
	for _, arg := range osArgs {
		if arg == "--serve" {
			// Just ignore this flag
		} else if strings.HasPrefix(arg, "--serve=") {
			portText = arg[len("--serve="):]
		} else if strings.HasPrefix(arg, "--servedir=") {
			servedir = arg[len("--servedir="):]
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// Specifying the host is optional
	if strings.ContainsRune(portText, ':') {
		var err error
		host, portText, err = net.SplitHostPort(portText)
		if err != nil {
			return err
		}
	}

	// Parse the port
	port, err := strconv.ParseInt(portText, 10, 32)
	if err != nil {
		return err
	}
	if port < 0 || port > 0xFFFF {
		return fmt.Errorf("Invalid port number: %s", portText)
	}

	options := newBuildOptions()

	// Apply defaults appropriate for the CLI
	options.ErrorLimit = 5
	options.LogLevel = api.LogLevelInfo

	if err := parseOptionsImpl(filteredArgs, &options, nil, nil); err != nil {
		logger.PrintErrorToStderr(filteredArgs, err.Error())
		return err
	}

	serveOptions := api.ServeOptions{
		Port:     uint16(port),
		Host:     host,
		Servedir: servedir,
		OnRequest: func(args api.ServeOnRequestArgs) {
			logger.PrintText(os.Stderr, logger.LevelInfo, filteredArgs, func(colors logger.Colors) string {
				statusColor := colors.Red
				if args.Status >= 200 && args.Status <= 299 {
					statusColor = colors.Green
				} else if args.Status >= 300 && args.Status <= 399 {
					statusColor = colors.Yellow
				}
				return fmt.Sprintf("%s%s - %q %s%d%s [%dms]%s\n",
					colors.Dim, args.RemoteAddress, args.Method+" "+args.Path,
					statusColor, args.Status, colors.Dim, args.TimeInMS, colors.Default)
			})
		},
	}

	result, err := api.Serve(serveOptions, options)
	if err != nil {
		return err
	}

	// Show what actually got bound if the port was 0
	logger.PrintText(os.Stderr, logger.LevelInfo, filteredArgs, func(colors logger.Colors) string {
		return fmt.Sprintf("%s\n > %shttp://%s:%d/%s\n\n",
			colors.Default, colors.Underline, result.Host, result.Port, colors.Default)
	})
	return result.Wait()
}
