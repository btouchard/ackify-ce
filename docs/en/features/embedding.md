# Embedding & Integrations

Integrate Ackify into your tools (Notion, Outline, Google Docs, etc.).

## URL Formats: When to Use What

Ackify provides two URL formats depending on your use case:

### `/?doc=<id>` - Full Page Experience

**Use for**:
- Direct links in emails, chat messages
- Standalone signature pages
- When you want full navigation and context

**Behavior**:
- Full page with header, navigation, and footer
- Optimized for direct user access
- Complete branding and organization context

**Example**:
```
https://sign.company.com/?doc=policy_2025
```

### `/embed?doc=<id>` - Embed-Optimized View

**Use for**:
- iFrame embeds in Notion, Outline, Confluence
- Widget integrations in third-party platforms
- When you want a clean, minimal interface

**Behavior**:
- Minimal interface without navigation
- Optimized for small iframe containers
- Focuses only on signature status and action button
- No automatic redirects

**Example**:
```
https://sign.company.com/embed?doc=policy_2025
```

> **Important**: For embedding in Notion/Outline, always use `/embed?doc=...` to avoid unwanted redirections and get the optimal embed experience.

## Integration Methods

### 1. Direct Link

The simplest:
```
https://sign.company.com/?doc=policy_2025
```

**Usage**:
- Email
- Chat (Slack, Teams)
- Wiki
- Documentation

**Behavior**:
- User clicks → Arrives at signature page
- Logs in via OAuth2
- Signs the document

### 2. iFrame Embed

To integrate in a web page:

```html
<iframe src="https://sign.company.com/embed?doc=policy_2025"
        width="600"
        height="200"
        frameborder="0"
        style="border: 1px solid #ddd; border-radius: 6px;">
</iframe>
```

**Render**:
```
┌─────────────────────────────────────┐
│ 📄 Security Policy 2025             │
│ 42 confirmations                    │
│ [Sign this document]                │
└─────────────────────────────────────┘
```

### 3. oEmbed (Auto-discovery)

For platforms supporting oEmbed (Notion, Outline, Confluence, etc.).

#### How it works

1. Paste URL in your editor:
   ```
   https://sign.company.com/?doc=policy_2025
   ```

2. Editor auto-detects via meta tag:
   ```html
   <link rel="alternate" type="application/json+oembed"
         href="https://sign.company.com/oembed?url=..." />
   ```

3. Editor calls `/oembed` and receives:
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

4. Editor displays iframe automatically

#### Supported Platforms

- ✅ **Notion** - Paste URL → Auto-embed
- ✅ **Outline** - Paste URL → Auto-embed
- ✅ **Confluence** - oEmbed macro
- ✅ **AppFlowy** - URL unfurling
- ✅ **Slack** - Link unfurling (Open Graph)
- ✅ **Microsoft Teams** - Card preview
- ✅ **Discord** - Rich embed

## Open Graph & Twitter Cards

Ackify automatically generates meta tags for previews:

```html
<!-- Auto-generated for /?doc=policy_2025 -->
<meta property="og:title" content="Security Policy 2025 - 42 confirmations" />
<meta property="og:description" content="42 people confirmed reading the document" />
<meta property="og:url" content="https://sign.company.com/?doc=policy_2025" />
<meta property="og:type" content="website" />
<meta name="twitter:card" content="summary" />
```

**Result in Slack/Teams**:

```
┌─────────────────────────────────────┐
│ 🔐 Ackify                           │
│ Security Policy 2025                │
│ 42 confirmations                    │
│ sign.company.com                    │
└─────────────────────────────────────┘
```

## Specific Integrations

### Notion

**Method 1: Auto-embed (Recommended)**

1. Paste URL in a Notion page:
   ```
   https://sign.company.com/embed?doc=policy_2025
   ```

2. Notion auto-detects oEmbed
3. Widget appears with signature button

**Method 2: Manual embed**

1. Type `/embed` in Notion
2. Paste the embed URL:
   ```
   https://sign.company.com/embed?doc=policy_2025
   ```

