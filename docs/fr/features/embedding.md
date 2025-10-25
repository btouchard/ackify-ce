# Embedding & Intégrations

Intégrer Ackify dans vos outils (Notion, Outline, Google Docs, etc.).

## Méthodes d'Intégration

### 1. Lien Direct

Le plus simple :
```
https://sign.company.com/?doc=policy_2025
```

**Usage** :
- Email
- Chat (Slack, Teams)
- Wiki
- Documentation

**Comportement** :
- L'utilisateur clique → Arrive sur la page de signature
- Se connecte via OAuth2
- Signe le document

### 2. iFrame Embed

Pour intégrer dans une page web :

```html
<iframe src="https://sign.company.com/?doc=policy_2025"
        width="600"
        height="200"
        frameborder="0"
        style="border: 1px solid #ddd; border-radius: 6px;">
</iframe>
```

**Rendu** :
```
┌─────────────────────────────────────┐
│ 📄 Security Policy 2025             │
│ 42 confirmations                    │
│ [Sign this document]                │
└─────────────────────────────────────┘
```

### 3. oEmbed (Auto-discovery)

Pour les plateformes supportant oEmbed (Notion, Outline, Confluence, etc.).

#### Comment ça marche

1. Coller l'URL dans votre éditeur :
   ```
   https://sign.company.com/?doc=policy_2025
   ```

2. L'éditeur détecte automatiquement via la balise meta :
   ```html
   <link rel="alternate" type="application/json+oembed"
         href="https://sign.company.com/oembed?url=..." />
   ```

3. L'éditeur appelle `/oembed` et reçoit :
   ```json
   {
     "type": "rich",
     "version": "1.0",
     "title": "Security Policy 2025 - 42 confirmations",
     "provider_name": "Ackify",
     "html": "<iframe src=\"https://sign.company.com/?doc=policy_2025\" ...>",
     "height": 200
   }
   ```

4. L'éditeur affiche l'iframe automatiquement

#### Plateformes Supportées

- ✅ **Notion** - Paste URL → Auto-embed
- ✅ **Outline** - Paste URL → Auto-embed
- ✅ **Confluence** - oEmbed macro
- ✅ **AppFlowy** - URL unfurling
- ✅ **Slack** - Link unfurling (Open Graph)
- ✅ **Microsoft Teams** - Card preview
- ✅ **Discord** - Rich embed

## Open Graph & Twitter Cards

Ackify génère automatiquement des meta tags pour les previews :

```html
<!-- Auto-généré pour /?doc=policy_2025 -->
<meta property="og:title" content="Security Policy 2025 - 42 confirmations" />
<meta property="og:description" content="42 personnes ont confirmé avoir lu le document" />
<meta property="og:url" content="https://sign.company.com/?doc=policy_2025" />
<meta property="og:type" content="website" />
<meta name="twitter:card" content="summary" />
```

**Résultat dans Slack/Teams** :

```
┌─────────────────────────────────────┐
│ 🔐 Ackify                           │
│ Security Policy 2025                │
│ 42 confirmations                    │
│ sign.company.com                    │
└─────────────────────────────────────┘
```

## Intégrations Spécifiques

### Notion

1. Coller l'URL dans une page Notion :
   ```
   https://sign.company.com/?doc=policy_2025
   ```

2. Notion détecte automatiquement l'oEmbed
3. Le widget apparaît avec bouton de signature

**Alternative** : Créer un embed manuel
- `/embed` → Paste URL

### Outline

1. Dans un document Outline, coller :
   ```
   https://sign.company.com/?doc=policy_2025
   ```

2. Outline charge automatiquement le widget

### Google Docs

Google Docs ne supporte pas les iframes directement, mais :

1. **Option 1 - Lien** :
   ```
   Veuillez signer : https://sign.company.com/?doc=policy_2025
   ```

2. **Option 2 - Image Badge** :
   ```
   ![Sign now](https://sign.company.com/badge/policy_2025.png)
   ```

3. **Option 3 - Google Sites** :
   - Créer une page Google Sites
   - Insérer l'iframe
   - Lier depuis Google Docs

Voir [docs/integrations/google-doc/](../integrations/google-doc/) pour plus de détails.

### Confluence

