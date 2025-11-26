# Expected Signers

Tracking des signataires attendus avec rappels email.

## Vue d'Ensemble

La feature "Expected Signers" permet de :
- Définir qui doit signer un document
- Tracker le taux de complétion
- Envoyer des rappels email automatiques
- Détecter les signatures inattendues

## Ajouter des Signataires

### Via le Dashboard Admin

1. Aller sur `/admin`
2. Sélectionner un document
3. Cliquer sur "Expected Signers"
4. Coller la liste d'emails :

```
Alice Smith <alice@company.com>
bob@company.com
charlie@company.com
```

**Formats supportés** :
- Un email par ligne
- Emails séparés par virgules
- Emails séparés par points-virgules
- Format avec nom : `Alice Smith <alice@company.com>`

### Via l'API

```http
POST /api/v1/admin/documents/policy_2025/signers
Content-Type: application/json
X-CSRF-Token: abc123

{
  "email": "alice@company.com",
  "name": "Alice Smith",
  "notes": "Engineering team lead"
}
```

### Import CSV (Recommandé)

La méthode la plus efficace pour ajouter de nombreux signataires est l'import CSV natif.

**Via le Dashboard Admin :**
1. Aller sur `/admin` et sélectionner un document
2. Dans la section "Lecteurs attendus", cliquer sur **Import CSV**
3. Sélectionner un fichier CSV
4. Prévisualiser les entrées (valides, existantes, invalides)
5. Confirmer l'import

**Format CSV supporté :**
```csv
email,name
alice@company.com,Alice Smith
bob@company.com,Bob Jones
charlie@company.com,Charlie Brown
```

**Fonctionnalités auto-détectées :**
- **Séparateur** : virgule (`,`) ou point-virgule (`;`)
- **En-tête** : détection automatique des colonnes `email` et `name`
- **Ordre des colonnes** : flexible (email/name ou name/email)
- **Colonne name** : optionnelle

**Exemples de formats valides :**
```csv
# Avec en-tête, séparateur virgule
email,name
alice@company.com,Alice Smith

# Sans en-tête (email seul)
bob@company.com
charlie@company.com

# En-têtes français, séparateur point-virgule
courriel;nom
alice@company.com;Alice Smith

# Colonnes inversées
name,email
Bob Jones,bob@company.com
```

**Limite configurable :**
```bash
# Par défaut : 500 signataires max par import
ACKIFY_IMPORT_MAX_SIGNERS=1000
```

**Via l'API :**

1. **Preview** (analyse du CSV) :
```http
POST /api/v1/admin/documents/{docId}/signers/preview-csv
Content-Type: multipart/form-data
X-CSRF-Token: abc123

file: [fichier CSV]
```

Response :
```json
{
  "signers": [
    {"lineNumber": 2, "email": "alice@company.com", "name": "Alice Smith"},
    {"lineNumber": 3, "email": "bob@company.com", "name": "Bob Jones"}
  ],
  "errors": [
    {"lineNumber": 5, "content": "invalid-email", "error": "invalid_email_format"}
  ],
  "totalLines": 4,
  "validCount": 2,
  "invalidCount": 1,
  "hasHeader": true,
  "existingEmails": ["charlie@company.com"],
  "maxSigners": 500
}
```

2. **Import** (après validation) :
```http
POST /api/v1/admin/documents/{docId}/signers/import
Content-Type: application/json
X-CSRF-Token: abc123

{
  "signers": [
    {"email": "alice@company.com", "name": "Alice Smith"},
    {"email": "bob@company.com", "name": "Bob Jones"}
  ]
}
```

Response :
```json
{
  "message": "Import completed",
  "imported": 2,
  "skipped": 0,
  "total": 2
}
```

## Tracking de Complétion

### Dashboard Admin

Affiche :
- **Barre de progression** - Visuelle avec pourcentage
- **Liste des signataires** :
  - ✓ Email (signé le DD/MM/YYYY HH:MM)
  - ⏳ Email (en attente)
- **Statistiques** :
  - Expected: 50
  - Signed: 42
  - Pending: 8
  - Completion: 84%

### Via l'API

```http
GET /api/v1/documents/policy_2025/expected-signers
```

**Response** :
```json
{
  "docId": "policy_2025",
  "expectedSigners": [
    {
      "email": "alice@company.com",
      "name": "Alice Smith",
      "addedAt": "2025-01-15T10:00:00Z",
      "hasSigned": true,
      "signedAt": "2025-01-15T14:30:00Z"
    },
    {
      "email": "bob@company.com",
      "name": "Bob Jones",
      "addedAt": "2025-01-15T10:00:00Z",
      "hasSigned": false
    }
  ],
  "completionStats": {
    "expected": 50,
    "signed": 42,
    "pending": 8,
    "completionPercentage": 84.0
  }
}
```

## Rappels Email

### Envoyer des Rappels

**Via le Dashboard** :
1. Sélectionner les destinataires (ou "Select all pending")
2. Choisir la langue (fr, en, es, de, it)
3. Cliquer "Send Reminders"

