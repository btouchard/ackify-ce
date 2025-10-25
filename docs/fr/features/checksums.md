# Checksums

Vérification d'intégrité des documents avec tracking.

## Vue d'Ensemble

Ackify permet de stocker et vérifier les checksums (empreintes) des documents pour garantir leur intégrité.

**Algorithmes supportés** :
- SHA-256 (recommandé)
- SHA-512
- MD5 (legacy)

## Calculer un Checksum

### Ligne de Commande

```bash
# Linux/Mac - SHA-256
sha256sum document.pdf
# Output: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  document.pdf

# SHA-512
sha512sum document.pdf

# MD5
md5sum document.pdf

# Windows PowerShell
Get-FileHash document.pdf -Algorithm SHA256
Get-FileHash document.pdf -Algorithm SHA512
Get-FileHash document.pdf -Algorithm MD5
```

### Client-Side (JavaScript)

Le frontend Vue.js utilise la **Web Crypto API** :

```javascript
async function calculateChecksum(file) {
  const arrayBuffer = await file.arrayBuffer()
  const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}

// Utilisation
const file = document.querySelector('input[type="file"]').files[0]
const checksum = await calculateChecksum(file)
console.log('SHA-256:', checksum)
```

## Stocker le Checksum

### Via le Dashboard Admin

1. Aller sur `/admin`
2. Sélectionner un document
3. Cliquer "Edit Metadata"
4. Remplir :
   - **Checksum** : e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
   - **Algorithm** : SHA-256
   - **Document URL** : https://docs.company.com/policy.pdf

### Via l'API

```http
PUT /api/v1/admin/documents/policy_2025/metadata
Content-Type: application/json
X-CSRF-Token: abc123

{
  "title": "Security Policy 2025",
  "url": "https://docs.company.com/policy.pdf",
  "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "checksumAlgorithm": "SHA-256",
  "description": "Annual security policy"
}
```

## Vérification

### Interface Utilisateur

Le frontend affiche :
```
Document: Security Policy 2025
Checksum (SHA-256): e3b0c44...52b855 [Copy]
URL: https://docs.company.com/policy.pdf [Open]

[Upload file to verify]
```

**Workflow utilisateur** :
1. Télécharge le document depuis l'URL
2. Upload dans l'interface de vérification
3. Le checksum est calculé client-side
4. Comparaison automatique avec le stocké
5. ✅ Match ou ❌ Mismatch

### Vérification Manuelle

```bash
# 1. Télécharger le document
wget https://docs.company.com/policy.pdf

# 2. Calculer le checksum
sha256sum policy.pdf
# e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

# 3. Comparer avec la valeur stockée (via API)
curl http://localhost:8080/api/v1/documents/policy_2025
# "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

# 4. Si identique → Document intègre
```


## Cas d'Usage

### Compliance Documentaire

```
Document: "ISO 27001 Certification"
Checksum: SHA-256 du PDF officiel
```

**Workflow** :
- Stocker le checksum du document certifié
- Chaque reviewer vérifie l'intégrité avant signature
- Audit trail de toutes les vérifications

### Contrat Légal

```
Document: "Service Agreement v2.3"
Checksum: SHA-512 pour sécurité maximale
URL: https://legal.company.com/contracts/sa-v2.3.pdf
```

**Garanties** :
- Le document signé correspond exactement à la version checksum
- Détection de toute modification
- Traçabilité des vérifications

### Formation avec Support

```
Document: "GDPR Training Materials"
Checksum: SHA-256 du fichier ZIP
```

**Utilisation** :
- Participants téléchargent le ZIP
- Vérifient le checksum avant de commencer
- Signent après complétion

## Sécurité

### Choix de l'Algorithme

| Algorithme | Sécurité | Performance | Recommandation |
|------------|----------|-------------|----------------|
| SHA-256 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ Recommandé |
| SHA-512 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Maximum security |
| MD5 | ⭐⭐ | ⭐⭐⭐⭐⭐ | ❌ Legacy only |

**Recommandation** : Utiliser **SHA-256** par défaut.

### Limitations de MD5

MD5 est **déprécié** pour la sécurité :
- Collisions possibles (deux fichiers différents = même hash)
- Utilisable uniquement pour compatibilité legacy

### Web Crypto API

La vérification client-side utilise l'API native du navigateur :
- Pas de dépendance externe
- Performance native
- Supporté par tous les navigateurs modernes

## Intégration avec Signatures

Workflow complet :

```
1. Admin upload document → calcule checksum → stocke metadata
2. User télécharge document → vérifie checksum client-side
3. Si checksum OK → User signe le document
4. Signature liée au doc_id avec checksum stocké
```

**Garantie** : La signature prouve que l'utilisateur a lu **exactement** la version checksum.

## Bonnes Pratiques

### Stockage

- ✅ Toujours stocker le checksum **avant** d'envoyer le lien de signature
- ✅ Inclure l'URL du document dans la metadata
- ✅ Utiliser SHA-256 minimum
- ✅ Documenter l'algorithme utilisé

### Vérification

- ✅ Encourager les utilisateurs à vérifier avant de signer
- ✅ Afficher le checksum de manière visible (avec bouton Copy)
- ✅ Alerter en cas de mismatch

### Audit

- ✅ Surveiller l'intégrité des documents
- ✅ Vérifier régulièrement les checksums

## Limitations

- **Vérification manuelle uniquement** - Les utilisateurs doivent calculer et comparer les checksums manuellement
- **Pas d'API de vérification côté serveur** - La vérification des checksums se fait côté client ou manuellement
- **Pas d'historique automatisé** - La table `checksum_verifications` existe dans le schéma de base de données mais n'est pas actuellement utilisée par l'API
- Pas de signature du checksum (fonctionnalité future : signer le checksum avec Ed25519)
- Pas d'intégration avec stockage cloud (S3, GCS) pour récupération automatique

## Implémentation Actuelle

Actuellement, Ackify supporte :
- ✅ Stockage des checksums dans les métadonnées de document (via dashboard admin ou API)
- ✅ Affichage des checksums aux utilisateurs pour vérification manuelle
- ✅ Calcul de checksum côté client avec Web Crypto API
- ✅ Calcul automatique de checksum pour les URLs distantes (admin uniquement)

Les fonctionnalités futures pourraient inclure :
- Endpoints API pour le suivi des vérifications de checksum
- Workflows de vérification automatisés
- Intégration avec des services de vérification externes
