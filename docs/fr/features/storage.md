# Stockage de Documents

Uploadez et stockez des documents directement dans Ackify.

## Vue d'Ensemble

Ackify supporte optionnellement le stockage de documents, permettant aux utilisateurs d'uploader des fichiers directement au lieu de fournir des URLs externes. Les documents sont stockés de manière sécurisée et servis via des endpoints API authentifiés.

**Options de stockage :**
- **Désactivé** (par défaut) - Les utilisateurs fournissent des URLs de documents
- **Système de fichiers local** - Documents stockés sur le serveur
- **Compatible S3** - AWS S3, MinIO, Wasabi, DigitalOcean Spaces, etc.

## Types de Fichiers Supportés

| Type | Extensions | Types MIME |
|------|------------|------------|
| PDF | `.pdf` | `application/pdf` |
| Images | `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp` | `image/*` |
| Office | `.doc`, `.docx` | `application/msword`, `application/vnd.openxmlformats-*` |
| Texte | `.txt` | `text/plain` |
| HTML | `.html`, `.htm` | `text/html` |

## Configuration

### Stockage Local

Stockez les documents sur le système de fichiers du serveur via un volume Docker.

```env
ACKIFY_STORAGE_TYPE=local
ACKIFY_STORAGE_LOCAL_PATH=/data/documents
ACKIFY_STORAGE_MAX_SIZE_MB=50
```

**Volume Docker Compose :**
```yaml
services:
  ackify-ce:
    volumes:
      - ackify_storage:/data/documents

volumes:
  ackify_storage:
```

### Stockage Compatible S3

Fonctionne avec tout fournisseur de stockage compatible S3.

```env
ACKIFY_STORAGE_TYPE=s3
ACKIFY_STORAGE_MAX_SIZE_MB=50
ACKIFY_STORAGE_S3_ENDPOINT=https://s3.amazonaws.com
ACKIFY_STORAGE_S3_BUCKET=ackify-documents
ACKIFY_STORAGE_S3_ACCESS_KEY=votre_access_key
ACKIFY_STORAGE_S3_SECRET_KEY=votre_secret_key
ACKIFY_STORAGE_S3_REGION=us-east-1
ACKIFY_STORAGE_S3_USE_SSL=true
```

### MinIO (S3 Auto-hébergé)

MinIO est une solution de stockage compatible S3 open-source populaire.

```env
ACKIFY_STORAGE_TYPE=s3
ACKIFY_STORAGE_S3_ENDPOINT=http://minio:9000
ACKIFY_STORAGE_S3_BUCKET=ackify-documents
ACKIFY_STORAGE_S3_ACCESS_KEY=minioadmin
ACKIFY_STORAGE_S3_SECRET_KEY=minioadmin
ACKIFY_STORAGE_S3_REGION=us-east-1
ACKIFY_STORAGE_S3_USE_SSL=false
```

**Docker Compose avec MinIO :**
```yaml
services:
  minio:
    image: minio/minio:latest
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  minio_data:
```

## Utilisation

### Interface Utilisateur

Quand le stockage est activé, un bouton d'upload apparaît à côté du champ URL :

1. Cliquez sur le bouton upload (ou glissez-déposez)
2. Sélectionnez un fichier depuis votre ordinateur
3. Le nom et la taille du fichier sont affichés
4. Cliquez sur "Upload" pour soumettre

**Fonctionnalités :**
- Barre de progression pendant l'upload
- Titre automatique depuis le nom de fichier
- Validation du type de fichier
- Respect de la limite de taille

### Endpoints API

#### Upload de Document

```http
POST /api/v1/documents/upload
Content-Type: multipart/form-data
X-CSRF-Token: abc123

file: (binaire)
title: Titre optionnel du document
```

**Réponse :**
```json
{
  "success": true,
  "data": {
    "doc_id": "abc123",
    "title": "document.pdf",
    "storage_key": "abc123/document.pdf",
    "storage_provider": "local",
    "file_size": 1048576,
    "mime_type": "application/pdf",
    "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "checksum_algorithm": "SHA-256",
    "created_at": "2025-01-07T12:00:00Z",
    "is_new": true
  }
}
```

#### Obtenir le Contenu du Document

```http
GET /api/v1/storage/{docId}/content
```

Retourne le fichier document avec l'en-tête `Content-Type` approprié.

**Note :** Nécessite une session authentifiée.

## Sécurité

### Authentification

- Tous les endpoints de stockage nécessitent une authentification
- Les documents ne sont accessibles qu'aux utilisateurs authentifiés
- Protection CSRF sur l'endpoint d'upload

### Vérification de Checksum

- Checksum SHA-256 calculé automatiquement à l'upload
- Stocké dans les métadonnées du document
- Garantit l'intégrité du fichier

### Validation des Fichiers

- Type de fichier vérifié contre les types MIME autorisés
- Taille du fichier validée contre `ACKIFY_STORAGE_MAX_SIZE_MB`
- Nom de fichier assaini pour éviter les traversées de chemin

## Schéma de Base de Données

Les documents uploadés ajoutent des champs à la table `documents` :

```sql
ALTER TABLE documents ADD COLUMN storage_key TEXT;
ALTER TABLE documents ADD COLUMN storage_provider TEXT;
ALTER TABLE documents ADD COLUMN file_size BIGINT;
ALTER TABLE documents ADD COLUMN mime_type TEXT;
```

## Bonnes Pratiques

### Choix du Stockage

| Scénario | Stockage Recommandé |
|----------|---------------------|
| Déploiement serveur unique | Local |
| Plusieurs serveurs / scaling | S3 |
| Environnement isolé | Local |
| Déploiement cloud-native | S3 |
| Développement / tests | Local ou MinIO |

### Sauvegarde

**Stockage local :**
- Inclure le volume `ackify_storage` dans la stratégie de backup
- Utiliser `docker-volume-backup` ou outils similaires

**Stockage S3 :**
- Configurer le versioning du bucket
- Activer la réplication cross-région si nécessaire
- Utiliser les politiques de cycle de vie pour la rétention

### Performance

- Activer SSL S3 en production (`ACKIFY_STORAGE_S3_USE_SSL=true`)
- Utiliser des endpoints S3 régionaux pour une latence réduite
- Considérer un CDN pour les documents fréquemment accédés

## Limitations

- Taille maximum de fichier : configurable, 50MB par défaut
- Pas de scan antivirus (à implémenter au niveau infrastructure)
- Pas de génération d'aperçu de document
- Pas de compression automatique

## Dépannage

### L'upload échoue avec "Storage not configured"

Assurez-vous que `ACKIFY_STORAGE_TYPE` est défini à `local` ou `s3`.

### Erreurs de connexion S3

1. Vérifiez le format de l'URL endpoint (inclure `http://` ou `https://`)
2. Vérifiez l'access key et la secret key
3. Vérifiez que le bucket existe et est accessible
4. Vérifiez que le paramètre SSL correspond à l'endpoint

### Erreur fichier trop volumineux

Augmentez `ACKIFY_STORAGE_MAX_SIZE_MB` ou réduisez la taille du fichier.

### Permission refusée (stockage local)

Assurez-vous que le conteneur a accès en écriture au chemin de stockage :
```bash
docker exec ackify-ce ls -la /data/documents
```
