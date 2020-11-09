package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/evanw/esbuild/internal/compat"
	"github.com/evanw/esbuild/internal/logger"
)

type LanguageTarget int8

type JSXOptions struct {
	Parse    bool
	Factory  []string
	Fragment []string
}

type TSOptions struct {
	Parse bool
}

type Platform uint8

const (
	PlatformBrowser Platform = iota
	PlatformNode
)

type StrictOptions struct {
	// Loose:  "class Foo { foo = 1 }" => "class Foo { constructor() { this.foo = 1; } }"
	// Strict: "class Foo { foo = 1 }" => "class Foo { constructor() { __publicField(this, 'foo', 1); } }"
	//
	// The disadvantage of strictness here is code bloat and performance. The
	// advantage is following the class field specification accurately. For
	// example, loose mode will incorrectly trigger setter methods while strict
	// mode won't.
	ClassFields bool
}

type SourceMap uint8

const (
	SourceMapNone SourceMap = iota
	SourceMapInline
	SourceMapLinkedWithComment
	SourceMapExternalWithoutComment
)

type Loader int

const (
	LoaderNone Loader = iota
	LoaderJS
	LoaderJSX
	LoaderTS
	LoaderTSX
	LoaderJSON
	LoaderText
	LoaderBase64
	LoaderDataURL
	LoaderFile
	LoaderBinary
	LoaderCSS
	LoaderDefault
)

func (loader Loader) IsTypeScript() bool {
	return loader == LoaderTS || loader == LoaderTSX
}

func (loader Loader) CanHaveSourceMap() bool {
	return loader == LoaderJS || loader == LoaderJSX || loader == LoaderTS || loader == LoaderTSX
}

type Format uint8

const (
	// This is used when not bundling. It means to preserve whatever form the
	// import or export was originally in. ES6 syntax stays ES6 syntax and
	// CommonJS syntax stays CommonJS syntax.
	FormatPreserve Format = iota

	// IIFE stands for immediately-invoked function expression. That looks like
	// this:
	//
	//   (() => {
	//     ... bundled code ...
	//   })();
	//
	// If the optional ModuleName is configured, then we'll write out this:
	//
	//   let moduleName = (() => {
	//     ... bundled code ...
	//     return exports;
	//   })();
	//
	FormatIIFE

	// The CommonJS format looks like this:
	//
	//   ... bundled code ...
	//   module.exports = exports;
	//
	FormatCommonJS

	// The UMD format without external dependencies looks like this:
	//
	// (function(root, factory) {
	//   if (typeof define === 'function' && define.amd) {
	//     define(factory);
	//   } else if (typeof module === 'object' && module.exports) {
	//     module.exports = factory();
	//   } else {
	//     root.returnExports = factory();
	//   }
	// }(typeof self !== 'undefined' ? self : this, function() {
	//   ... bundled code ...
	// }));
	FormatUMD

	// The ES module format looks like this:
	//
	//   ... bundled code ...
	//   export {...};
	//
	FormatESModule

	// Just concatenates source modules, which are supposed to be in a
	// concatenate-able format already. Like the AMD modules, for example.
	//
	// This format cannot be selected by the configuration. It will be
	// chosen automatically when an AMD bundle is going to be produced.
	//
	//   define("module1", ...);
	//   define("module2", ...);
	//
	FormatJoin
)

func (f Format) KeepES6ImportExportSyntax() bool {
	return f == FormatPreserve || f == FormatESModule
}

func (f Format) String() string {
	switch f {
	case FormatIIFE:
		return "iife"
	case FormatCommonJS:
		return "cjs"
	case FormatUMD:
		return "umd"
	case FormatESModule:
		return "esm"
	}
	return ""
}

type StdinInfo struct {
	Loader        Loader
	Contents      string
	SourceFile    string
	AbsResolveDir string
}

type WildcardPattern struct {
	Prefix string
	Suffix string
}

type ExternalModules struct {
	NodeModules map[string]bool
	AbsPaths    map[string]bool
	Patterns    []WildcardPattern
}

type Mode uint8

