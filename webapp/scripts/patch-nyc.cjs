#!/usr/bin/env node
/**
 * Patch nyc@15 to handle Node.js 20+ promisify issues
 * This fixes the "original argument must be of type function" error
 */

const fs = require('fs');
const path = require('path');

const nycPath = path.join(__dirname, '../node_modules/@cypress/code-coverage/node_modules/nyc/lib/fs-promises.js');

if (!fs.existsSync(nycPath)) {
  console.log('⚠️  nyc fs-promises.js not found, skipping patch');
  process.exit(0);
}

const patchedContent = `'use strict'

const fs = require('fs')

const { promisify } = require('util')

module.exports = { ...fs }

// Promisify all functions for consistency
const fns = [
  'access',
  'appendFile',
  'chmod',
  'chown',
  'close',
  'copyFile',
  'fchmod',
  'fchown',
  'fdatasync',
  'fstat',
  'fsync',
  'ftruncate',
  'futimes',
  'lchmod',
  'lchown',
  'link',
  'lstat',
  'mkdir',
  'mkdtemp',
  'open',
  'read',
  'readdir',
  'readFile',
  'readlink',
  'realpath',
  'rename',
  'rmdir',
  'stat',
  'symlink',
  'truncate',
  'unlink',
  'utimes',
  'write',
  'writeFile'
]
fns.forEach(fn => {
  /* istanbul ignore else: all functions exist on OSX */
  if (fs[fn]) {
    try {
      module.exports[fn] = promisify(fs[fn])
    } catch (err) {
      // Fallback to original function if promisify fails (Node.js 20+ compat)
      module.exports[fn] = fs[fn]
    }
  }
})
`;

fs.writeFileSync(nycPath, patchedContent);
console.log('✅ Patched nyc@15 for Node.js 20+ compatibility');