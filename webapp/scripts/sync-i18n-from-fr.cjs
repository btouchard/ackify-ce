#!/usr/bin/env node

/**
 * Script to sync i18n keys from French to all other languages
 * Uses French as the source of truth for structure
 */

const fs = require('fs');
const path = require('path');

const localesDir = path.join(__dirname, '..', 'src', 'locales');
const frPath = path.join(localesDir, 'fr.json');

// Translation maps from French to other languages
const translations = {
  en: {
    // Signatures page
    'Mes confirmations de lecture': 'My reading confirmations',
    'Liste de tous les documents dont vous avez confirmé la lecture cryptographiquement': 'List of all documents you have cryptographically confirmed reading',
    'résultat': 'result',
    'résultats': 'results',
    'Total': 'Total',
    'Total confirmations': 'Total confirmations',
    'Uniques': 'Unique',
    'Documents uniques': 'Unique documents',
    'Dernier': 'Last',
    'Dernière confirmation': 'Last confirmation',
    'Toutes mes confirmations': 'All my confirmations',
    'À propos des confirmations': 'About confirmations',
    'Chaque confirmation est enregistrée de manière cryptographique avec Ed25519 et chaînée pour garantir l\'intégrité. Les confirmations sont non répudiables et horodatées de façon précise.': 'Each confirmation is cryptographically recorded with Ed25519 and chained to ensure integrity. Confirmations are non-repudiable and precisely timestamped.',
    'Rechercher...': 'Search...',

    // Admin page
    'Gérer les documents et les lecteurs attendus': 'Manage documents and expected readers',
    'Chargement des données...': 'Loading data...',
    'Documents': 'Documents',
    'Lecteurs': 'Readers',
    'Actifs': 'Active',
    'Attendus': 'Expected',
    'Signés': 'Signed',
    'En attente': 'Pending',
    'Complétude': 'Completion',
    'Tous les documents': 'All documents',
    'Créer un nouveau document': 'Create new document',
    'Préparer la référence d\'un document pour suivre les confirmations de lecture': 'Prepare a document reference to track reading confirmations',
    'Rechercher par ID, titre ou URL...': 'Search by ID, title or URL...',
    'ID du document': 'Document ID',
    'Lettres, chiffres, tirets et underscores uniquement': 'Letters, numbers, hyphens and underscores only',
    'ex: politique-securite-2025': 'eg: security-policy-2025',
    'Créé le': 'Created on',
    'Par': 'By',
    'URL': 'URL',
    'Document': 'Document',
    'Gérer': 'Manage',

    // Document detail
    'Métadonnées et checksum du document': 'Document metadata and checksum',
    'Titre': 'Title',
    'Politique de sécurité 2025': 'Security Policy 2025',
    'Description': 'Description',
    'Description du document...': 'Document description...',
    'Signataires attendus': 'Expected signers',
    'Ajouter un signataire attendu': 'Add expected signer',
    'Ajouter des lecteurs attendus': 'Add expected readers',
    'Emails (un par ligne)': 'Emails (one per line)',
    'Email *': 'Email *',
    'email@example.com': 'email@example.com',
    'Nom': 'Name',
    'Nom complet': 'Full name',
    'Lecteur': 'Reader',
    'Utilisateur': 'User',
    'Statut': 'Status',
    'Confirmé le': 'Confirmed on',
    'Aucun lecteur attendu': 'No expected readers',
    'Aucune confirmation': 'No confirmations',
    'Relances email': 'Email reminders',
    'Relances envoyées': 'Reminders sent',
    'À relancer': 'To remind',
    'Dernière relance': 'Last reminder',
    'Envoyer une relance': 'Send reminder',
    'Zone de danger': 'Danger zone',
    'Actions irréversibles sur ce document': 'Irreversible actions on this document',
    'Supprimer ce document': 'Delete this document',
    'Cette action est irréversible !': 'This action is irreversible!',
    'Cette action supprimera définitivement :': 'This action will permanently delete:',
    'Toutes les métadonnées du document': 'All document metadata',
    'La liste des lecteurs attendus': 'The list of expected readers',
    'Toutes les confirmations cryptographiques': 'All cryptographic confirmations',
    'L\'historique des relances': 'The reminder history'
  },
  es: {
    'Mes confirmations de lecture': 'Mis confirmaciones de lectura',
    'Liste de tous les documents dont vous avez confirmé la lecture cryptographiquement': 'Lista de todos los documentos cuya lectura has confirmado criptográficamente',
    'résultat': 'resultado',
    'résultats': 'resultados',
    'Total': 'Total',
    'Total confirmations': 'Total confirmaciones',
    'Uniques': 'Únicos',
    'Documents uniques': 'Documentos únicos',
    'Dernier': 'Último',
    'Dernière confirmation': 'Última confirmación',
    'Toutes mes confirmations': 'Todas mis confirmaciones',
    'À propos des confirmations': 'Acerca de las confirmaciones',
    'Chaque confirmation est enregistrée de manière cryptographique avec Ed25519 et chaînée pour garantir l\'intégrité. Les confirmations sont non répudiables et horodatées de façon précise.': 'Cada confirmación se registra criptográficamente con Ed25519 y se encadena para garantizar la integridad. Las confirmaciones son irrefutables y tienen una marca de tiempo precisa.',
    'Rechercher...': 'Buscar...',

    'Gérer les documents et les lecteurs attendus': 'Gestionar documentos y lectores esperados',
    'Chargement des données...': 'Cargando datos...',
    'Documents': 'Documentos',
    'Lecteurs': 'Lectores',
    'Actifs': 'Activos',
    'Attendus': 'Esperados',
    'Signés': 'Firmados',
    'En attente': 'Pendientes',
    'Complétude': 'Completitud',
    'Tous les documents': 'Todos los documentos',
    'Créer un nouveau document': 'Crear nuevo documento',
    'Préparer la référence d\'un document pour suivre les confirmations de lecture': 'Preparar una referencia de documento para seguir las confirmaciones de lectura',
    'Rechercher par ID, titre ou URL...': 'Buscar por ID, título o URL...',
    'ID du document': 'ID del documento',
    'Lettres, chiffres, tirets et underscores uniquement': 'Solo letras, números, guiones y guiones bajos',
    'ex: politique-securite-2025': 'ej: politica-seguridad-2025',
    'Créé le': 'Creado el',
    'Par': 'Por',
    'URL': 'URL',
    'Document': 'Documento',
    'Gérer': 'Gestionar',

    'Métadonnées et checksum du document': 'Metadatos y checksum del documento',
    'Titre': 'Título',
    'Politique de sécurité 2025': 'Política de seguridad 2025',
    'Description': 'Descripción',
    'Description du document...': 'Descripción del documento...',
    'Signataires attendus': 'Firmantes esperados',
    'Ajouter un signataire attendu': 'Añadir firmante esperado',
    'Ajouter des lecteurs attendus': 'Añadir lectores esperados',
    'Emails (un par ligne)': 'Emails (uno por línea)',
    'Email *': 'Email *',
    'email@example.com': 'email@example.com',
    'Nom': 'Nombre',
    'Nom complet': 'Nombre completo',
    'Lecteur': 'Lector',
    'Utilisateur': 'Usuario',
    'Statut': 'Estado',
    'Confirmé le': 'Confirmado el',
    'Aucun lecteur attendu': 'Ningún lector esperado',
    'Aucune confirmation': 'Ninguna confirmación',
    'Relances email': 'Recordatorios por email',
    'Relances envoyées': 'Recordatorios enviados',
    'À relancer': 'Para recordar',
    'Dernière relance': 'Último recordatorio',
    'Envoyer une relance': 'Enviar recordatorio',
    'Zone de danger': 'Zona de peligro',
    'Actions irréversibles sur ce document': 'Acciones irreversibles sobre este documento',
    'Supprimer ce document': 'Eliminar este documento',
    'Cette action est irréversible !': '¡Esta acción es irreversible!',
    'Cette action supprimera définitivement :': 'Esta acción eliminará permanentemente:',
    'Toutes les métadonnées du document': 'Todos los metadatos del documento',
    'La liste des lecteurs attendus': 'La lista de lectores esperados',
    'Toutes les confirmations cryptographiques': 'Todas las confirmaciones criptográficas',
    'L\'historique des relances': 'El historial de recordatorios'
  },
  de: {
    'Mes confirmations de lecture': 'Meine Lesebestätigungen',
    'Liste de tous les documents dont vous avez confirmé la lecture cryptographiquement': 'Liste aller Dokumente, deren Lektüre Sie kryptografisch bestätigt haben',
    'résultat': 'Ergebnis',
    'résultats': 'Ergebnisse',
    'Total': 'Gesamt',
    'Total confirmations': 'Gesamtbestätigungen',
    'Uniques': 'Einzigartig',
    'Documents uniques': 'Einzigartige Dokumente',
    'Dernier': 'Letzte',
    'Dernière confirmation': 'Letzte Bestätigung',
    'Toutes mes confirmations': 'Alle meine Bestätigungen',
    'À propos des confirmations': 'Über Bestätigungen',
    'Chaque confirmation est enregistrée de manière cryptographique avec Ed25519 et chaînée pour garantir l\'intégrité. Les confirmations sont non répudiables et horodatées de façon précise.': 'Jede Bestätigung wird kryptografisch mit Ed25519 aufgezeichnet und verkettet, um die Integrität zu gewährleisten. Bestätigungen sind unwiderruflich und präzise mit Zeitstempel versehen.',
    'Rechercher...': 'Suchen...',

    'Gérer les documents et les lecteurs attendus': 'Dokumente und erwartete Leser verwalten',
    'Chargement des données...': 'Daten werden geladen...',
    'Documents': 'Dokumente',
    'Lecteurs': 'Leser',
    'Actifs': 'Aktiv',
    'Attendus': 'Erwartet',
    'Signés': 'Signiert',
    'En attente': 'Ausstehend',
    'Complétude': 'Vollständigkeit',
    'Tous les documents': 'Alle Dokumente',
    'Créer un nouveau document': 'Neues Dokument erstellen',
    'Préparer la référence d\'un document pour suivre les confirmations de lecture': 'Dokumentreferenz vorbereiten, um Lesebestätigungen zu verfolgen',
    'Rechercher par ID, titre ou URL...': 'Nach ID, Titel oder URL suchen...',
    'ID du document': 'Dokument-ID',
    'Lettres, chiffres, tirets et underscores uniquement': 'Nur Buchstaben, Zahlen, Bindestriche und Unterstriche',
    'ex: politique-securite-2025': 'z.B.: sicherheitsrichtlinie-2025',
    'Créé le': 'Erstellt am',
    'Par': 'Von',
    'URL': 'URL',
    'Document': 'Dokument',
    'Gérer': 'Verwalten',

    'Métadonnées et checksum du document': 'Dokumentmetadaten und Prüfsumme',
    'Titre': 'Titel',
    'Politique de sécurité 2025': 'Sicherheitsrichtlinie 2025',
    'Description': 'Beschreibung',
    'Description du document...': 'Dokumentbeschreibung...',
    'Signataires attendus': 'Erwartete Unterzeichner',
    'Ajouter un signataire attendu': 'Erwarteten Unterzeichner hinzufügen',
    'Ajouter des lecteurs attendus': 'Erwartete Leser hinzufügen',
    'Emails (un par ligne)': 'E-Mails (eine pro Zeile)',
    'Email *': 'E-Mail *',
    'email@example.com': 'email@example.com',
    'Nom': 'Name',
    'Nom complet': 'Vollständiger Name',
    'Lecteur': 'Leser',
    'Utilisateur': 'Benutzer',
    'Statut': 'Status',
    'Confirmé le': 'Bestätigt am',
    'Aucun lecteur attendu': 'Keine erwarteten Leser',
    'Aucune confirmation': 'Keine Bestätigungen',
    'Relances email': 'E-Mail-Erinnerungen',
    'Relances envoyées': 'Gesendete Erinnerungen',
    'À relancer': 'Zu erinnern',
    'Dernière relance': 'Letzte Erinnerung',
    'Envoyer une relance': 'Erinnerung senden',
    'Zone de danger': 'Gefahrenzone',
    'Actions irréversibles sur ce document': 'Irreversible Aktionen für dieses Dokument',
    'Supprimer ce document': 'Dieses Dokument löschen',
    'Cette action est irréversible !': 'Diese Aktion ist irreversibel!',
    'Cette action supprimera définitivement :': 'Diese Aktion wird dauerhaft löschen:',
    'Toutes les métadonnées du document': 'Alle Dokumentmetadaten',
    'La liste des lecteurs attendus': 'Die Liste der erwarteten Leser',
    'Toutes les confirmations cryptographiques': 'Alle kryptografischen Bestätigungen',
    'L\'historique des relances': 'Der Erinnerungsverlauf'
  },
  it: {
    'Mes confirmations de lecture': 'Le mie conferme di lettura',
    'Liste de tous les documents dont vous avez confirmé la lecture cryptographiquement': 'Elenco di tutti i documenti di cui hai confermato la lettura crittograficamente',
    'résultat': 'risultato',
    'résultats': 'risultati',
    'Total': 'Totale',
    'Total confirmations': 'Conferme totali',
    'Uniques': 'Unici',
    'Documents uniques': 'Documenti unici',
    'Dernier': 'Ultimo',
    'Dernière confirmation': 'Ultima conferma',
    'Toutes mes confirmations': 'Tutte le mie conferme',
    'À propos des confirmations': 'Informazioni sulle conferme',
    'Chaque confirmation est enregistrée de manière cryptographique avec Ed25519 et chaînée pour garantir l\'intégrité. Les confirmations sont non répudiables et horodatées de façon précise.': 'Ogni conferma viene registrata crittograficamente con Ed25519 e concatenata per garantire l\'integrità. Le conferme sono irrevocabili e timestampate in modo preciso.',
    'Rechercher...': 'Cerca...',

    'Gérer les documents et les lecteurs attendus': 'Gestire documenti e lettori previsti',
    'Chargement des données...': 'Caricamento dati...',
    'Documents': 'Documenti',
    'Lecteurs': 'Lettori',
    'Actifs': 'Attivi',
    'Attendus': 'Previsti',
    'Signés': 'Firmati',
    'En attente': 'In attesa',
    'Complétude': 'Completamento',
    'Tous les documents': 'Tutti i documenti',
    'Créer un nouveau document': 'Crea nuovo documento',
    'Préparer la référence d\'un document pour suivre les confirmations de lecture': 'Preparare un riferimento documento per tracciare le conferme di lettura',
    'Rechercher par ID, titre ou URL...': 'Cerca per ID, titolo o URL...',
    'ID du document': 'ID documento',
    'Lettres, chiffres, tirets et underscores uniquement': 'Solo lettere, numeri, trattini e underscore',
    'ex: politique-securite-2025': 'es: politica-sicurezza-2025',
    'Créé le': 'Creato il',
    'Par': 'Da',
    'URL': 'URL',
    'Document': 'Documento',
    'Gérer': 'Gestisci',

    'Métadonnées et checksum du document': 'Metadati e checksum del documento',
    'Titre': 'Titolo',
    'Politique de sécurité 2025': 'Politica di sicurezza 2025',
    'Description': 'Descrizione',
    'Description du document...': 'Descrizione del documento...',
    'Signataires attendus': 'Firmatari previsti',
    'Ajouter un signataire attendu': 'Aggiungi firmatario previsto',
    'Ajouter des lecteurs attendus': 'Aggiungi lettori previsti',
    'Emails (un par ligne)': 'Email (una per riga)',
    'Email *': 'Email *',
    'email@example.com': 'email@example.com',
    'Nom': 'Nome',
    'Nom complet': 'Nome completo',
    'Lecteur': 'Lettore',
    'Utilisateur': 'Utente',
    'Statut': 'Stato',
    'Confirmé le': 'Confermato il',
    'Aucun lecteur attendu': 'Nessun lettore previsto',
    'Aucune confirmation': 'Nessuna conferma',
    'Relances email': 'Promemoria email',
    'Relances envoyées': 'Promemoria inviati',
    'À relancer': 'Da ricordare',
    'Dernière relance': 'Ultimo promemoria',
    'Envoyer une relance': 'Invia promemoria',
    'Zone de danger': 'Zona di pericolo',
    'Actions irréversibles sur ce document': 'Azioni irreversibili su questo documento',
    'Supprimer ce document': 'Elimina questo documento',
    'Cette action est irréversible !': 'Questa azione è irreversibile!',
    'Cette action supprimera définitivement :': 'Questa azione eliminerà permanentemente:',
    'Toutes les métadonnées du document': 'Tutti i metadati del documento',
    'La liste des lecteurs attendus': 'L\'elenco dei lettori previsti',
    'Toutes les confirmations cryptographiques': 'Tutte le conferme crittografiche',
    'L\'historique des relances': 'La cronologia dei promemoria'
  }
};

function translateValue(value, lang, map) {
  if (typeof value === 'string') {
    return map[value] || value;
  }
  if (typeof value === 'object' && value !== null) {
    const result = {};
    for (const [k, v] of Object.entries(value)) {
      result[k] = translateValue(v, lang, map);
    }
    return result;
  }
  return value;
}

function syncLocale(targetLang) {
  const targetPath = path.join(localesDir, `${targetLang}.json`);
  const frData = JSON.parse(fs.readFileSync(frPath, 'utf8'));
  const targetData = JSON.parse(fs.readFileSync(targetPath, 'utf8'));

  const translationMap = translations[targetLang];
  const synced = translateValue(frData, targetLang, translationMap);

  // Write back
  fs.writeFileSync(targetPath, JSON.stringify(synced, null, 2) + '\n', 'utf8');
  console.log(`✅ Synced ${targetLang}.json`);
}

// Sync all languages
['en', 'es', 'de', 'it'].forEach(syncLocale);

console.log('\n✨ All locales synced from fr.json');