const (
	ModePassThrough Mode = iota
	ModeConvertFormat
	ModeBundle
)

type AMDLoadableScript struct {
	ReplacementPattern string
	ReplacementValue   string
	ReplacementRegexp  *regexp.Regexp
}

type AMDPlugin struct {
	FileExtensions      []string
	AppendFileExtension bool
	LoadScript          *AMDLoadableScript
}

type AMDOptions struct {
	BaseUrl   string
	Paths     map[string]string
	Map       map[string]map[string]string
	StarMap   map[string]string
	Namespace string
	Plugins   map[string]*AMDPlugin

	Parse               bool
	MappedModuleNames   bool
	KnownFileExtensions map[string]bool

	backwardPaths          map[string]string
	backwardPathsMutex     *sync.RWMutex
	pluginExpressions      map[string]string
	pluginExpressionsMutex *sync.RWMutex
}

type Options struct {
	Mode              Mode
	RemoveWhitespace  bool
	MinifyIdentifiers bool
	MangleSyntax      bool
	CodeSplitting     bool

	// Setting this to true disables warnings about code that is very likely to
	// be a bug. This is used to ignore issues inside "node_modules" directories.
	// This has caught real issues in the past. However, it's not esbuild's job
	// to find bugs in other libraries, and these warnings are problematic for
	// people using these libraries with esbuild. The only fix is to either
	// disable all esbuild warnings and not get warnings about your own code, or
	// to try to get the warning fixed in the affected library. This is
	// especially annoying if the warning is a false positive as was the case in
	// https://github.com/firebase/firebase-js-sdk/issues/3814. So these warnings
	// are now disabled for code inside "node_modules" directories.
	SuppressWarningsAboutWeirdCode bool

	// If true, make sure to generate a single file that can be written to stdout
	WriteToStdout bool

	OmitRuntimeForTests     bool
	PreserveUnusedImportsTS bool
	UseDefineForClassFields bool
	AvoidTDZ                bool
	ASCIIOnly               bool
	KeepNames               bool
	IgnoreDCEAnnotations    bool

	Defines  *ProcessedDefines
	AMD      AMDOptions
	TS       TSOptions
	JSX      JSXOptions
	Platform Platform

	UnsupportedJSFeatures  compat.JSFeature
	UnsupportedCSSFeatures compat.CSSFeature

	ExtensionOrder  []string
	MainFields      []string
	ExternalModules ExternalModules

	AbsOutputFile     string
	AbsOutputDir      string
	AbsOutputBase     string
	OutputExtensions  map[string]string
	ModuleName        []string
	AMDConfig         string
	TsConfigOverride  string
	ExtensionToLoader map[string]Loader
	OutputFormat      Format
	PublicPath        string
	InjectAbsPaths    []string
	InjectedFiles     []InjectedFile
	Banner            string
	Footer            string

	Plugins []Plugin

	// If present, metadata about the bundle is written as JSON here
	AbsMetadataFile string

	SourceMap SourceMap
	Stdin     *StdinInfo
}

type InjectedFile struct {
	Path        string
	SourceIndex uint32
	Exports     []string
}

func (options *Options) OutputExtensionFor(key string) string {
	if ext, ok := options.OutputExtensions[key]; ok {
		return ext
	}
	return key
}

var filterMutex sync.Mutex
var filterCache map[string]*regexp.Regexp

func compileFilter(filter string) (result *regexp.Regexp) {
	if filter == "" {
		// Must provide a filter
		return nil
	}
	ok := false

	// Cache hit?
	(func() {
		filterMutex.Lock()
		defer filterMutex.Unlock()
		if filterCache != nil {
			result, ok = filterCache[filter]
		}
	})()
	if ok {
		return
	}

	// Cache miss
	result, err := regexp.Compile(filter)
	if err != nil {
		return nil
	}

	// Cache for next time
	filterMutex.Lock()
	defer filterMutex.Unlock()
	if filterCache == nil {
		filterCache = make(map[string]*regexp.Regexp)
	}
	filterCache[filter] = result
	return
}

