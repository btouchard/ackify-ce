# Intégration Google Docs/Sheets avec Ackify

Guide d'intégration d'Ackify pour valider la lecture de documents Google via Apps Script.

## 🎯 Principe

Permettre aux utilisateurs de signer/valider qu'ils ont lu un document Google Docs ou Google Sheets directement depuis le document, avec validation cryptographique via Ackify.

## 📋 Prérequis

1. **Document Google Docs/Sheets** publié ou partagé
2. **Instance Ackify** déployée et accessible (ex: `https://ackify.votre-domaine.com`)
3. **Accès Google Apps Script** (compte Google)

## 🚀 Configuration rapide

### Étape 1 : Obtenir l'ID du document

L'ID se trouve dans l'URL de votre document :
```
https://docs.google.com/document/d/[DOCUMENT_ID]/edit
https://docs.google.com/spreadsheets/d/[DOCUMENT_ID]/edit
```

### Étape 2 : Créer le script Apps Script

1. Ouvrir le document Google
2. **Extensions** → **Apps Script**
3. Remplacer le code par défaut par :

```javascript
// Configuration Ackify
const ACKIFY_BASE_URL = 'https://ackify.votre-domaine.com';
const DOCUMENT_ID = 'votre-document-id'; // À remplacer

/**
 * Ajoute un menu Ackify dans Google Docs/Sheets
 */
function onOpen() {
  const ui = DocumentApp.getUi(); // Pour Docs
  // const ui = SpreadsheetApp.getUi(); // Pour Sheets - décommenter si nécessaire
  
  ui.createMenu('📝 Ackify')
    .addItem('✅ Valider ma lecture', 'validateReading')
    .addItem('📊 Voir les validations', 'viewSignatures')
    .addSeparator()
    .addItem('🔗 Intégrer widget', 'showEmbedCode')
    .addToUi();
}

/**
 * Redirige vers la page de validation Ackify
 */
function validateReading() {
  const signUrl = `${ACKIFY_BASE_URL}/sign?doc=${DOCUMENT_ID}&referrer=${encodeURIComponent(getDocumentUrl())}`;
  
  const html = `
    <div style="padding: 20px; font-family: Arial, sans-serif;">
      <h2>🔒 Validation de lecture</h2>
      <p>Cliquez pour valider que vous avez lu ce document :</p>
      <p><a href="${signUrl}" target="_blank" style="
        display: inline-block;
        background: #4285f4;
        color: white;
        padding: 12px 24px;
        text-decoration: none;
        border-radius: 6px;
        font-weight: bold;
      ">✅ Valider ma lecture</a></p>
      <p><small>Une signature cryptographique sera générée pour prouver votre lecture.</small></p>
    </div>
  `;
  
  const htmlOutput = HtmlService.createHtmlOutput(html)
    .setWidth(400)
    .setHeight(200);
  
  const ui = DocumentApp.getUi(); // Pour Docs
  // const ui = SpreadsheetApp.getUi(); // Pour Sheets
  
  ui.showModalDialog(htmlOutput, 'Validation Ackify');
}

/**
 * Affiche les validations existantes
 */
function viewSignatures() {
  const statusUrl = `${ACKIFY_BASE_URL}/api/v1/documents/${DOCUMENT_ID}/signatures`;

  try {
    const response = UrlFetchApp.fetch(statusUrl);
    const result = JSON.parse(response.getContentText());
    const signatures = result.data || [];
    
    let html = `
      <div style="padding: 20px; font-family: Arial, sans-serif;">
        <h2>📊 Validations de lecture</h2>
    `;
    
    if (signatures.length === 0) {
      html += '<p><em>Aucune validation pour ce document.</em></p>';
    } else {
      html += `<p><strong>${signatures.length}</strong> validation(s) :</p><ul>`;
      
      signatures.forEach(sig => {
        const date = new Date(sig.signed_at).toLocaleDateString('fr-FR', {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit'
        });
        
        const name = sig.user_name || sig.user_email;
        html += `<li><strong>${name}</strong> - ${date}</li>`;
      });
      
      html += '</ul>';
    }
    
    html += `
        <p><a href="${statusUrl}" target="_blank">🔗 Voir les détails</a></p>
      </div>
    `;
    
    const htmlOutput = HtmlService.createHtmlOutput(html)
      .setWidth(500)
      .setHeight(400);
    
    const ui = DocumentApp.getUi(); // Pour Docs
    // const ui = SpreadsheetApp.getUi(); // Pour Sheets
    
    ui.showModalDialog(htmlOutput, 'Validations Ackify');
    
  } catch (error) {
    const ui = DocumentApp.getUi();
    ui.alert('Erreur', `Impossible de récupérer les validations : ${error.message}`, ui.ButtonSet.OK);
  }
}

/**
 * Affiche le code d'intégration HTML
 */
function showEmbedCode() {
  const embedCode = `<!-- Widget Ackify -->
