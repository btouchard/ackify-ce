# Email Setup

Configuration SMTP pour l'envoi de rappels email aux signataires attendus.

## Vue d'Ensemble

Le service email d'Ackify permet d'envoyer des rappels automatiques aux utilisateurs qui n'ont pas encore signé un document.

**Fonctionnalités** :
- Envoi de rappels multilingues (fr, en, es, de, it)
- Templates HTML et texte brut
- Historique des envois dans PostgreSQL
- Support TLS/STARTTLS
- Timeout configurable

**Note** : Le service email est **optionnel**. Si `ACKIFY_MAIL_HOST` n'est pas défini, les emails sont désactivés.

## Configuration de Base

### Variables Obligatoires

```bash
# Serveur SMTP
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=your-email@gmail.com
ACKIFY_MAIL_PASSWORD=your-app-password

# Adresse expéditeur
ACKIFY_MAIL_FROM=noreply@company.com
```

### Variables Optionnelles

```bash
# Nom affiché de l'expéditeur (défaut: ACKIFY_ORGANISATION)
ACKIFY_MAIL_FROM_NAME="Ackify - ACME Corporation"

# Préfixe pour le sujet des emails (optionnel)
ACKIFY_MAIL_SUBJECT_PREFIX="[Ackify]"

# Activer TLS (défaut: true)
ACKIFY_MAIL_TLS=true

# Activer STARTTLS (défaut: true)
ACKIFY_MAIL_STARTTLS=true

# Désactiver la vérification des certificats TLS (défaut: false)
# Utile pour les certificats auto-signés en développement/test
# /!\ NE PAS UTILISER EN PRODUCTION
ACKIFY_MAIL_INSECURE_SKIP_VERIFY=false

# Timeout de connexion (défaut: 10s)
ACKIFY_MAIL_TIMEOUT=10s

# Répertoire des templates email (défaut: templates/emails)
ACKIFY_MAIL_TEMPLATE_DIR=templates/emails

# Langue par défaut pour les emails (défaut: en)
# Langues supportées : en, fr, es, de, it
ACKIFY_MAIL_DEFAULT_LOCALE=fr
```

## Providers SMTP Populaires

### Gmail

**Configuration** :
```bash
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=your-email@gmail.com
ACKIFY_MAIL_PASSWORD=your-app-password
ACKIFY_MAIL_TLS=true
ACKIFY_MAIL_STARTTLS=true
```

**Prérequis** :
1. Activer la validation en 2 étapes sur votre compte Google
2. Générer un "App Password" : https://myaccount.google.com/apppasswords
3. Utiliser ce mot de passe dans `ACKIFY_MAIL_PASSWORD`

### SendGrid

```bash
ACKIFY_MAIL_HOST=smtp.sendgrid.net
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=apikey
ACKIFY_MAIL_PASSWORD=your-sendgrid-api-key
ACKIFY_MAIL_FROM=noreply@your-domain.com
ACKIFY_MAIL_TLS=true
```

### Amazon SES

```bash
ACKIFY_MAIL_HOST=email-smtp.us-east-1.amazonaws.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=your-smtp-username
ACKIFY_MAIL_PASSWORD=your-smtp-password
ACKIFY_MAIL_FROM=noreply@verified-domain.com
ACKIFY_MAIL_TLS=true
```

**Important** : Vérifier votre domaine dans AWS SES avant d'envoyer.

### Mailgun

```bash
ACKIFY_MAIL_HOST=smtp.mailgun.org
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=postmaster@your-domain.mailgun.org
ACKIFY_MAIL_PASSWORD=your-mailgun-smtp-password
ACKIFY_MAIL_FROM=noreply@your-domain.com
ACKIFY_MAIL_TLS=true
```

### SMTP Custom (Self-hosted)

```bash
ACKIFY_MAIL_HOST=mail.company.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=ackify@company.com
ACKIFY_MAIL_PASSWORD=secure_password
ACKIFY_MAIL_FROM=ackify@company.com
ACKIFY_MAIL_TLS=true
ACKIFY_MAIL_STARTTLS=true
# Pour certificats auto-signés uniquement (/!\ pas en production)
# ACKIFY_MAIL_INSECURE_SKIP_VERIFY=true
```

## Templates Email

Les templates sont dans `/backend/templates/emails/` avec support multilingue.

### Structure

```
templates/emails/
├── fr/
│   ├── reminder.html      # Template HTML français
│   └── reminder.txt       # Template texte brut français
├── en/
│   ├── reminder.html      # Template HTML anglais
│   └── reminder.txt       # Template texte brut anglais
└── ...
```

### Variables Disponibles

Dans les templates, vous pouvez utiliser :

```go
{{.RecipientName}}     // Nom du destinataire
{{.DocumentID}}        // ID du document
{{.DocumentTitle}}     // Titre du document
{{.DocumentURL}}       // URL du document (si définie dans metadata)
{{.SignURL}}           // URL pour signer
{{.OrganisationName}}  // Nom de l'organisation
{{.SenderName}}        // Nom de l'expéditeur (admin)
```

