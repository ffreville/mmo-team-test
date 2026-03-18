# Sprint 1 Plan - Fondations Techniques (2D)

**Sprint**: 1/12  
**Période**: Semaines 1-2 (Mois 1)  
**Phase**: 1 - Fondations (Mois 1-6)  
**Statut**: Planification  
**Perspective**: **2D Top-Down**

---

## Participants

- **@ceo** - Validation scope et priorités
- **@cto** - Architecture technique et validation
- **@dev-team** - Implémentation client/serveur
- **@design-team** - Spécifications gameplay minimal viable

---

## Objectifs du Sprint

### Objectif Principal
Établir les fondations techniques du MMO **2D** avec un prototype de connexion et mouvement de base.

### Objectifs Spécifiques
1. **Infrastructure**: Setup des projets Godot (2D) et Go avec CI/CD
2. **Authentification**: Système de connexion JWT fonctionnel
3. **Réseau**: Connection WebSocket client-serveur établie
4. **Database**: Schema PostgreSQL v1 implémenté (positions 2D)
5. **Mouvement**: Système de déplacement 2D basique synchronisé

### Definition of Done
- [ ] Code review complétée pour toutes les tâches
- [ ] Tests unitaires > 70% coverage sur backend
- [ ] Documentation API mise à jour
- [ ] Déploiement local fonctionnel
- [ ] **Client 2D fonctionnel** (login → création perso → mouvement 2D)

---

## Timeline du Sprint

```
Semaine 1 (Jours 1-5)
├── J1: Setup infrastructure + Architecture validation (2D)
├── J2: Database schema (2D positions) + Auth backend
├── J3: Auth backend + WebSocket gateway
├── J4: Client network manager + Auth UI (2D)
└── J5: Integration testing + Bug fixes

Semaine 2 (Jours 6-10)
├── J6: Character creation backend
├── J7: Character creation client (2D UI)
├── J8: Movement system server (2D)
├── J9: Movement system client 2D + sync
└── J10: Sprint review + Demo preparation
```

---

## Tâches Détaillées

### 1. Infrastructure & Setup (Équipe: Dev Team, CTO)

#### 1.1 Setup Projet Go - **Estimation: 8h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 1

**Tâches**:
```
server/
├── go.mod                              # Module initialization
├── Makefile                            # Build commands
├── Dockerfile                          # Container config
├── docker-compose.yml                  # Local dev environment
├── cmd/
│   ├── game-server/main.go            # Server entry point
│   ├── auth-server/main.go            # Auth service entry
│   └── gateway/main.go                # WebSocket gateway
├── internal/
│   ├── auth/                          # Auth package stub
│   ├── game/                          # Game package stub
│   │   └── world/                     # 2D world management
│   ├── network/                       # Network package stub
│   └── database/                      # Database package stub
├── pkg/
│   ├── models/                        # Shared models (2D)
│   └── protocol/                      # ProtoBuf generated
└── configs/
    └── config.yaml                    # Default config
```

**Acceptance Criteria**:
- [ ] `go build` réussit sans erreurs
- [ ] `make test` exécute les tests unitaires
- [ ] Docker container démarre correctement
- [ ] Configuration load depuis config.yaml

#### 1.2 Setup Projet Godot 2D - **Estimation: 6h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 1

**Tâches**:
```
client/
├── project.godot                       # Project settings (2D)
├── .godot/                             # Editor metadata
├── scenes/
│   ├── main/
│   │   └── Main.tscn                  # 2D World root scene
│   └── ui/
│       └── menus/
│           ├── AuthMenu.tscn
│           ├── CharacterCreation.tscn
│           └── CharacterSelection.tscn
├── scripts/
│   ├── network/                       # Network package
│   ├── gameplay/
│   │   └── PlayerController2D.gd      # 2D player controller
│   └── ui/                            # UI package
├── assets/                            # 2D placeholder assets
│   ├── sprites/
│   └── tilesets/
└── protocol/                          # Proto definitions
```