<iframe src="${ACKIFY_BASE_URL}/?doc=${DOCUMENT_ID}&referrer=${encodeURIComponent(getDocumentUrl())}"
        width="100%" height="200" frameborder="0"
        style="border: 1px solid #ddd; border-radius: 6px;">
</iframe>`;

  const html = `
    <div style="padding: 20px; font-family: Arial, sans-serif;">
      <h2>🔗 Code d'intégration</h2>
      <p>Copiez ce code HTML pour intégrer le widget Ackify :</p>
      <textarea readonly style="width: 100%; height: 100px; font-family: monospace; font-size: 12px;">${embedCode}</textarea>
      <p><small>À intégrer dans une page web, wiki, ou plateforme supportant l'HTML.</small></p>
    </div>
  `;
  
  const htmlOutput = HtmlService.createHtmlOutput(html)
    .setWidth(600)
    .setHeight(300);
  
  const ui = DocumentApp.getUi(); // Pour Docs
  // const ui = SpreadsheetApp.getUi(); // Pour Sheets
  
  ui.showModalDialog(htmlOutput, 'Code d\'intégration');
}

/**
 * Récupère l'URL du document actuel
 */
function getDocumentUrl() {
  try {
    // Pour Google Docs
    return DocumentApp.getActiveDocument().getUrl();
  } catch (e) {
    try {
      // Pour Google Sheets
      return SpreadsheetApp.getActiveSpreadsheet().getUrl();
    } catch (e2) {
      return `https://docs.google.com/document/d/${DOCUMENT_ID}/edit`;
    }
  }
}
```

### Étape 3 : Configuration du script

1. **Remplacer les variables** :
   ```javascript
   const ACKIFY_BASE_URL = 'https://votre-ackify.com';
   const DOCUMENT_ID = 'votre-id-document-google';
   ```

2. **Pour Google Sheets** : Décommenter les lignes `SpreadsheetApp` et commenter celles de `DocumentApp`

3. **Sauvegarder** le script (Ctrl+S)

### Étape 4 : Autoriser les permissions

1. Cliquer sur **▶️ Exécuter** → `onOpen`
2. **Autoriser** l'accès aux APIs Google (première fois)
3. Recharger le document Google

## ✅ Utilisation

### Menu Ackify

Un nouveau menu **📝 Ackify** apparaît dans votre document avec :

- **✅ Valider ma lecture** : Redirige vers Ackify pour signer
- **📊 Voir les validations** : Liste des signatures existantes  
- **🔗 Intégrer widget** : Code HTML pour intégration externe

### Processus de validation

1. **Utilisateur** clique sur "Valider ma lecture"
2. **Redirection** vers Ackify avec authentification OAuth2
3. **Signature cryptographique** générée (Ed25519)
4. **Retour** au document avec confirmation

## 🔧 Personnalisation avancée

### Notifications automatiques

Ajouter une fonction de notification lors de nouvelles signatures :

```javascript
/**
 * Vérifie périodiquement les nouvelles validations
 */
function checkNewSignatures() {
  // Logique de vérification et notification
  // (à implémenter selon vos besoins)
}

/**
 * Déclenche des vérifications périodiques
 */
function createTrigger() {
  ScriptApp.newTrigger('checkNewSignatures')
    .timeBased()
    .everyHours(1)
    .create();
}
```

### Badge dans le document

Intégrer un badge directement dans le document :

```javascript
/**
 * Insère un lien vers la page de signature Ackify dans le document
 */
function insertSignatureLink() {
  const doc = DocumentApp.getActiveDocument();
  const body = doc.getBody();

  const signUrl = `${ACKIFY_BASE_URL}/?doc=${DOCUMENT_ID}`;

  // Insérer lien de signature
  const paragraph = body.appendParagraph('');
  paragraph.appendText('Signer ce document avec Ackify').setLinkUrl(signUrl);
}
```

## 🛡️ Sécurité

- **Authentification** : OAuth2 requis pour signer
- **Non-répudiation** : Signatures Ed25519 cryptographiquement vérifiables
- **Traçabilité** : Horodatage UTC + hash SHA-256
- **Intégrité** : Chaînage cryptographique des signatures

## 🌐 Intégration multi-plateforme

Le même principe s'applique à d'autres plateformes :

- **Notion** : Via API et webhooks
- **Confluence** : Apps Script ou macros
- **SharePoint** : Power Automate + Custom Connector
- **Wiki** : Widget HTML intégré

## 📞 Support

- **Documentation** : [Ackify GitHub](https://github.com/btouchard/ackify-ce)
- **API** : `GET /api/v1/documents/{docId}/signatures` et `POST /api/v1/signatures`
- **Embed** : Vue SPA gère les embeds via `/?doc=<id>` avec méta tags Open Graph pour l'unfurling automatique
- **Widget** : Utiliser iframe avec `/?doc=<id>` (voir fonction `showEmbedCode()` ci-dessus)

---

**Architecture validée selon CLAUDE.md - Clean Architecture Go 2025** ✨