func CompileFilterForPlugin(pluginName string, kind string, filter string) (*regexp.Regexp, error) {
	if filter == "" {
		return nil, fmt.Errorf("[%s] %q is missing a filter", pluginName, kind)
	}

	result := compileFilter(filter)
	if result == nil {
		return nil, fmt.Errorf("[%s] %q filter is not a valid Go regular expression: %q", pluginName, kind, filter)
	}

	return result, nil
}

func PluginAppliesToPath(path logger.Path, filter *regexp.Regexp, namespace string) bool {
	return (namespace == "" || path.Namespace == namespace) && filter.MatchString(path.Text)
}

////////////////////////////////////////////////////////////////////////////////
// Plugin API

type Plugin struct {
	Name      string
	OnResolve []OnResolve
	OnLoad    []OnLoad
}

type OnResolve struct {
	Name      string
	Filter    *regexp.Regexp
	Namespace string
	Callback  func(OnResolveArgs) OnResolveResult
}

type OnResolveArgs struct {
	Path       string
	Importer   logger.Path
	ResolveDir string
}

type OnResolveResult struct {
	PluginName string

	Path     logger.Path
	External bool

	Msgs        []logger.Msg
	ThrownError error
}

type OnLoad struct {
	Name      string
	Filter    *regexp.Regexp
	Namespace string
	Callback  func(OnLoadArgs) OnLoadResult
}

type OnLoadArgs struct {
	Path logger.Path
}

type OnLoadResult struct {
	PluginName string

	Contents      *string
	AbsResolveDir string
	Loader        Loader

	Msgs        []logger.Msg
	ThrownError error
}

func (options *AMDOptions) Init(baseUrl string) {
	options.BaseUrl = baseUrl
	options.Paths = make(map[string]string)
	options.Map = make(map[string]map[string]string)
	options.StarMap = make(map[string]string)
	options.Plugins = map[string]*AMDPlugin{
		"json": {
			FileExtensions:      []string{".json"},
			AppendFileExtension: false,
		},
		"text": {
			FileExtensions:      []string{".txt"},
			AppendFileExtension: false,
		},
		"css": {
			FileExtensions:      []string{".css"},
			AppendFileExtension: true,
		},
	}
	options.KnownFileExtensions = map[string]bool{
		".js":   true,
		".json": true,
		".txt":  true,
		".css":  true,
	}
	options.backwardPaths = make(map[string]string)
	options.backwardPathsMutex = &sync.RWMutex{}
	options.pluginExpressions = make(map[string]string)
	options.pluginExpressionsMutex = &sync.RWMutex{}
}

// The following configuration:
//
// {
//   "paths": {
//     "foo":     "foo2",
//     "foo/bar": "foo/bar2"
//   }
// }
//
// will map "foo/baz" to "foo2/baz" and "foo/bar/baz" to "foo/bar2/baz"
// using the longest matching path segment first.
func (options *AMDOptions) ModuleNameToPath(importPath string, sourcePath string) string {
	mappedSource := options.ModulePathToName(sourcePath)
	if mappedSource == "" {
		mappedSource, _ = filepath.Rel(options.BaseUrl, sourcePath)
		if strings.HasSuffix(mappedSource, ".js") {
			mappedSource = mappedSource[:len(mappedSource)-3]
		}
	}
	var inputPath string
	var mappedPath string
	scope := options.Map[mappedSource]
	if scope != nil {
		mappedPath = mapSegmentedPath(importPath, scope)
		if mappedPath != "" {
			inputPath = mappedPath
		}
	}
	if inputPath == "" {
		mappedPath = mapSegmentedPath(importPath, options.StarMap)
		if mappedPath != "" {
			inputPath = mappedPath
		} else {
			inputPath = importPath
		}
	}
	targetPath := mapSegmentedPath(inputPath, options.Paths)
	if targetPath == "" {
		targetPath = mappedPath
	}
	if targetPath != "" {
		if !options.HasKnownFileExtension(targetPath) {
			targetPath += ".js"
		}
		options.backwardPathsMutex.Lock()
		options.backwardPaths[targetPath] = importPath
		options.backwardPathsMutex.Unlock()
		return targetPath
	}
	if mappedPath != "" && !options.HasKnownFileExtension(mappedPath) {
		mappedPath += ".js"
	}
	return mappedPath
}

