# 🚀 Ackify CE — Présentation

Ackify CE est une plateforme open-source et auto-hébergeable qui permet de prouver la lecture d’un document, sans signature électronique.

Pensé pour les équipes modernes (Outline, Notion, Confluence…), Ackify apporte une preuve de lecture horodatée et vérifiable, idéale pour les politiques internes, formations ou procédures de conformité.

## Fonctionnalités principales

- 🔐 Preuve de lecture vérifiable : une confirmation unique par utilisateur et par document, avec horodatage et hachage cryptographique (SHA-256 / Ed25519).
- ✉️ Authentification flexible : connexion par **Magic Link** (sans mot de passe, e-mail 15 min) ou **OAuth2** (Google, GitHub, GitLab, etc.).
- 🧭 Tableau de bord administrateur : gestion des documents, liste des lecteurs attendus, suivi des confirmations et envois de rappels.
- 🌐 Intégrations : insertion simple dans **Outline, Notion, Confluence ou tout site via iframe et oEmbed**.
- 🔁 **API REST & Webhooks** pour automatiser les notifications et intégrer Ackify à d’autres outils.
- 🧱 Installation guidée en 5 minutes via un script interactif Docker (distroless + PostgreSQL 16).
- 🌍 Interface multilingue (FR, EN, ES, DE, IT).
- 🧑‍💻 Mode “admin-only” (variable ACKIFY_ONLY_ADMIN_CAN_CREATE) pour restreindre la création de documents.

## Cas d’usage

- Validation de politiques internes (sécurité, RGPD, conformité).
- Attestation de lecture de formation ou de procédures.
- Suivi de documents sensibles sans recourir à une signature électronique complète.

## Stack & sécurité

- Backend : Go + PostgreSQL 16
- Frontend : Vue 3 + TypeScript + Tailwind
- Architecture : API-first, distroless, healthchecks intégrés
- Sécurité : Ed25519, SHA-256, PKCE, cookies sécurisés, taux de requêtes limités

## Démarrage rapide 🚀

```bash
curl -fsSL https://raw.githubusercontent.com/btouchard/ackify-ce/main/install/install.sh | bash
```

Le script installe Ackify, configure le `.env` en mode interactif, génère les secrets.

## Licence & liens

- Licence : AGPL v3
- Code source : https://github.com/btouchard/ackify-ce
- Site officiel : https://www.ackify.eu
- Documentation : /docs/ du dépôt GitHub