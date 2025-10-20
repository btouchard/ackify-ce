# Ackify WebApp - Interface Moderne Vue 3

Interface moderne pour Ackify avec Vue 3, Shadcn Vue et thème Claymorphism.

## Stack Technique

- **Vue 3.5** - Framework JavaScript progressif
- **TypeScript** - Typage statique
- **Vite 7** - Build tool ultra-rapide
- **Tailwind CSS 4** - Framework CSS utility-first
- **Shadcn Vue** - Composants UI modernes et accessibles
- **Radix Vue** - Primitives UI headless accessibles
- **Lucide Vue Next** - Icons modernes
- **Vue Router 4** - Routing officiel
- **Pinia 3** - State management
- **Axios** - Client HTTP

## Thème Claymorphism

L'interface utilise un design claymorphism avec:
- Effets de glassmorphism subtils
- Ombres douces et dégradés
- Backdrop blur pour profondeur
- Palette de couleurs sombre par défaut
- Support du mode clair/sombre

### Variables de thème

Les couleurs sont définies dans `src/style.css` via les variables CSS:

```css
--color-background
--color-foreground
--color-primary (vert: 142.1 76.2% 36.3%)
--color-secondary
--color-muted
--color-accent
--color-destructive
--color-border
--color-input
--color-ring
```

## Structure des Composants

```
src/
├── components/
│   ├── ui/              # Composants Shadcn de base
│   │   ├── Button.vue
│   │   ├── Card.vue
│   │   ├── Input.vue
│   │   └── ...
│   ├── layout/          # Composants de mise en page
│   │   ├── AppHeader.vue     # Header avec navigation
│   │   ├── AppFooter.vue     # Footer avec liens
│   │   ├── AppShell.vue      # Layout principal
│   │   └── ThemeToggle.vue   # Toggle dark/light mode
│   └── accessibility/   # Composants a11y
│       └── SkipToContent.vue
├── pages/               # Pages de l'application
│   ├── HomePage.vue     # Page d'accueil avec hero
│   ├── SignPage.vue     # Page de signature
│   ├── SignaturesPage.vue # Liste des signatures
│   ├── admin/
│   │   ├── AdminDashboard.vue
│   │   └── AdminDocument.vue
│   └── ...
├── stores/              # Pinia stores
├── services/            # Services API
├── composables/         # Composables Vue
│   └── useClickOutside.ts
└── lib/
    └── utils.ts         # Utilitaires (cn, etc.)
```

## Fonctionnalités UI

### Header (AppHeader.vue)
- Navigation responsive avec menu mobile
- Menu utilisateur avec dropdown
- Toggle de thème (dark/light)
- Indicateur de page active
- Support clavier complet (accessibilité)

### Footer (AppFooter.vue)
- Navigation par catégories
- Liens vers ressources
- Liens légaux
- Icônes sociales
- Responsive

### Page d'accueil (HomePage.vue)
- Hero moderne avec gradient background
- Badge d'information
- Boutons CTA avec animations
- Section statistiques (stats cards claymorphism)
- Section "Pourquoi Ackify ?" avec cartes de features
- Section "Comment ça marche ?" avec étapes numérotées
- Section CTA finale
- Effets hover subtils

### Accessibilité
- Support navigation clavier
- Focus rings visibles
- ARIA labels et roles
- Skip to content link
- Contraste AA/AAA
- Support `prefers-reduced-motion`

## Classes Claymorphism Personnalisées

```css
.clay-card         # Carte avec effet glassmorphism
.clay-card-hover   # Carte avec effet hover
.clay-button       # Bouton avec effet glassmorphism
.clay-input        # Input avec effet glassmorphism
```

## Développement

### Installer les dépendances
```bash
npm install
```

### Lancer le serveur de développement
```bash
npm run dev
```
L'application sera accessible sur `http://localhost:5173`

### Build de production
```bash
npm run build
```

### Preview du build
```bash
npm run preview
```

## Configuration

### Proxy API (vite.config.ts)
Le serveur de développement proxie les requêtes API vers le backend Go:
- `/api/*` → `http://localhost:8080`
- `/oauth2/*` → `http://localhost:8080`

### Alias TypeScript
L'alias `@/` pointe vers `src/`

## État d'avancement

### ✅ Complété
- [x] Configuration Tailwind CSS 4
- [x] Installation Shadcn Vue (manuel pour v4)
- [x] Thème claymorphism appliqué
- [x] Composants UI de base (Button, Card, Input)
- [x] Layout complet (Header, Footer, Shell)
- [x] Composants d'accessibilité
- [x] Page d'accueil modernisée avec hero
- [x] Dark mode avec toggle

### 🚧 À faire
- [ ] Moderniser page Sign avec formulaire accessible
- [ ] Moderniser page Signatures avec table Shadcn
- [ ] Moderniser page Admin avec dashboard KPI
- [ ] Ajouter plus de composants UI (Table, Dialog, Dropdown, etc.)
- [ ] Animations de transition entre pages
- [ ] Loading states et skeletons
- [ ] Toasts/notifications améliorés
- [ ] Gestion d'erreurs optimisée
- [ ] Tests unitaires et E2E

## Personnalisation

### Changer les couleurs du thème
Modifier les variables dans `src/style.css`:
```css
@theme {
  --color-primary: <nouvelle-couleur>;
  /* ... */
}
```

### Ajouter un nouveau composant Shadcn
1. Créer le fichier dans `src/components/ui/`
2. Utiliser les utilities `cn()` de `@/lib/utils`
3. Respecter les variants avec `class-variance-authority`
4. S'inspirer des composants existants

## Support navigateurs

- Chrome/Edge (dernières 2 versions)
- Firefox (dernières 2 versions)
- Safari 15+

## Ressources

- [Vue 3 Documentation](https://vuejs.org/)
- [Tailwind CSS v4](https://tailwindcss.com/)
- [Radix Vue](https://www.radix-vue.com/)
- [Shadcn UI](https://ui.shadcn.com/) (référence React, adapté pour Vue)

## Licence

MIT - Voir LICENSE dans le répertoire racine