func (options *AMDOptions) ModulePathToName(sourcePath string) string {
	modulePath, _ := filepath.Rel(options.BaseUrl, sourcePath)
	options.backwardPathsMutex.RLock()
	importPath := options.backwardPaths[modulePath]
	options.backwardPathsMutex.RUnlock()
	return importPath
}

func (options *AMDOptions) UsesPlugin(importPath string) bool {
	return strings.Contains(importPath, "!")
}

func (options *AMDOptions) ParsePluginExpression(importPath string) (string, string, bool) {
	separator := strings.Index(importPath, "!")
	if separator > 0 {
		pluginPrefix := importPath[:separator]
		if options.Plugins[pluginPrefix] != nil {
			return pluginPrefix, importPath[separator+1:], true
		}
	}
	return "", "", false
}

func (options *AMDOptions) PluginExpressionToModulePath(importPath string, moduleName string, sourcePath string) string {
	separator := strings.Index(importPath, "!")
	pluginPrefix := importPath[:separator]
	plugin := options.Plugins[pluginPrefix]
	importPath = importPath[separator+1:]
	if plugin.AppendFileExtension {
		importPath += plugin.FileExtensions[0]
	}
	if plugin.LoadScript != nil {
		importPath = plugin.LoadScript.ReplacementRegexp.ReplaceAllString(importPath, plugin.LoadScript.ReplacementValue)
	}
	modulePath := options.ModuleNameToPath(importPath, sourcePath)
	if modulePath == "" {
		modulePath = importPath
		if !options.HasKnownFileExtension(modulePath) {
			modulePath += ".js"
		}
	}
	if strings.HasPrefix(modulePath, "./") || strings.HasPrefix(modulePath, "../") {
		modulePath, _ = filepath.Rel(options.BaseUrl, filepath.Join(filepath.Dir(sourcePath), modulePath))
	}
	options.pluginExpressionsMutex.Lock()
	options.pluginExpressions[modulePath] = moduleName
	options.pluginExpressionsMutex.Unlock()
	return importPath
}

func (options *AMDOptions) ModulePathToPluginExpression(sourcePath string) string {
	modulePath, _ := filepath.Rel(options.BaseUrl, sourcePath)
	options.pluginExpressionsMutex.RLock()
	moduleName := options.pluginExpressions[modulePath]
	options.pluginExpressionsMutex.RUnlock()
	return moduleName
}

func (options *AMDOptions) HasKnownFileExtension(sourcePath string) bool {
	return options.KnownFileExtensions[filepath.Ext(sourcePath)]
}

func (options *AMDOptions) IsSpecialModule(importPath string) bool {
	return importPath == "require" || importPath == "module" || importPath == "exports"
}

func (options *AMDOptions) IsExternalModule(importPath string) bool {
	mappedPath := mapSegmentedPath(importPath, options.StarMap)
	if mappedPath != "" {
		importPath = mappedPath
	}
	if len(options.Paths) > 0 {
		paths := options.Paths
		separator := len(importPath)
		parentPath := importPath
		for {
			parentPath := parentPath[:separator]
			targetPath := paths[parentPath]
			if targetPath != "" {
				return strings.HasPrefix(targetPath, "empty:")
			}
			separator = strings.LastIndex(parentPath, "/")
			if separator < 0 {
				break
			}
		}
	}
	return false
}

func mapSegmentedPath(importPath string, paths map[string]string) string {
	if len(paths) > 0 {
		separator := len(importPath)
		parentPath := importPath
		for {
			parentPath := parentPath[:separator]
			targetPath := paths[parentPath]
			if targetPath != "" {
				if targetPath == "." {
					return importPath[separator+1:]
				}
				return filepath.Join(targetPath, importPath[separator:])
			}
			separator = strings.LastIndex(parentPath, "/")
			if separator < 0 {
				break
			}
		}
	}
	return ""
}
