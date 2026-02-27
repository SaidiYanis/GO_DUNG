# Dungeons API (Go + Gin + MongoDB)

API REST pour un jeu dungeon géolocalisé.

Architecture utilisée sur tous les domaines:
- `models` (entités + DTO)
- `repositories/<domain>` (Mongo CRUD pur)
- `services/<domain>` (règles métier + validation)
- `controllers/<domain>` (HTTP parse + mapping erreurs)
- `routes/<domain>` (wiring Gin)

Format d'erreur JSON standard:
```json
{"error":{"code":"...","message":"..."}}
```

## Variables d'environnement
Copier `.env.example` vers `.env`.

- `API_VERSION` version exposée par `/version`
- `MODE` `DEVELOP` ou autre
- `DB_HOST` URI MongoDB
- `DB_NAME` nom de DB
- `DB_TIMEOUT_SECONDS` timeout des opérations DB
- `TOKEN_KEY` secret de signature des tokens
- `TOKEN_TTL_HOURS` durée de vie token
- `API_PORT` port API (`8080` ou `:8080`)
- `ALLOW_ORIGIN` CORS
- `LOG_FORMAT` `HUMAN` ou `JSON`
- `SEED_ON_BOOT` `true/false`

## Lancer l'API
```bash
go run ./cmd/api
```

Health:
```bash
curl http://localhost:8080/ping
curl http://localhost:8080/version
```

## Seeder
Seeder autonome:
```bash
go run ./cmd/seed\n# ou\ngo run ./cmd/bootloader
```

Ou auto-seed au boot:
```env
SEED_ON_BOOT=true
```

Comptes seed:
- MJ: `mj@seed.local` / `Password123!`
- Player: `player@seed.local` / `Password123!`

## Endpoints MVP

### Auth / Player
- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `GET /v1/me`

### Dungeon (MJ)
- `POST /v1/mj/dungeons`
- `PUT /v1/mj/dungeons/{id}`
- `POST /v1/mj/dungeons/{id}/publish`
- `POST /v1/mj/dungeons/{id}/steps`
- `PUT /v1/mj/dungeons/{id}/steps/{stepId}`
- `PUT /v1/mj/dungeons/{id}/steps/reorder`

### Dungeon (Player)
- `GET /v1/dungeons`
- `GET /v1/dungeons/{id}`

### Runs / Attempt
- `POST /v1/runs`
- `GET /v1/runs`
- `GET /v1/runs/{id}`
- `POST /v1/runs/{id}/steps/{stepId}/attempt`

### Inventory / Auction
- `GET /v1/inventory`
- `POST /v1/auction/listings`
- `GET /v1/auction/listings`
- `POST /v1/auction/listings/{id}/buy`
- `POST /v1/auction/listings/{id}/cancel`

## Exemples cURL

### 1) Register + Login + Me
```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"new@player.local","displayName":"NewPlayer","password":"Password123!","role":"player"}'

curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"player@seed.local","password":"Password123!"}'
```

Puis:
```bash
TOKEN="<token>"
curl http://localhost:8080/v1/me -H "Authorization: Bearer $TOKEN"
```

### 2) Flow Run + Attempt
```bash
TOKEN="<token-player>"

curl -X POST http://localhost:8080/v1/runs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"dungeonId":"seed-dungeon-1"}'
```

Puis attempt step 1:
```bash
RUN_ID="<run-id>"
curl -X POST http://localhost:8080/v1/runs/$RUN_ID/steps/seed-step-1/attempt \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"lat":48.8566,"lon":2.3522,"idempotencyKey":"attempt-001"}'
```

### 3) Flow Auction
Créer listing (MJ seed):
```bash
TOKEN_MJ="<token-mj>"
curl -X POST http://localhost:8080/v1/auction/listings \
  -H "Authorization: Bearer $TOKEN_MJ" \
  -H "Content-Type: application/json" \
  -d '{"itemId":"seed-item-sword","qty":1,"pricePerUnit":150}'
```

Acheter listing (player):
```bash
TOKEN_PLAYER="<token-player>"
LISTING_ID="<listing-id>"
curl -X POST http://localhost:8080/v1/auction/listings/$LISTING_ID/buy \
  -H "Authorization: Bearer $TOKEN_PLAYER" \
  -H "Content-Type: application/json" \
  -d '{"qty":1}'
```

## Tests
```bash
go test ./...
```

Tests inclus:
- Haversine (distance)
- wrong step order (`409`)
- idempotence replay attempt
- validation/service Player
