const { readFileSync, writeFileSync } = require('fs')
const ver = readFileSync(`${__dirname}/../version.txt`, 'utf8').trim()

const dirs = [
  'esbuild',
  'esbuild-darwin-64',
  'esbuild-freebsd-64',
  'esbuild-freebsd-arm64',
  'esbuild-linux-32',
  'esbuild-linux-64',
  'esbuild-linux-arm',
  'esbuild-linux-arm64',
  'esbuild-linux-mips64le',
  'esbuild-linux-ppc64le',
  'esbuild-wasm',
  'esbuild-windows-32',
  'esbuild-windows-64'
]

for (const dir of dirs) {
  const pkgName = `${__dirname}/../npm/${dir}/package.json`
  writeFileSync(pkgName, fixPackage(readFileSync(pkgName, 'utf8')))
}

function fixPackage (pkg) {
  let skip
  return pkg
    .split(/\r?\n/)
    .filter(line => {
      if (skip) {
        if (line.startsWith('=======')) skip = false
        return
      }
      if (line.startsWith('<<<<<<<')) {
        skip = true
        return
      }
      if (line.startsWith('>>>>>>>')) return
      return true
    })
    .map(line => line.replace(/"version": "[^"]+"/, `"version": "${ver}"`))
    .join('\n')
}
