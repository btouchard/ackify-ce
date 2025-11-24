#!/usr/bin/env node
/**
 * Patch nyc@15 to handle glob@10+ and rimraf@5+
 * This fixes the "original argument must be of type function" error
 */

const fs = require('fs');
const path = require('path');

const nycIndexPath = path.join(__dirname, '../node_modules/@cypress/code-coverage/node_modules/nyc/index.js');

if (!fs.existsSync(nycIndexPath)) {
  console.log('⚠️  nyc/index.js not found, skipping patch');
  process.exit(0);
}

// Read current content
let content = fs.readFileSync(nycIndexPath, 'utf-8');

// Check if already patched
if (content.includes('// PATCHED for glob@10+')) {
  console.log('✅ nyc@15 already patched');
  process.exit(0);
}

// Patch glob and rimraf imports to handle modern versions
content = content.replace(
  "const glob = promisify(require('glob'))",
  `// PATCHED for glob@10+ compatibility
const globModule = require('glob')
const glob = typeof globModule === 'function' ? promisify(globModule) : globModule.glob`
);

content = content.replace(
  "const rimraf = promisify(require('rimraf'))",
  `// PATCHED for rimraf@5+ compatibility
const rimrafModule = require('rimraf')
const rimraf = typeof rimrafModule === 'function' ? promisify(rimrafModule) : rimrafModule.rimraf`
);

fs.writeFileSync(nycIndexPath, content);
console.log('✅ Patched nyc@15 for glob@10+ and rimraf@5+ compatibility');
