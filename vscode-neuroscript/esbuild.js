// vscode-neuroscript/esbuild.js
const { build } = require('esbuild');

const entryPoint = 'src/extension.ts'; // Your confirmed entry point
const outfile = 'dist/extension.js';   // Your confirmed output
const isWatchMode = process.argv.includes('--watch'); // Check for --watch flag

// Define the base build options
const buildOptions = {
  entryPoints: [entryPoint],
  outfile: outfile,
  bundle: true,
  platform: 'node',
  target: 'node16', // Or a version appropriate for the VS Code version you target
  external: ['vscode'], // Exclude 'vscode' module, as it's provided by the VS Code runtime
  format: 'cjs', // CommonJS format, standard for VS Code extensions
  sourcemap: true, // Creates .map files for debugging
  logLevel: 'info',
};

// Conditionally add the watch configuration ONLY if isWatchMode is true
if (isWatchMode) {
  buildOptions.watch = {
    onRebuild(error, result) {
      if (error) {
        console.error('esbuild: watch build failed:', error);
      } else {
        console.log('esbuild: watch build succeeded.');
      }
    },
  };
}

build(buildOptions)
  .then(() => {
    if (isWatchMode) {
      console.log(`esbuild: watching for changes to ${entryPoint}...`);
      // esbuild keeps the process alive in watch mode
    } else {
      console.log(`esbuild: Bundling ${entryPoint} to ${outfile} complete.`);
    }
  })
  .catch((err) => {
    console.error("esbuild build failed:", err);
    process.exit(1);
  });