1. Éditer une page Confluence
2. Insérer macro "oEmbed" ou "HTML Embed"
3. Coller :
   ```
   https://sign.company.com/?doc=policy_2025
   ```

### Slack

**Link Unfurling** :

1. Poster l'URL dans un channel :
   ```
   Hey team, please sign: https://sign.company.com/?doc=policy_2025
   ```

2. Slack affiche automatiquement une preview (Open Graph)

**Slash Command** (futur) :
```
/ackify sign policy_2025
```

### Microsoft Teams

1. Poster l'URL dans une conversation
2. Teams affiche une card preview (Open Graph)

## Badge PNG

Générer un badge visuel pour README, wiki, etc.

### URL du Badge

```
https://sign.company.com/badge/policy_2025.png
```

**Rendu** :

![Signature Status](https://img.shields.io/badge/Signatures-42%2F50-green)

### Markdown

```markdown
[![Sign this document](https://sign.company.com/badge/policy_2025.png)](https://sign.company.com/?doc=policy_2025)
```

### HTML

```html
<a href="https://sign.company.com/?doc=policy_2025">
  <img src="https://sign.company.com/badge/policy_2025.png" alt="Signature status">
</a>
```

## API oEmbed

### Endpoint

```http
GET /oembed?url=https://sign.company.com/?doc=policy_2025
```

**Response** :
```json
{
  "type": "rich",
  "version": "1.0",
  "title": "Document policy_2025 - 42 confirmations",
  "provider_name": "Ackify",
  "provider_url": "https://sign.company.com",
  "html": "<iframe src=\"https://sign.company.com/?doc=policy_2025\" width=\"100%\" height=\"200\" frameborder=\"0\" style=\"border: 1px solid #ddd; border-radius: 6px;\" allowtransparency=\"true\"></iframe>",
  "width": null,
  "height": 200
}
```

### Paramètres

| Paramètre | Description | Exemple |
|-----------|-------------|---------|
| `url` | URL du document (obligatoire) | `?url=https://...` |
| `maxwidth` | Largeur max (optionnel) | `?maxwidth=800` |
| `maxheight` | Hauteur max (optionnel) | `?maxheight=300` |

### Discovery

Toutes les pages incluent la balise de discovery :

```html
<link rel="alternate"
      type="application/json+oembed"
      href="https://sign.company.com/oembed?url=..."
      title="Document title" />
```

## Personnalisation

### Thème Dark Mode

Le widget détecte automatiquement le dark mode du navigateur :

```css
@media (prefers-color-scheme: dark) {
  /* Thème sombre automatique */
}
```

### Taille Personnalisée

```html
<iframe src="https://sign.company.com/?doc=policy_2025"
        width="800"
        height="300"
        frameborder="0">
</iframe>
```

### Langue

Le widget détecte automatiquement la langue du navigateur :
- `fr` - Français
- `en` - English
- `es` - Español
- `de` - Deutsch
- `it` - Italiano

## Sécurité

### iFrame Sandboxing

Par défaut, les iframes Ackify autorisent :
- `allow-same-origin` - Cookies OAuth2
- `allow-scripts` - Fonctionnalités Vue.js
- `allow-forms` - Soumission de signatures
- `allow-popups` - OAuth redirect

### CORS

Ackify configure automatiquement CORS pour :
- Toutes les origines (lecture publique)
- Credentials via `Access-Control-Allow-Credentials`

### CSP

Content Security Policy headers configurés pour permettre l'embedding :

```
X-Frame-Options: SAMEORIGIN
Content-Security-Policy: frame-ancestors 'self' https://notion.so https://outline.com
```

## Troubleshooting

### L'iframe ne s'affiche pas

Vérifier :
- HTTPS activé (required pour OAuth)
- CSP headers permettent l'embedding
- Pas de bloqueur de contenu (uBlock, Privacy Badger)

### oEmbed non détecté

Vérifier :
- La balise `<link rel="alternate" type="application/json+oembed">` est présente
- L'URL est exacte (avec `?doc=...`)
- La plateforme supporte oEmbed discovery

### Preview Slack vide

Vérifier :
- Open Graph meta tags présents
- URL publiquement accessible
- Pas de redirect infini

## Exemples Complets

Voir :
- [docs/integrations/google-doc/](../integrations/google-doc/) - Intégration Google Workspace
- Plus d'exemples à venir...
