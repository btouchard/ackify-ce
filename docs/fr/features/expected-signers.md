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

### Ajout en Batch

```bash
# Liste d'emails dans un fichier
cat emails.txt | while read email; do
  curl -X POST http://localhost:8080/api/v1/admin/documents/policy_2025/signers \
    -b cookies.txt \
    -H "X-CSRF-Token: $CSRF_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$email\"}"
done
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

Pour importer massivement :

```python
import csv
import requests

with open('employees.csv') as f:
    reader = csv.DictReader(f)
    for row in reader:
        requests.post(
            'http://localhost:8080/api/v1/admin/documents/policy_2025/signers',
            json={'email': row['email'], 'name': row['name']},
            headers={'X-CSRF-Token': csrf_token},
            cookies=cookies
        )
```

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