**Via l'API** :
```http
POST /api/v1/admin/documents/policy_2025/reminders
Content-Type: application/json
X-CSRF-Token: abc123

{
  "emails": ["bob@company.com", "charlie@company.com"],
  "locale": "fr"
}
```

**Response** :
```json
{
  "sent": 2,
  "failed": 0,
  "errors": []
}
```

### Contenu de l'Email

Les templates sont dans `/backend/templates/emails/{locale}/reminder.html` :

```html
Bonjour {{.RecipientName}},

Vous êtes attendu(e) pour signer le document "{{.DocumentTitle}}".

[Bouton: Signer maintenant] → {{.SignURL}}

Document disponible ici : {{.DocumentURL}}

Cordialement,
{{.OrganisationName}}
```

**Variables disponibles** :
- `RecipientName` - Nom du destinataire
- `DocumentTitle` - Titre du document
- `DocumentURL` - URL du document (metadata)
- `SignURL` - Lien direct vers la page de signature
- `OrganisationName` - Nom de votre organisation

### Historique des Rappels

```http
GET /api/v1/admin/documents/policy_2025/reminders
```

**Response** :
```json
{
  "reminders": [
    {
      "recipientEmail": "bob@company.com",
      "sentAt": "2025-01-15T15:00:00Z",
      "sentBy": "admin@company.com",
      "status": "sent",
      "templateUsed": "reminder"
    },
    {
      "recipientEmail": "charlie@company.com",
      "sentAt": "2025-01-15T15:00:05Z",
      "sentBy": "admin@company.com",
      "status": "failed",
      "errorMessage": "SMTP timeout"
    }
  ]
}
```

**Statuts** :
- `sent` - Envoyé avec succès
- `failed` - Échec d'envoi
- `bounced` - Email invalide (bounce)

## Signatures Inattendues

Détecte automatiquement les utilisateurs qui ont signé **sans être attendus**.

### Via le Dashboard

Section "Unexpected Signatures" affiche :
```
⚠️ 3 signatures inattendues détectées
- stranger@external.com (signé le 15/01/2025)
- unknown@gmail.com (signé le 16/01/2025)
```

### Via l'API

Requête SQL pour détecter :
```sql
SELECT s.user_email, s.signed_at
FROM signatures s
LEFT JOIN expected_signers e ON s.user_email = e.email AND s.doc_id = e.doc_id
WHERE s.doc_id = 'policy_2025' AND e.id IS NULL;
```

## Cas d'Usage

### Formation Obligatoire

```
Document: "GDPR Training 2025"
Expected: Tous les employés (CSV import)
```

**Workflow** :
1. Import CSV avec emails employés
2. Envoi du lien de signature à tous
3. Rappel automatique J+7 aux non-signants
4. Export final pour RH

### Politique de Sécurité

```
Document: "Security Policy v3"
Expected: Engineers + DevOps (50 personnes)
```

**Features utilisées** :
- Tracking temps réel (tableau de bord)
- Rappels sélectifs (seulement certains)
- Métadonnées document (URL + checksum)

### Contractuel

```
Document: "NDA 2025"
Expected: Prestataires externes (liste manuelle)
```

**Particularité** :
- Domaine OAuth restreint désactivé
- Permet aux emails externes de signer
- Detection des signatures inattendues cruciale

## Retirer un Signataire

```http
DELETE /api/v1/admin/documents/policy_2025/signers/alice@company.com
X-CSRF-Token: abc123
```

**Comportement** :
- Retire de la liste expected_signers
- La signature (si existante) reste en base
- Le taux de complétion est recalculé

## Configuration Email

Pour que les rappels fonctionnent, configurer SMTP :

```bash
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=noreply@company.com
ACKIFY_MAIL_PASSWORD=app_password
ACKIFY_MAIL_FROM=noreply@company.com
```

Voir [Email Setup](../configuration/email-setup.md) pour plus de détails.

## Bonnes Pratiques

### Import CSV

Utilisez l'import CSV natif (voir section "Import CSV" ci-dessus) pour les imports en masse. Avantages :
- Preview avant import avec détection d'erreurs
- Détection automatique des doublons
- Limite configurable (`ACKIFY_IMPORT_MAX_SIGNERS`)
- Support multi-formats (virgule/point-virgule, avec/sans en-tête)

### Personnalisation

Pour des rappels plus personnalisés :
1. Modifier les templates dans `/backend/templates/emails/`
2. Ajouter des variables custom dans le service email
3. Rebuild l'image Docker

### Monitoring

Surveiller les `reminder_logs` pour détecter :
- Taux de bounce élevé (emails invalides)
- Échecs SMTP répétés
- Efficacité des rappels (taux de conversion)

## Limitations

- Maximum **1000 expected signers** par document (soft limit)
- Rappels envoyés **synchrones** (pas de queue)
- Pas de rappels automatiques planifiés (manuel uniquement)

## API Reference

Voir [API Documentation](../api.md#expected-signers-admin) pour tous les endpoints.