**Acceptance Criteria**:
- [ ] Projet ouvre dans Godot 4.3 sans erreurs
- [ ] Scene root se charge correctement (2D)
- [ ] Scripts GDScript validés (pas d'erreurs)
- [ ] Structure de dossiers conforme à l'architecture 2D

#### 1.3 CI/CD Pipeline - **Estimation: 4h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 2

**Tâches**:
- Setup GitHub Actions workflow
- Build automatique sur push
- Tests unitaires automatiques
- Linting (golangci-lint, gdscript linter)

**Acceptance Criteria**:
- [ ] Push sur main déclenche build
- [ ] Tests échouent si tests échouent
- [ ] Linting errors bloquent merge

---

### 2. Authentification (Équipe: Dev Team, CTO)

#### 2.1 Database Schema v1 (2D) - **Estimation: 6h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 2

**Tâches**:
```sql
-- users table
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_banned BOOLEAN DEFAULT FALSE
);

-- characters table (2D positions)
CREATE TABLE characters (
    character_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id),
    name VARCHAR(50) NOT NULL,
    class_type VARCHAR(20) DEFAULT 'warrior',
    level INTEGER DEFAULT 1,
    exp BIGINT DEFAULT 0,
    current_zone VARCHAR(50) DEFAULT 'starter_zone',
    position_x FLOAT DEFAULT 0,    -- 2D X
    position_y FLOAT DEFAULT 0,    -- 2D Y
    orientation FLOAT DEFAULT 0,   -- Rotation degrees
    is_online BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_characters_user_id ON characters(user_id);
```

**Acceptance Criteria**:
- [ ] Schema appliqué sur PostgreSQL local
- [ ] Migrations versionnées (golang-migrate ou équivalent)
- [ ] Seed data pour testing
- [ ] Documentation du schema

#### 2.2 Auth Service Backend - **Estimation: 12h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 3

**Tâches**:
```go
// internal/auth/service.go
type AuthService struct {
    db     *database.PostgresDB
    redis  *redis.Client
    jwtKey []byte
}

func (s *AuthService) Register(username, email, password string) (*User, error)
func (s *AuthService) Login(username, password string) (string, error) // returns JWT
func (s *AuthService) ValidateToken(token string) (*Claims, error)
func (s *AuthService) Logout(token string) error
```

**Features**:
- Registration avec bcrypt password hashing
- Login avec JWT generation (24h expiry)
- Token validation middleware
- Redis session management
- Rate limiting (5 req/min par IP)

**Acceptance Criteria**:
- [ ] POST /auth/register crée un user
- [ ] POST /auth/login retourne JWT
- [ ] Middleware protège les routes
- [ ] Token invalidé après logout
- [ ] Rate limiting fonctionne

#### 2.3 Auth UI Client (2D) - **Estimation: 8h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 4

**Tâches**:
```gdscript
# scenes/ui/menus/AuthMenu.tscn (2D Control nodes)
- Login panel
  - Username input
  - Password input
  - Login button
  - Register button
- Register panel
  - Username input
  - Email input
  - Password input
  - Confirm password input
  - Register button
  - Back to login button
- Error message label
- Loading spinner
```

**Acceptance Criteria**:
- [ ] Panel login fonctionne
- [ ] Panel register fonctionne
- [ ] Validation input client-side
- [ ] Communication avec auth API
- [ ] Affichage erreurs serveur
- [ ] Transition vers menu principal après login

---

### 3. Réseau & Synchronisation (Équipe: Dev Team, CTO)

#### 3.1 Protocol Buffers Definitions (2D) - **Estimation: 4h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 2

**Tâches**:
```protobuf
// protocol/auth.proto
message AuthRequest {
    string username = 1;
    string password = 2;
}

message AuthResponse {
    bool success = 1;
    string token = 2;
    string message = 3;
}

// protocol/character.proto
message CreateCharacterRequest {
    string name = 1;
    string class_type = 2;
}

message CreateCharacterResponse {
    bool success = 1;
    Character character = 2;
    string message = 3;
}

message Character {
    string character_id = 1;
    string name = 2;
    string class_type = 3;
    int32 level = 4;
    float position_x = 5;    // 2D X
    float position_y = 6;    // 2D Y
    float orientation = 7;   // Rotation
}

// protocol/movement.proto
message MoveCommand2D {
    string command_id = 1;
    int64 timestamp = 2;
    float target_x = 3;      // 2D target
    float target_y = 4;
}

message MoveResponse2D {
    string command_id = 1;
    bool success = 2;
    float server_x = 3;      // 2D server position
    float server_y = 4;
}

// protocol/common.proto
message Packet {
    PacketType type = 1;
    bytes payload = 2;
}

enum PacketType {
    AUTH_REQUEST = 0;
    AUTH_RESPONSE = 1;
    CREATE_CHARACTER_REQUEST = 2;
    CREATE_CHARACTER_RESPONSE = 3;
    MOVE_COMMAND_2D = 4;
    MOVE_RESPONSE_2D = 5;
}
```

**Acceptance Criteria**:
- [ ] Proto files valides
- [ ] Generated code pour Go
- [ ] Generated code pour GDScript (plugin ou script)
- [ ] Serialization/deserialization testée

#### 3.2 WebSocket Gateway - **Estimation: 10h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 3

**Tâches**:
```go
// internal/network/websocket.go
type Gateway struct {
    clients    map[string]*Client
    auth       *auth.AuthService
    gameServer *game.Server
}

func (g *Gateway) HandleConnection(conn *websocket.Conn) {
    // WS handshake
    // Auth validation
    // Message loop
}

func (g *Gateway) HandleMessage(client *Client, packet *Packet) {
    switch packet.Type {
    case AUTH_REQUEST:
        g.handleAuth(client, packet)
    case MOVE_COMMAND_2D:
        g.handleMove2D(client, packet)
    // ...
    }
}
```

**Features**:
- WebSocket upgrade handler
- Connection management (connect/disconnect)
- Message routing basé sur PacketType
- Heartbeat/ping-pong (30s interval)
- Graceful disconnect handling

**Acceptance Criteria**:
- [ ] Client peut établir connection WS
- [ ] Messages peuvent être envoyés/reçus
- [ ] Auth required pour game messages
- [ ] Heartbeat fonctionne
- [ ] Disconnect propre

#### 3.3 Network Manager Client (2D) - **Estimation: 8h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 4

**Tâches**:
```gdscript
# scripts/network/NetworkManager.gd
class_name NetworkManager
extends Node

signal connected()
signal disconnected()
signal authenticated()
signal state_updated(state: Dictionary)

var ws_peer: WebSocketPeer
var server_url: String
var auth_token: String
var player_id: String

func connect_to_server(url: String) -> void
func authenticate(username: String, password: String) -> bool
func create_character(name: String, class_type: String) -> Character
func send_move_2d(target_x: float, target_y: float) -> void
func _process(delta: float) -> void
```

**Acceptance Criteria**:
- [ ] Connection WS établie
- [ ] Auth flow fonctionne
- [ ] Packet serialization/deserialization
- [ ] Reconnect automatique sur disconnect
- [ ] Timeout handling (10s)

---

### 4. Système de Mouvement 2D (Équipe: Dev Team, Design Team)

#### 4.1 Movement Server (2D) - **Estimation: 10h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 8

**Tâches**:
```go
// internal/game/world/entities.go
type Player2D struct {
    ID           string
    Username     string
    Position     Vector2      // 2D position
    Orientation  float64      // Rotation
    Velocity     Vector2      // 2D velocity
    IsMoving     bool
    LastUpdate   time.Time
}

// internal/game/world/world.go
type World2D struct {
    players  map[string]*Player2D
    zones    map[string]*Zone2D
    bounds   Bounds2D
}

func (w *World2D) MovePlayer2D(playerID string, target Vector2) error {
    // Validate target within bounds
    // Validate distance (speed limit)
    // Update position
    // Broadcast to other players
}

func (w *World2D) ValidateMove2D(player *Player2D, target Vector2) error {
    // Check bounds
    // Check obstacles (2D collision)
    // Check speed (anti-cheat)
}
```

**Features**:
- Position validation 2D
- Speed limit enforcement (anti-speedhack)
- Bounds checking
- Basic collision (zone boundaries)
- State broadcast to clients

**Acceptance Criteria**:
- [ ] Player position update 2D fonctionne
- [ ] Bounds validation fonctionne
- [ ] Speed limit enforced
- [ ] Position broadcast aux autres clients

#### 4.2 Movement Client 2D - **Estimation: 8h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 9

**Tâches**:
```gdscript
# scripts/gameplay/PlayerController2D.gd
class_name PlayerController2D
extends CharacterBody2D

var network: NetworkManager
var target_position: Vector2
var is_moving: bool = false
var move_speed: float = 200.0

func _input(event: InputEvent) -> void:
    # Mouse click to move (2D)
    # WASD movement

func _physics_process(delta: float) -> void:
    # Local movement prediction
    # Animation state (2D sprites)
    # Interpolation from server

func send_move_command_2d() -> void:
    # Send to server
    # Start local animation
```

**Features**:
- Mouse click to move (2D)
- Client-side prediction
- Server reconciliation
- Simple animation state machine (idle, run) - 2D sprite sheets
- Interpolation pour smooth movement

**Acceptance Criteria**:
- [ ] Click to move 2D fonctionne
- [ ] Animation idle/run transition (2D)
- [ ] Server position reconciliation
- [ ] Smooth interpolation

#### 4.3 Zone de Test 2D - **Estimation: 4h**
**Owner**: @design-team  
**Priorité**: Moyenne  
**Jalon**: Jour 9

**Tâches**:
- Créer zone de test simple 2D (100x100 unités)
- TileMap avec sol (grass texture)
- Quelques objets de référence (2D sprites)
- Spawn point défini

**Acceptance Criteria**:
- [ ] Zone visible dans client 2D
- [ ] Player spawn au bon endroit
- [ ] Bounds visibles (wireframe ou couleur différente)

---

### 5. Création de Personnage (Équipe: Dev Team, Design Team)

#### 5.1 Character Creation Backend - **Estimation: 6h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 6

**Tâches**:
```go
func (s *CharacterService) CreateCharacter(userID string, name string, classType string) (*Character, error) {
    // Validate name (length, chars, uniqueness)
    // Validate classType (warrior, rogue, mage)
    // Generate starting position (2D)
    // Insert into database
    // Return character
}

func (s *CharacterService) ListCharacters(userID string) ([]Character, error) {
    // Query characters by user_id
    // Return list
}
```

**Acceptance Criteria**:
- [ ] POST /characters crée un character
- [ ] Validation name fonctionne
- [ ] Validation classType fonctionne
- [ ] GET /characters retourne la liste

#### 5.2 Character Creation Client (2D UI) - **Estimation: 6h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 7

**Tâches**:
```gdscript
# scenes/ui/menus/CharacterCreation.tscn (2D Control nodes)
- Character preview (2D sprite placeholder)
- Name input
- Class selection (3 buttons: Warrior, Rogue, Mage)
- Create button
- Character list (existing characters)
- Back button
```

**Acceptance Criteria**:
- [ ] Input name fonctionne
- [ ] Class selection fonctionne
- [ ] Create button envoie request
- [ ] Affichage erreurs
- [ ] Transition vers world après création

---

## Répartition par Équipe

### @ceo
- **Validation initiale du scope**: 2h (Jour 1)
- **Revue sprint plan**: 1h (Jour 1)
- **Sprint review/demo**: 1h (Jour 10)
- **Total**: 4h

### @cto
- **Architecture review**: 4h (Jours 1-2)
- **Tech guidance**: 6h (Semaine 1-2, en parallèle)
- **Code review**: 8h (Tout le sprint)
- **Sprint review**: 1h (Jour 10)
- **Total**: 19h

### @dev-team
- **Backend development**: 48h
- **Frontend development (2D)**: 40h
- **Integration testing**: 16h
- **Bug fixes**: 8h
- **Documentation**: 4h
- **Total**: 116h (~23h/personne pour 5 développeurs)

### @design-team
- **Gameplay specs 2D**: 4h
- **Zone de test 2D design**: 4h
- **UI/UX wireframes 2D**: 4h
- **Sprint review**: 1h
- **Total**: 13h

---

## Jalons du Sprint

| Jalon | Jour | Livrable | Validateur |
|-------|------|----------|------------|
| **J1**: Infrastructure Ready | J1 | Projects setup, CI/CD working | @cto |
| **J2**: Database Ready (2D) | J2 | Schema applied, migrations working | @cto |
| **J3**: Auth Backend Ready | J3 | Auth API functional | @cto |
| **J4**: Network Ready | J4 | WS connection, auth flow working | @cto |
| **J5**: Sprint Mid-point | J5 | Demo auth + basic UI 2D | @ceo |
| **J6**: Character Creation Backend | J6 | Character API functional | @cto |
| **J7**: Character Creation Client 2D | J7 | Creation UI functional | @design-team |
| **J8**: Movement Backend 2D | J8 | Move server logic working | @cto |
| **J9**: Movement Client 2D | J9 | Full 2D movement loop working | @design-team |
| **J10**: Sprint Complete | J10 | Demo complete 2D flow | @ceo |

---

## Risques & Atténuation

| Risque | Impact | Probabilité | Atténuation |
|--------|--------|-------------|-------------|
| **Complexité WebSocket sous-estimée** | High | Medium | CTO review early, fallback à HTTP polling si nécessaire |
| **ProtoBuf tooling Godot complexe** | Medium | High | Utiliser JSON pour V1 si ProtoBuf trop long |
| **CI/CD setup bloqueant** | Low | Medium | Déploiement manuel en parallèle |
| **Auth JWT security issues** | High | Low | CTO security review, utiliser librairies matures |
| **Client prediction 2D trop complexe** | High | Medium | Implémenter server-authoritative simple d'abord |
| **Assets 2D manquants** | Medium | Medium | Utiliser placeholders (carrés/cercles) pour V1 |

---

## Dépendances Externes

- [ ] Accès à serveur PostgreSQL pour dev
- [ ] Accès à Redis pour dev
- [ ] Godot 4.3 installé sur machines dev
- [ ] Go 1.21+ installé sur machines dev
- [ ] Docker Desktop pour local environment

---

## Métriques de Succès

### Techniques
- [ ] Auth API: <100ms response time
- [ ] WebSocket: <50ms latency
- [ ] Movement sync 2D: <200ms roundtrip
- [ ] Tests coverage: >70% backend

### Fonctionnelles
- [ ] Login/Register: 100% working
- [ ] Character creation: 100% working
- [ ] Movement 2D: 90%+ smooth (no major glitches)
- [ ] No critical bugs blocking demo

---

## Sprint Review Agenda (Jour 10, 1h)

1. **Demo** (20 min)
   - Login flow
   - Character creation
   - **2D Movement in test zone**
   
2. **Démonstration technique** (15 min)
   - Code walkthrough
   - Architecture decisions (2D)
   - Challenges overcome
   
3. **Rétrospective** (15 min)
   - What went well
   - What didn't
   - Improvements for Sprint 2
   
4. **Planning Sprint 2** (10 min)
   - Initial priorities
   - Resource allocation

---

## Notes pour Sprint 2

Si Sprint 1 est réussi, Sprint 2 se concentrera sur:
- **Combat system 2D** de base
- **Inventory system** minimal
- **3 zones 2D** de test (TileMaps)
- **PNJ 2D** de base
- **Système de chat**

---

*Document créé: 16 Mars 2026*  
*Mis à jour: 17 Mars 2026 (Transition 2D)*  
*Propriétaire: @meetings*