> **Tip**: Use `/embed?doc=...` instead of `/?doc=...` to get the optimal embed view without navigation elements.

### Outline

1. In an Outline document, paste:
   ```
   https://sign.company.com/embed?doc=policy_2025
   ```

2. Outline automatically loads the widget

> **Note**: Using `/embed?doc=...` ensures you get the clean widget view without redirects.

### Google Docs

Google Docs doesn't support iframes directly, but:

1. **Option 1 - Link**:
   ```
   Please sign: https://sign.company.com/?doc=policy_2025
   ```

2. **Option 2 - Image Badge**:
   ```
   ![Sign now](https://sign.company.com/badge/policy_2025.png)
   ```

3. **Option 3 - Google Sites**:
   - Create a Google Sites page
   - Insert iframe
   - Link from Google Docs

See [docs/integrations/google-doc/](../integrations/google-doc/) for more details.

### Confluence

1. Edit a Confluence page
2. Insert "oEmbed" or "HTML Embed" macro
3. Paste:
   ```
   https://sign.company.com/embed?doc=policy_2025
   ```

> **Note**: Use `/embed?doc=...` for the best iframe experience.

### Slack

**Link Unfurling**:

1. Post URL in a channel:
   ```
   Hey team, please sign: https://sign.company.com/?doc=policy_2025
   ```

2. Slack automatically displays preview (Open Graph)

**Slash Command** (future):
```
/ackify sign policy_2025
```

### Microsoft Teams

1. Post URL in a conversation
2. Teams displays card preview (Open Graph)

## PNG Badge

Generate a visual badge for README, wiki, etc.

### Badge URL

```
https://sign.company.com/badge/policy_2025.png
```

**Render**:

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

## oEmbed API

### Endpoint

```http
GET /oembed?url=https://sign.company.com/?doc=policy_2025
```

**Response**:
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

### Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `url` | Document URL (required) | `?url=https://...` |
| `maxwidth` | Max width (optional) | `?maxwidth=800` |
| `maxheight` | Max height (optional) | `?maxheight=300` |

### Discovery

All pages include discovery tag:

```html
<link rel="alternate"
      type="application/json+oembed"
      href="https://sign.company.com/oembed?url=..."
      title="Document title" />
```

## Customization

### Dark Mode Theme

Widget automatically detects browser's dark mode:

```css
@media (prefers-color-scheme: dark) {
  /* Automatic dark theme */
}
```

### Custom Size

```html
<iframe src="https://sign.company.com/?doc=policy_2025"
        width="800"
        height="300"
        frameborder="0">
</iframe>
```

### Language

Widget automatically detects browser language:
- `fr` - Français
- `en` - English
- `es` - Español
- `de` - Deutsch
- `it` - Italiano

## Security

### iFrame Sandboxing

By default, Ackify iframes allow:
- `allow-same-origin` - OAuth2 cookies
- `allow-scripts` - Vue.js features
- `allow-forms` - Signature submission
- `allow-popups` - OAuth redirect

### CORS

Ackify automatically configures CORS for:
- All origins (public reading)
- Credentials via `Access-Control-Allow-Credentials`

### CSP

Content Security Policy headers configured to allow embedding:

```
X-Frame-Options: SAMEORIGIN
Content-Security-Policy: frame-ancestors 'self' https://notion.so https://outline.com
```

## Troubleshooting

### iframe not displaying

Verify:
- HTTPS enabled (required for OAuth)
- CSP headers allow embedding
- No content blocker (uBlock, Privacy Badger)

### oEmbed not detected

Verify:
- Tag `<link rel="alternate" type="application/json+oembed">` is present
- URL is exact (with `?doc=...`)
- Platform supports oEmbed discovery

### Slack preview empty

Verify:
- Open Graph meta tags present
- URL publicly accessible
- No infinite redirect

## Complete Examples

See:
- [docs/integrations/google-doc/](../integrations/google-doc/) - Google Workspace integration
- More examples coming...
