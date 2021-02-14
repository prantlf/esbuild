import * as types from "./types"
import * as common from "./common"

declare const ESBUILD_VERSION: string;
declare let WEB_WORKER_SOURCE_CODE: string

export let version = ESBUILD_VERSION;

export const build: typeof types.build = () => {
  throw new Error(`The "build" API only works in node`);
};

export const serve: typeof types.serve = () => {
  throw new Error(`The "serve" API only works in node`);
};

export const transform: typeof types.transform = () => {
  throw new Error(`The "transform" API only works in node`);
};

export const analyse: typeof types.analyse = () => {
  throw new Error(`The "analyse" API only works in node`);
};

export const buildSync: typeof types.buildSync = () => {
  throw new Error(`The "buildSync" API only works in node`);
};

export const transformSync: typeof types.transformSync = () => {
  throw new Error(`The "transformSync" API only works in node`);
};

export const analyseSync: typeof types.analyseSync = () => {
  throw new Error(`The "analyseSync" API only works in node`);
};

export const startService: typeof types.startService = common.longLivedService(() => '', async (options) => {
  if (!options) throw new Error('Must provide an options object to "startService"');
  options = common.validateServiceOptions(options)!;
  let wasmURL = options.wasmURL;
  let useWorker = options.worker !== false;
  if (!wasmURL) throw new Error('Must provide the "wasmURL" option');
  wasmURL += '';
  let res = await fetch(wasmURL);
  if (!res.ok) throw new Error(`Failed to download ${JSON.stringify(wasmURL)}`);
  let wasm = await res.arrayBuffer();
  let code = `{` +
    `let global={};` +
    `for(let o=self;o;o=Object.getPrototypeOf(o))` +
    `for(let k of Object.getOwnPropertyNames(o))` +
    `if(!(k in global))` +
    `Object.defineProperty(global,k,{get:()=>self[k]});` +
    WEB_WORKER_SOURCE_CODE +
    `}`
  let worker: {
    onmessage: ((event: any) => void) | null
    postMessage: (data: Uint8Array | ArrayBuffer) => void
    terminate: () => void
  }

  if (useWorker) {
    // Run esbuild off the main thread
    let blob = new Blob([code], { type: 'application/javascript' })
    worker = new Worker(URL.createObjectURL(blob))
  } else {
    // Run esbuild on the main thread
    let fn = new Function('postMessage', code + `var onmessage; return m => onmessage(m)`)
    let onmessage = fn((data: Uint8Array) => worker.onmessage!({ data }))
    worker = {
      onmessage: null,
      postMessage: data => onmessage({ data }),
      terminate() {
      },
    }
  }

  worker.postMessage(wasm)
  worker.onmessage = ({ data }) => readFromStdout(data)

  let { readFromStdout, afterClose, service } = common.createChannel({
    writeToStdin(bytes) {
      worker.postMessage(bytes)
    },
    isSync: false,
    isBrowser: true,
  })

  return {
    build: (options: types.BuildOptions): Promise<any> =>
      new Promise<types.BuildResult>((resolve, reject) =>
        service.buildOrServe('build', null, null, options, false, '/', (err, res) =>
          err ? reject(err) : resolve(res as types.BuildResult))),
    transform: (input, options) => {
      input += '';
      return new Promise((resolve, reject) =>
        service.transform('transform', null, input, options || {}, false, {
          readFile(_, callback) { callback(new Error('Internal error'), null); },
          writeFile(_, callback) { callback(null); },
        }, (err, res) => err ? reject(err) : resolve(res!)))
    },
    analyse: (options: types.AnalyseOptions): Promise<any> =>
      new Promise<types.AnalyseResult>((resolve, reject) =>
        service.analyse(null, options, false, "/", (err, res) =>
          err ? reject(err) : resolve(res as types.AnalyseResult))),
    serve() {
      throw new Error(`The "serve" API only works in node`)
    },
    buildSync() {
      throw new Error(`The "buildSync" API only works in node`);
    },
    transformSync() {
      throw new Error(`The "transformSync" API only works in node`);
    },
    analyseSync() {
      throw new Error(`The "analyseSync" API only works in node`);
    },
    stop() {
      // Note: This is now never called
      worker.terminate()
      afterClose()
    },
  }
});