### Exemple de Template HTML

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Rappel de signature</title>
</head>
<body>
  <h1>Bonjour {{.RecipientName}},</h1>
  <p>
    Vous êtes attendu(e) pour signer le document
    <strong>{{.DocumentTitle}}</strong>.
  </p>
  <p>
    <a href="{{.SignURL}}" style="background: #0066cc; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">
      Signer maintenant
    </a>
  </p>
  <p>
    Cordialement,<br>
    {{.OrganisationName}}
  </p>
</body>
</html>
```

### Personnaliser les Templates

Pour utiliser des templates personnalisés :

```bash
ACKIFY_MAIL_TEMPLATE_DIR=/custom/path/to/email/templates
```

Assurez-vous de conserver la même structure de répertoires (langue/reminder.html).

## Envoi de Rappels

### Via le Dashboard Admin

1. Aller sur `/admin`
2. Sélectionner un document
3. Cliquer sur "Expected Signers"
4. Sélectionner les destinataires
5. Cliquer sur "Send Reminders"

### Via l'API

```bash
curl -X POST http://localhost:8080/api/v1/admin/documents/doc_id/reminders \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_TOKEN" \
  -d '{
    "emails": ["user1@company.com", "user2@company.com"],
    "locale": "fr"
  }'
```

**Réponse** :
```json
{
  "sent": 2,
  "failed": 0,
  "errors": []
}
```

## Historique des Rappels

Les envois sont tracés dans la table `reminder_logs` :

```sql
SELECT
  recipient_email,
  sent_at,
  status,
  error_message
FROM reminder_logs
WHERE doc_id = 'my_document'
ORDER BY sent_at DESC;
```

**Statuts possibles** :
- `sent` - Envoyé avec succès
- `failed` - Échec d'envoi
- `bounced` - Rebond (email invalide)

## Tester la Configuration

### Test Manuel via API

```bash
# 1. Se connecter en tant qu'admin
# 2. Ajouter un expected signer avec votre email
curl -X POST http://localhost:8080/api/v1/admin/documents/test_doc/signers \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_TOKEN" \
  -d '{
    "email": "your-email@company.com",
    "name": "Test User"
  }'

# 3. Envoyer un rappel test
curl -X POST http://localhost:8080/api/v1/admin/documents/test_doc/reminders \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_TOKEN" \
  -d '{
    "emails": ["your-email@company.com"],
    "locale": "en"
  }'
```

### Vérifier les Logs

```bash
docker compose logs -f ackify-ce | grep -i mail
```

Vous devriez voir :
```
INFO  Email sent successfully to: your-email@company.com
```

## Troubleshooting

### Erreur "SMTP connection failed"

Vérifier :
- `ACKIFY_MAIL_HOST` et `ACKIFY_MAIL_PORT` sont corrects
- Votre serveur autorise les connexions sortantes sur le port SMTP
- `ACKIFY_MAIL_TLS=true` si le serveur requiert TLS

### Erreur "tls: failed to verify certificate: x509: certificate signed by unknown authority"

Cette erreur se produit avec des certificats auto-signés. **Pour les environnements de développement/test uniquement** :

```bash
ACKIFY_MAIL_INSECURE_SKIP_VERIFY=true
```

/!\ **Attention** : Cette option désactive la vérification des certificats TLS. Ne JAMAIS l'utiliser en production !

### Erreur "Authentication failed"

Vérifier :
- `ACKIFY_MAIL_USERNAME` et `ACKIFY_MAIL_PASSWORD` sont corrects
- Pour Gmail : utiliser un "App Password", pas votre mot de passe principal
- Pour SendGrid : le username doit être `apikey`

### Email non reçu mais status "sent"

Vérifier :
- Dossier spam/courrier indésirable
- SPF/DKIM/DMARC de votre domaine (pour éviter les filtres anti-spam)
- L'adresse `ACKIFY_MAIL_FROM` est vérifiée chez votre provider

### Template non trouvé

Vérifier :
- `ACKIFY_MAIL_TEMPLATE_DIR` pointe vers le bon répertoire
- La structure `{locale}/reminder.html` existe
- Les fichiers ont les bonnes permissions (readable)

### Timeout lors de l'envoi

Augmenter le timeout :
```bash
ACKIFY_MAIL_TIMEOUT=30s
```

## Bonnes Pratiques

### Production

- ✅ Utiliser un service SMTP dédié (SendGrid, Mailgun, SES)
- ✅ Vérifier votre domaine (SPF, DKIM, DMARC)
- ✅ Utiliser une adresse `noreply@` pour `ACKIFY_MAIL_FROM`
- ✅ Monitorer les `reminder_logs` pour détecter les échecs
- ✅ Tester régulièrement l'envoi d'emails

### Sécurité

- ✅ Ne jamais commiter `ACKIFY_MAIL_PASSWORD` dans git
- ✅ Utiliser des secrets Docker ou variables d'environnement
- ✅ Restreindre les permissions du compte SMTP
- ✅ Activer TLS/STARTTLS en production

### Performance

- Les emails sont envoyés **synchrones** lors de l'appel API
- Pour de gros volumes, envisager une queue asynchrone
- Limiter le nombre de destinataires par batch (recommandé : < 100)

## Désactiver les Emails

Pour désactiver complètement le service email :

```bash
# Supprimer ou commenter ACKIFY_MAIL_HOST
# ACKIFY_MAIL_HOST=
```

Le dashboard admin n'affichera plus les options d'envoi de rappels.
