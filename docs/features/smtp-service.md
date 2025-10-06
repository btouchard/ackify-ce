# Guide d'utilisation – Service SMTP

## 📧 Vue d'ensemble

Le service SMTP d'Ackify permet d'envoyer des emails de rappel de signature aux utilisateurs. Il supporte :
- Templates multilingues (HTML + texte)
- Configuration complète via variables d'environnement
- Désactivation automatique si non configuré (pas d'erreur)
- Support TLS/STARTTLS
- Templates personnalisables

## ⚙️ Configuration

### Variables d'environnement

| Variable | Type | Défaut | Description |
|----------|------|--------|-------------|
| `ACKIFY_MAIL_HOST` | string | - | **Obligatoire** : Hôte SMTP (ex: smtp.gmail.com) |
| `ACKIFY_MAIL_PORT` | int | `587` | Port SMTP |
| `ACKIFY_MAIL_USERNAME` | string | - | Identifiant SMTP (optionnel si auth non requise) |
| `ACKIFY_MAIL_PASSWORD` | string | - | Mot de passe SMTP |
| `ACKIFY_MAIL_TLS` | bool | `true` | Activer TLS implicite (port 465) |
| `ACKIFY_MAIL_STARTTLS` | bool | `true` | Activer STARTTLS (port 587) |
| `ACKIFY_MAIL_TIMEOUT` | duration | `10s` | Timeout de connexion |
| `ACKIFY_MAIL_FROM` | string | - | **Obligatoire** : Adresse expéditeur |
| `ACKIFY_MAIL_FROM_NAME` | string | `ACKIFY_ORGANISATION` | Nom expéditeur |
| `ACKIFY_MAIL_SUBJECT_PREFIX` | string | `""` | Préfixe ajouté aux sujets |
| `ACKIFY_MAIL_TEMPLATE_DIR` | path | `templates/emails` | Répertoire des templates |
| `ACKIFY_MAIL_DEFAULT_LOCALE` | string | `en` | Locale par défaut (en/fr) |

### Exemple de configuration

**.env (développement avec MailHog)** :
```bash
ACKIFY_MAIL_HOST=localhost
ACKIFY_MAIL_PORT=1025
ACKIFY_MAIL_FROM=noreply@ackify.local
ACKIFY_MAIL_FROM_NAME=Ackify CE
```

**.env (production Gmail)** :
```bash
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=your-email@gmail.com
ACKIFY_MAIL_PASSWORD=your-app-password
ACKIFY_MAIL_TLS=false
ACKIFY_MAIL_STARTTLS=true
ACKIFY_MAIL_FROM=noreply@yourdomain.com
ACKIFY_MAIL_FROM_NAME="Ackify - Proof of Read"
ACKIFY_MAIL_SUBJECT_PREFIX="[Ackify] "
```

### Désactivation

Si `ACKIFY_MAIL_HOST` n'est pas défini, le service est **automatiquement désactivé** sans erreur. Les appels d'envoi d'email retournent `nil` avec un log informatif.

## 📝 Utilisation dans le code

### Initialisation

```go
import (
    "github.com/btouchard/ackify-ce/internal/infrastructure/config"
    "github.com/btouchard/ackify-ce/internal/infrastructure/email"
)

// Charger config
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Créer renderer et sender
renderer := email.NewRenderer(
    cfg.Mail.TemplateDir,
    cfg.App.BaseURL,
    cfg.App.Organisation,
    cfg.Mail.FromName,
    cfg.Mail.From,
    cfg.Mail.DefaultLocale,
)

sender := email.NewSMTPSender(cfg.Mail, renderer)
```

### Envoyer un rappel de signature

```go
import (
    "context"
    "github.com/btouchard/ackify-ce/internal/infrastructure/email"
)

ctx := context.Background()

err := email.SendSignatureReminderEmail(
    ctx,
    sender,
    []string{"user@example.com"},
    "fr", // ou "en"
    "doc_123abc",
    "https://example.com/documents/doc_123abc",
    "https://example.com/sign?doc=doc_123abc",
)

if err != nil {
    log.Printf("Failed to send reminder: %v", err)
}
```

### Envoyer un email personnalisé

```go
data := map[string]any{
    "UserName": "John Doe",
    "CustomField": "custom value",
}

err := email.SendEmail(
    ctx,
    sender,
    "custom_template", // nom du template (sans extension)
    []string{"user@example.com"},
    "en",
    "Your Custom Subject",
    data,
)
```

## 🎨 Créer des templates personnalisés

### Structure des templates

Les templates utilisent le système de `html/template` et `text/template` de Go.

**Répertoire** : `/templates/emails/`

**Fichiers requis** :
- `base.html.tmpl` - Template de base HTML
- `base.txt.tmpl` - Template de base texte
- `<nom>.en.html.tmpl` - Version anglaise HTML
- `<nom>.en.txt.tmpl` - Version anglaise texte
- `<nom>.fr.html.tmpl` - Version française HTML
- `<nom>.fr.txt.tmpl` - Version française texte

### Variables automatiques

Chaque template reçoit automatiquement :
- `.Organisation` - Nom de l'organisation (depuis config)
- `.BaseURL` - URL de base de l'application
- `.FromName` - Nom de l'expéditeur
- `.FromMail` - Email de l'expéditeur
- `.Data.*` - Vos données personnalisées

### Exemple : Template de rappel de signature

**signature_reminder.en.html.tmpl** :
```html
{{define "content"}}
<h2>Document Signature Reminder</h2>

<p>Hello,</p>

<p>The following document requires your signature:</p>

<div style="background-color: #f3f4f6; padding: 15px;">
    <p><strong>Document ID:</strong> {{.Data.DocID}}</p>
</div>

<p>To sign: <a href="{{.Data.SignURL}}">Click here</a></p>

<p>Best regards,<br>
The {{.Organisation}} Team</p>
{{end}}
```

**signature_reminder.en.txt.tmpl** :
```
{{define "content"}}
Document Signature Reminder

Hello,

The following document requires your signature:
Document ID: {{.Data.DocID}}

To sign, visit: {{.Data.SignURL}}

Best regards,
The {{.Organisation}} Team
{{end}}
```

### Résolution des templates

Le système résout les templates dans cet ordre :
1. `<nom>.<locale>.html.tmpl` (ex: `welcome.fr.html.tmpl`)
2. `<nom>.en.html.tmpl` (fallback anglais)
3. Erreur si aucun template trouvé

## 🧪 Tests locaux avec MailHog

MailHog est inclus dans `compose.local.yml` pour tester l'envoi d'emails.

### Lancement

```bash
docker compose -f compose.local.yml up -d mailhog
```

### Interface web

Accédez à http://localhost:8025 pour voir les emails envoyés.

### Configuration

```bash
ACKIFY_MAIL_HOST=mailhog
ACKIFY_MAIL_PORT=1025
ACKIFY_MAIL_FROM=test@ackify.local
```

## 🔍 Troubleshooting

### Email non envoyé

**Problème** : Aucun email n'est envoyé, pas d'erreur.

**Solution** : Vérifiez que `ACKIFY_MAIL_HOST` est défini. Si non défini, le service est désactivé silencieusement.

### Erreur "failed to send email"

**Problème** : Erreur lors de l'envoi.

**Solutions** :
- Vérifiez les credentials SMTP (`ACKIFY_MAIL_USERNAME`, `ACKIFY_MAIL_PASSWORD`)
- Vérifiez le port et TLS/STARTTLS
- Pour Gmail, utilisez un "App Password" (pas votre mot de passe principal)

### Template non trouvé

**Problème** : `template not found: <name> (locale: <locale>)`

**Solutions** :
- Vérifiez que le template existe dans `ACKIFY_MAIL_TEMPLATE_DIR`
- Vérifiez le nom du fichier : `<nom>.<locale>.<html|txt>.tmpl`
- Au minimum, créez la version anglaise `.en.html.tmpl` et `.en.txt.tmpl`

### Secrets dans les logs

**Problème** : Mot de passe SMTP dans les logs.

**Solution** : Le système ne logue **jamais** les secrets. Si vous voyez des secrets, c'est un bug à signaler.

## 📊 Monitoring

Le service logue automatiquement :
- `INFO` : "SMTP not configured, email not sent" (si désactivé)
- `INFO` : "Sending email" avec destinataires, template, locale
- `INFO` : "Email sent successfully" avec destinataires
- `ERROR` : Erreurs de rendu ou d'envoi

Exemple :
```
{"level":"INFO","msg":"Sending email","to":["user@example.com"],"template":"signature_reminder","locale":"fr"}
{"level":"INFO","msg":"Email sent successfully","to":["user@example.com"]}
```

## 🔐 Sécurité

- ✅ Aucun secret (password, credentials) n'est loggé
- ✅ TLS/STARTTLS supporté pour chiffrement
- ✅ Timeout pour éviter les blocages
- ✅ Service désactivé par défaut (opt-in explicite)

## 🚀 Intégration dans les handlers

Exemple d'utilisation dans un handler :

```go
func (h *SignatureHandlers) SendReminder(w http.ResponseWriter, r *http.Request) {
    docID := r.URL.Query().Get("doc")
    userEmail := getUserEmail(r) // votre logique

    docURL := fmt.Sprintf("%s/status?doc=%s", h.baseURL, docID)
    signURL := fmt.Sprintf("%s/sign?doc=%s", h.baseURL, docID)

    locale := getLocaleFromRequest(r) // "en" ou "fr"

    err := email.SendSignatureReminderEmail(
        r.Context(),
        h.emailSender,
        []string{userEmail},
        locale,
        docID,
        docURL,
        signURL,
    )

    if err != nil {
        http.Error(w, "Failed to send reminder", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

---

**Implémentation complète et testée** ✅
