#!/usr/bin/env node

/**
 * Script to verify i18n translation coverage
 * Checks that all keys in en.json exist in other locale files
 */

import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const localesDir = join(__dirname, '../src/locales');
const referenceLocale = 'en';
const otherLocales = ['fr', 'es', 'de', 'it'];

/**
 * Flatten nested object keys
 * e.g. { a: { b: { c: 'value' } } } => ['a.b.c']
 */
function flattenKeys(obj, prefix = '') {
  return Object.keys(obj).reduce((acc, key) => {
    const newKey = prefix ? `${prefix}.${key}` : key;
    if (typeof obj[key] === 'object' && obj[key] !== null && !Array.isArray(obj[key])) {
      return acc.concat(flattenKeys(obj[key], newKey));
    }
    return acc.concat(newKey);
  }, []);
}

/**
 * Check if a key exists in an object
 * Handles keys with literal dots (e.g., "document.created")
 */
function hasKey(obj, keyPath) {
  const keys = keyPath.split('.');
  let current = obj;

  for (let i = 0; i < keys.length; i++) {
    if (typeof current !== 'object' || current === null) {
      return false;
    }

    // Try the remaining path as a single key first (for literal dots)
    const remainingPath = keys.slice(i).join('.');
    if (remainingPath in current) {
      current = current[remainingPath];
      return true;
    }

    // Otherwise, try the next segment
    const key = keys[i];
    if (!(key in current)) {
      return false;
    }
    current = current[key];
  }

  return true;
}

// Load reference locale
let referenceMessages;
try {
  referenceMessages = JSON.parse(
    readFileSync(join(localesDir, `${referenceLocale}.json`), 'utf-8')
  );
} catch (error) {
  console.error(`❌ Failed to load reference locale (${referenceLocale}.json):`, error.message);
  process.exit(1);
}

const referenceKeys = flattenKeys(referenceMessages);
console.log(`📚 Reference locale (${referenceLocale}): ${referenceKeys.length} keys\n`);

let hasErrors = false;
const report = [];

// Check each locale
for (const locale of otherLocales) {
  try {
    const messages = JSON.parse(
      readFileSync(join(localesDir, `${locale}.json`), 'utf-8')
    );

    const localeKeys = flattenKeys(messages);
    const missingKeys = referenceKeys.filter(key => !hasKey(messages, key));
    const extraKeys = localeKeys.filter(key => !hasKey(referenceMessages, key));

    if (missingKeys.length === 0 && extraKeys.length === 0) {
      console.log(`✅ ${locale}.json: ${localeKeys.length} keys (complete)`);
      report.push({ locale, status: 'ok', total: localeKeys.length });
    } else {
      hasErrors = true;
      console.log(`⚠️  ${locale}.json: ${localeKeys.length} keys`);

      if (missingKeys.length > 0) {
        console.log(`   Missing ${missingKeys.length} keys:`);
        missingKeys.slice(0, 10).forEach(key => console.log(`     - ${key}`));
        if (missingKeys.length > 10) {
          console.log(`     ... and ${missingKeys.length - 10} more`);
        }
      }

      if (extraKeys.length > 0) {
        console.log(`   Extra ${extraKeys.length} keys (not in reference):`);
        extraKeys.slice(0, 5).forEach(key => console.log(`     - ${key}`));
        if (extraKeys.length > 5) {
          console.log(`     ... and ${extraKeys.length - 5} more`);
        }
      }

      report.push({
        locale,
        status: 'incomplete',
        total: localeKeys.length,
        missing: missingKeys.length,
        extra: extraKeys.length
      });
    }
    console.log('');
  } catch (error) {
    console.error(`❌ Failed to load ${locale}.json:`, error.message);
    hasErrors = true;
    report.push({ locale, status: 'error', error: error.message });
  }
}

// Summary
console.log('='.repeat(60));
if (hasErrors) {
  console.log('❌ Translation coverage check FAILED');
  console.log('\nSome locales have missing or extra keys.');
  console.log('Please update the translations to match the reference locale.');
  process.exit(1);
} else {
  console.log('✅ All translations are complete!');
  console.log(`\nAll ${otherLocales.length} locales have ${referenceKeys.length} keys matching the reference.`);
  process.exit(0);
}
