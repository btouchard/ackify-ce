// SPDX-License-Identifier: AGPL-3.0-or-later
/**
 * Script to synchronize translation structure from fr.json to other locales
 * This script copies the structure and provides placeholder translations
 */

const fs = require('fs')
const path = require('path')

const localesDir = path.join(__dirname, '../src/locales')
const sourceLang = 'fr'
const targetLangs = ['it', 'de', 'es']

// Translation mappings for common terms
const commonTranslations = {
  it: {
    'Confirmation de Lecture': 'Conferma di Lettura',
    'Certifiez votre lecture avec une confirmation cryptographique Ed25519': 'Certifica la tua lettura con una conferma crittografica Ed25519',
    'Chargement du document...': 'Caricamento del documento...',
    'Veuillez patienter pendant que nous préparons le document pour la signature.': 'Attendere mentre prepariamo il documento per la firma.',
    'Aucun document spécifié': 'Nessun documento specificato',
    'Pour signer un document, ajoutez le paramètre {code} à l\'URL': 'Per firmare un documento, aggiungere il parametro {code} all\'URL',
    'Exemples :': 'Esempi:',
    'Lecture confirmée avec succès !': 'Lettura confermata con successo!',
    'Votre confirmation a été enregistrée de manière cryptographique et sécurisée.': 'La tua conferma è stata registrata in modo crittografico e sicuro.',
    'Une erreur est survenue': 'Si è verificato un errore',
  },
  de: {
    'Confirmation de Lecture': 'Lesebestätigung',
    'Certifiez votre lecture avec une confirmation cryptographique Ed25519': 'Bestätigen Sie Ihre Lektüre mit einer kryptografischen Ed25519-Bestätigung',
    'Chargement du document...': 'Dokument wird geladen...',
    'Veuillez patienter pendant que nous préparons le document pour la signature.': 'Bitte warten Sie, während wir das Dokument zur Unterschrift vorbereiten.',
    'Aucun document spécifié': 'Kein Dokument angegeben',
    'Pour signer un document, ajoutez le paramètre {code} à l\'URL': 'Um ein Dokument zu signieren, fügen Sie den Parameter {code} zur URL hinzu',
    'Exemples :': 'Beispiele:',
    'Lecture confirmée avec succès !': 'Lektüre erfolgreich bestätigt!',
    'Votre confirmation a été enregistrée de manière cryptographique et sécurisée.': 'Ihre Bestätigung wurde kryptografisch und sicher gespeichert.',
    'Une erreur est survenue': 'Ein Fehler ist aufgetreten',
  },
  es: {
    'Confirmation de Lecture': 'Confirmación de Lectura',
    'Certifiez votre lecture avec une confirmation cryptographique Ed25519': 'Certifique su lectura con una confirmación criptográfica Ed25519',
    'Chargement du document...': 'Cargando el documento...',
    'Veuillez patienter pendant que nous préparons le document pour la signature.': 'Por favor espere mientras preparamos el documento para la firma.',
    'Aucun document spécifié': 'Ningún documento especificado',
    'Pour signer un document, ajoutez le paramètre {code} à l\'URL': 'Para firmar un documento, agregue el parámetro {code} a la URL',
    'Exemples :': 'Ejemplos:',
    'Lecture confirmée avec succès !': '¡Lectura confirmada con éxito!',
    'Votre confirmation a été enregistrée de manière cryptographique et sécurisée.': 'Su confirmación ha sido registrada de forma criptográfica y segura.',
    'Une erreur est survenue': 'Ha ocurrido un error',
  }
}

// Read source file
const sourcePath = path.join(localesDir, `${sourceLang}.json`)
const sourceContent = JSON.parse(fs.readFileSync(sourcePath, 'utf8'))

// Function to translate value if mapping exists
function translateValue(value, targetLang) {
  if (typeof value !== 'string') return value

  // Check if we have a translation
  const translations = commonTranslations[targetLang] || {}
  if (translations[value]) {
    return translations[value]
  }

  // Return original value (will need manual translation)
  return value
}

// Function to recursively translate object
function translateObject(obj, targetLang) {
  const result = {}
  for (const key in obj) {
    if (typeof obj[key] === 'object' && obj[key] !== null && !Array.isArray(obj[key])) {
      result[key] = translateObject(obj[key], targetLang)
    } else {
      result[key] = translateValue(obj[key], targetLang)
    }
  }
  return result
}

// Process each target language
for (const targetLang of targetLangs) {
  console.log(`Syncing ${targetLang}.json...`)

  const targetPath = path.join(localesDir, `${targetLang}.json`)

  // Translate the entire object
  const translatedContent = translateObject(sourceContent, targetLang)

  // Write to file
  fs.writeFileSync(targetPath, JSON.stringify(translatedContent, null, 2) + '\n', 'utf8')

  console.log(`✓ ${targetLang}.json synchronized`)
}

console.log('\n✓ All translations synchronized!')
console.log('\nNote: This script provides automatic translations for common terms.')
console.log('Many strings still need manual translation by a human translator.')
