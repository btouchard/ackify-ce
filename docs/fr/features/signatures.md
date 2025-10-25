# Signatures Cryptographiques

Flow complet de signature avec Ed25519 et garanties de sécurité.

## Principe

Ackify utilise **Ed25519** (courbe elliptique) pour créer des signatures cryptographiques non-répudiables.

**Garanties** :
- ✅ **Non-répudiation** - La signature prouve l'identité du signataire
- ✅ **Intégrité** - Le hash SHA-256 détecte toute modification
- ✅ **Horodatage immutable** - Triggers PostgreSQL empêchent la backdating
- ✅ **Unicité** - Une seule signature par utilisateur/document

## Flow de Signature

### 1. Utilisateur accède au document

```
https://sign.company.com/?doc=policy_2025
```

Le frontend Vue.js charge et affiche :
- Titre du document (si metadata existe)
- Nombre de signatures existantes
- Bouton "Sign this document"

### 2. Vérification de session

Le frontend appelle :
```http
GET /api/v1/users/me
```

**Si non connecté** → Redirection OAuth2
**Si connecté** → Affichage du bouton de signature

### 3. Signature

Au clic sur "Sign", le frontend :

1. Obtient un token CSRF :
```http
GET /api/v1/csrf
```

2. Envoie la signature :
```http
POST /api/v1/signatures
Content-Type: application/json
X-CSRF-Token: abc123

{
  "doc_id": "policy_2025"
}
```

### 4. Backend Processing

Le backend (Go) :

1. **Vérifie la session** - Utilisateur authentifié
2. **Génère la signature Ed25519** :
   ```go
   payload := fmt.Sprintf("%s:%s:%s:%s", docID, userSub, userEmail, timestamp)
   hash := sha256.Sum256([]byte(payload))
   signature := ed25519.Sign(privateKey, hash[:])
   ```
3. **Calcule prev_hash** - Hash de la dernière signature (chaînage)
4. **Insère en base** :
   ```sql
   INSERT INTO signatures (doc_id, user_sub, user_email, signed_at, payload_hash, signature, nonce, prev_hash)
   VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
   ```
5. **Retourne la signature** au frontend

### 5. Confirmation

Le frontend affiche :
- ✅ Signature confirmée
- Horodatage
- Lien vers la liste des signatures

## Structure de la Signature

```json
{
  "docId": "policy_2025",
  "userEmail": "alice@company.com",
  "userName": "Alice Smith",
  "signedAt": "2025-01-15T14:30:00Z",
  "payloadHash": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "signature": "ed25519:3045022100...",
  "nonce": "abc123xyz",
  "prevHash": "sha256:prev..."
}
```

**Champs** :
- `payloadHash` - SHA-256 du payload (doc_id:user_sub:email:timestamp)
- `signature` - Signature Ed25519 en base64
- `nonce` - Protection anti-replay
- `prevHash` - Hash de la signature précédente (blockchain-like)

## Vérification de Signature

### Manuelle (via API)

```http
GET /api/v1/documents/policy_2025/signatures
```

Retourne toutes les signatures avec :
- Email signataire
- Horodatage
- Hash + signature

### Programmation (Go)

```go
import "crypto/ed25519"

func VerifySignature(publicKey ed25519.PublicKey, payload, signature []byte) bool {
    hash := sha256.Sum256(payload)
    return ed25519.Verify(publicKey, hash[:], signature)
}
```

## Contraintes PostgreSQL

### Une signature par user/document

```sql
UNIQUE (doc_id, user_sub)
```

**Comportement** :
- Si l'utilisateur tente de signer 2 fois → Erreur 409 Conflict
- Le frontend détecte cela et affiche "Already signed"

### Immutabilité de `created_at`

Trigger PostgreSQL :
```sql
CREATE TRIGGER prevent_signatures_created_at_update
    BEFORE UPDATE ON signatures
    FOR EACH ROW
    EXECUTE FUNCTION prevent_created_at_update();
```

**Garantie** : Impossible de backdater une signature.

## Chaînage (Blockchain-like)

Chaque signature référence la précédente via `prev_hash` :

```
Signature 1 → hash1
Signature 2 → hash2 (prev_hash = hash1)
Signature 3 → hash3 (prev_hash = hash2)
```

**Détection de tampering** :
- Si une signature est modifiée, le `prev_hash` de la suivante ne correspond plus
- Permet de détecter toute modification de l'historique

## Sécurité

### Clé Privée Ed25519

Générée automatiquement au premier démarrage ou via :

```bash
ACKIFY_ED25519_PRIVATE_KEY=$(openssl rand -base64 64)
```

**Important** :
- La clé privée ne quitte jamais le serveur
- Stockée en mémoire uniquement (pas en base)
- Backup requis si vous voulez garder la même clé après redéploiement

### Protection Anti-Replay

Le `nonce` unique empêche la réutilisation d'une signature :
```go
nonce := fmt.Sprintf("%s-%d", userSub, time.Now().UnixNano())
```

### Rate Limiting

Les signatures sont limitées à **100 requêtes/minute** par IP.

## Cas d'Usage

### Validation de lecture de politique

```
Document: "Security Policy 2025"
URL: https://sign.company.com/?doc=security_policy_2025
```

**Workflow** :
1. Admin envoie le lien aux employés
2. Chaque employé clique, lit, et signe
3. Admin voit la completion dans `/admin`

### Accusé de réception formation

```
Document: "GDPR Training 2025"
Expected signers: 50 employés
```

**Features** :
- Tracking de complétion (42/50 = 84%)
- Rappels email automatiques
- Export des signatures

### Acknowledgment contractuel

```
Document: "Terms of Service v3"
Checksum: SHA-256 du PDF
```

**Vérification** :
- Utilisateur calcule le checksum du PDF
- Compare avec la metadata stockée
- Signe si identique

Voir [Checksums](checksums.md) pour plus de détails.

## API Reference

Voir [API Documentation](../api.md) pour tous les endpoints liés aux signatures.
