package bundler

import (
	"github.com/evanw/esbuild/internal/parser"
	"testing"
)

func TestTSImportStarES6Unused(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  0() {
    // /entry.ts
    let foo = 234;
    console.log(foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarES6Capture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  0() {
    // /foo.ts
    var exports = {};
    __export(exports, {
      foo: () => foo2
    });
    const foo2 = 123;

    // /entry.ts
    let foo = 234;
    console.log(exports, foo2, foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarES6NoCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns.foo, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  0() {
    // /foo.ts
    const foo2 = 123;

    // /entry.ts
    let foo = 234;
    console.log(foo2, foo2, foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarES6ExportImportStarUnused(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import {ns} from './bar'
				let foo = 234
				console.log(foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				import * as ns from './foo'
				export {ns}
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  0() {
    // /entry.ts
    let foo = 234;
    console.log(foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarES6ExportImportStarNoCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import {ns} from './bar'
				let foo = 234
				console.log(ns.foo, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				import * as ns from './foo'
				export {ns}
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1() {
    // /foo.ts
    var exports = {};
    __export(exports, {
      foo: () => foo2
    });
    const foo2 = 123;

    // /bar.ts

    // /entry.ts
    let foo = 234;
    console.log(exports.foo, exports.foo, foo);
  }
}, 1);
`,
		},
	})
}

func TestTSImportStarES6ExportImportStarCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import {ns} from './bar'
				let foo = 234
				console.log(ns, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				import * as ns from './foo'
				export {ns}
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1() {
    // /foo.ts
    var exports = {};
    __export(exports, {
      foo: () => foo2
    });
    const foo2 = 123;

    // /bar.ts

    // /entry.ts
    let foo = 234;
    console.log(exports, exports.foo, foo);
  }
}, 1);
`,
		},
	})
}

func TestTSImportStarES6ExportStarAsUnused(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import {ns} from './bar'
				let foo = 234
				console.log(foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				export * as ns from './foo'
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  0() {
    // /entry.ts
    let foo = 234;
    console.log(foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarES6ExportStarAsNoCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import {ns} from './bar'
				let foo = 234
				console.log(ns.foo, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				export * as ns from './foo'
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1() {
    // /foo.ts
    var exports = {};
    __export(exports, {
      foo: () => foo2
    });
    const foo2 = 123;

    // /bar.ts

    // /entry.ts
    let foo = 234;
    console.log(exports.foo, exports.foo, foo);
  }
}, 1);
`,
		},
	})
}

func TestTSImportStarES6ExportStarAsCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import {ns} from './bar'
				let foo = 234
				console.log(ns, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				export * as ns from './foo'
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1() {
    // /foo.ts
    var exports = {};
    __export(exports, {
      foo: () => foo2
    });
    const foo2 = 123;

    // /bar.ts

    // /entry.ts
    let foo = 234;
    console.log(exports, exports.foo, foo);
  }
}, 1);
`,
		},
	})
}

func TestTSImportStarES6ExportStarUnused(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './bar'
				let foo = 234
				console.log(foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				export * from './foo'
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  0() {
    // /entry.ts
    let foo = 234;
    console.log(foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarES6ExportStarNoCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './bar'
				let foo = 234
				console.log(ns.foo, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				export * from './foo'
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1() {
    // /foo.ts
    const foo2 = 123;

    // /bar.ts

    // /entry.ts
    let foo = 234;
    console.log(foo2, foo2, foo);
  }
}, 1);
`,
		},
	})
}

func TestTSImportStarES6ExportStarCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './bar'
				let foo = 234
				console.log(ns, ns.foo, foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
			"/bar.ts": `
				export * from './foo'
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1() {
    // /foo.ts
    const foo2 = 123;

    // /bar.ts
    var bar = {};
    __export(bar, {
      foo: () => foo2
    });

    // /entry.ts
    let foo = 234;
    console.log(bar, foo2, foo);
  }
}, 1);
`,
		},
	})
}

func TestTSImportStarCommonJSUnused(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(foo)
			`,
			"/foo.ts": `
				exports.foo = 123
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  0() {
    // /entry.ts
    let foo = 234;
    console.log(foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarCommonJSCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns, ns.foo, foo)
			`,
			"/foo.ts": `
				exports.foo = 123
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1(exports) {
    // /foo.ts
    exports.foo = 123;
  },

  0() {
    // /entry.ts
    const ns = __import(1 /* ./foo */);
    let foo = 234;
    console.log(ns, ns.foo, foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarCommonJSNoCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns.foo, ns.foo, foo)
			`,
			"/foo.ts": `
				exports.foo = 123
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    true,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `bootstrap({
  1(exports) {
    // /foo.ts
    exports.foo = 123;
  },

  0() {
    // /entry.ts
    const ns = __import(1 /* ./foo */);
    let foo = 234;
    console.log(ns.foo, ns.foo, foo);
  }
}, 0);
`,
		},
	})
}

func TestTSImportStarES6AndCommonJS(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry1.js": `
				import * as ns from './foo'
				console.log(ns.foo)
			`,
			"/entry2.js": `
				const ns = require('./foo')
				console.log(ns.foo)
			`,
			"/foo.ts": `
				export const foo = 123
			`,
		},
		entryPaths: []string{"/entry1.js", "/entry2.js"},
		parseOptions: parser.ParseOptions{
			IsBundling: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:   true,
			AbsOutputDir: "/out",
		},
		expected: map[string]string{
			"/out/entry1.js": `bootstrap({
  2(exports) {
    // /foo.ts
    __export(exports, {
      foo: () => foo
    });
    const foo = 123;
  },

  0() {
    // /entry1.js
    const ns = __import(2 /* ./foo */);
    console.log(ns.foo);
  }
}, 0);
`,
			"/out/entry2.js": `bootstrap({
  2(exports) {
    // /foo.ts
    __export(exports, {
      foo: () => foo
    });
    const foo = 123;
  },

  1() {
    // /entry2.js
    const ns = __require(2 /* ./foo */);
    console.log(ns.foo);
  }
}, 1);
`,
		},
	})
}

func TestTSImportStarES6NoBundleUnused(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(foo)
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: false,
		},
		bundleOptions: BundleOptions{
			IsBundling:    false,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `let foo = 234;
console.log(foo);
`,
		},
	})
}

func TestTSImportStarES6NoBundleCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns, ns.foo, foo)
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: false,
		},
		bundleOptions: BundleOptions{
			IsBundling:    false,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `import * as ns from "./foo";
let foo = 234;
console.log(ns, ns.foo, foo);
`,
		},
	})
}

func TestTSImportStarES6NoBundleNoCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns.foo, ns.foo, foo)
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling: false,
		},
		bundleOptions: BundleOptions{
			IsBundling:    false,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `import * as ns from "./foo";
let foo = 234;
console.log(ns.foo, ns.foo, foo);
`,
		},
	})
}

func TestTSImportStarES6MangleNoBundleUnused(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(foo)
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling:   false,
			MangleSyntax: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    false,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `let foo = 234;
console.log(foo);
`,
		},
	})
}

func TestTSImportStarES6MangleNoBundleCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns, ns.foo, foo)
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling:   false,
			MangleSyntax: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    false,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `import * as ns from "./foo";
let foo = 234;
console.log(ns, ns.foo, foo);
`,
		},
	})
}

func TestTSImportStarES6MangleNoBundleNoCapture(t *testing.T) {
	expectBundled(t, bundled{
		files: map[string]string{
			"/entry.ts": `
				import * as ns from './foo'
				let foo = 234
				console.log(ns.foo, ns.foo, foo)
			`,
		},
		entryPaths: []string{"/entry.ts"},
		parseOptions: parser.ParseOptions{
			IsBundling:   false,
			MangleSyntax: true,
		},
		bundleOptions: BundleOptions{
			IsBundling:    false,
			AbsOutputFile: "/out.js",
		},
		expected: map[string]string{
			"/out.js": `import * as ns from "./foo";
let foo = 234;
console.log(ns.foo, ns.foo, foo);
`,
		},
	})
}
