# Architecture Technique - MMO Fantastique (2D)

## 1. Stack Technique

| Composant | Technologie | Justification |
|-----------|-------------|---------------|
| **Client** | Godot 4.3 (GDScript) | Léger, open-source, **excellent rendu 2D natif** |
| **Serveur** | Go 1.21+ | Performance, concurrence native, typage fort |
| **Database** | PostgreSQL | ACID, relations complexes, scalabilité verticale |
| **Cache** | Redis | Sessions, state temporaire, rate limiting |
| **Network** | WebSocket + TCP | Temps réel + fiabilité pour données critiques |
| **Message Format** | Protocol Buffers | Performance, schema évolution |

**Note**: Transition vers **2D Top-Down** adoptée pour le Sprint 1. Voir [DECISION_2D.md](./DECISION_2D.md) pour justifications.

---

## 2. Architecture Serveur (Go)

### 2.1 Pattern: Monolith Modulaire

```
server/
├── cmd/
│   ├── game-server/      # Serveur de jeu principal
│   ├── auth-server/      # Service d'authentification
│   └── gateway/          # API Gateway (WebSocket termination)
├── internal/
│   ├── auth/
│   │   ├── service.go       # Logique auth
│   │   ├── jwt.go           # Token management
│   │   └── middleware.go    # Auth middleware
│   ├── game/
│   │   ├── world/           # Gestion du monde
│   │   │   ├── world.go         # World state (2D Grid)
│   │   │   ├── entities.go      # Player/NPC entities
│   │   │   └── zones.go         # Zone management
│   │   ├── combat/
│   │   │   ├── combat.go        # Combat engine
│   │   │   ├── skills.go        # Skill system
│   │   │   └── damage.go        # Damage calculation
│   │   ├── inventory/
│   │   │   └── inventory.go     # Item management
│   │   └── quest/
│   │       └── quest.go         # Quest logic
│   ├── network/
│   │   ├── websocket.go     # WS handler
│   │   ├── protocol.go      # Proto parsing
│   │   └── sync.go          # State sync
│   └── database/
│       ├── postgres/        # PostgreSQL queries
│       └── redis/           # Cache layer
├── pkg/
│   ├── models/          # Shared data structures
│   ├── protocol/        # ProtoBuf generated code
│   └── utils/           # Helpers
└── configs/
    └── config.yaml
```

### 2.2 Architecture de Synchronisation

```
┌─────────────┐     WebSocket      ┌──────────────┐
│   Client    │ ◄───────────────── │   Gateway    │
│   (Godot 2D)│     TCP/WS         │   (Go)       │
└─────────────┘                    └──────┬───────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
              ┌─────▼─────┐         ┌─────▼─────┐         ┌─────▼─────┐
              │  Auth     │         │  Game     │         │  Chat     │
              │  Service  │         │  Service  │         │  Service  │
              └─────┬─────┘         └─────┬─────┘         └─────┬─────┘
                    │                     │                     │
                    └─────────────────────┼─────────────────────┘
                                          │
                                    ┌─────▼─────┐
                                    │  Redis    │
                                    │  Cache    │
                                    └─────┬─────┘
                                          │
                                    ┌─────▼─────┐
                                    │ Postgres  │
                                    │   DB      │
                                    └───────────┘
```

### 2.3 Patterns de Communication

**Client → Serveur (Commands)**
```protobuf
message ClientCommand {
  string command_id = 1;
  int64 timestamp = 2;
  oneof payload {
    MoveCommand2D move = 10;     // Changed to 2D
    AttackCommand attack = 11;
    UseSkillCommand use_skill = 12;
    ChatMessage chat = 13;
  }
}
```

**Serveur → Client (State Updates)**
```protobuf
message ServerState {
  int64 server_tick = 1;
  repeated EntityUpdate2D entities = 2;  // Changed to 2D
  repeated WorldEvent events = 3;
  StateDiff diff = 4;
}

message Position2D {
  float x = 1;
  float y = 2;
}
```

---

## 3. Architecture Client (Godot 2D)

### 3.1 Structure des Scènes

```
client/
├── scenes/
│   ├── main/
│   │   ├── Main.tscn           # Scene root (2D World)
│   │   └── Main.gd              # Game manager
│   ├── ui/
│   │   ├── HUD.tscn            # Heads-up display (2D)
│   │   ├── inventory/
│   │   │   ├── Inventory.tscn
│   │   │   └── Inventory.gd
│   │   ├── chat/
│   │   │   └── ChatPanel.tscn
│   │   └── menus/
│   │       ├── AuthMenu.tscn
│   │       ├── CharacterCreation.tscn
│   │       └── CharacterSelection.tscn
│   ├── character/
│   │   ├── Player.tscn         # 2D Sprite-based player
│   │   ├── Player.gd
│   │   ├── NPC.tscn
│   │   └── animations/
│   ├── world/
│   │   ├── World.tscn          # 2D TileMap-based world
│   │   ├── Zone.tscn
│   │   └── minimap/
│   └── combat/
│       ├── CombatHUD.tscn
│       └── SkillBar.tscn
├── scripts/
│   ├── network/
│   │   ├── NetworkManager.gd    # WS connection
│   │   ├── Protocol.gd          # Proto encoding
│   │   └── SyncManager.gd       # State sync
│   ├── gameplay/
│   │   ├── PlayerController2D.gd  # Input → commands (2D)
│   │   ├── CombatSystem.gd      # Combat logic
│   │   └── QuestSystem.gd
│   ├── ui/
│   │   ├── UIManager.gd
│   │   └── InventoryManager.gd
│   └── utils/
│       ├── Logger.gd
│       └── Config.gd
├── assets/
│   ├── sprites/                # 2D sprites (was: models/)
│   ├── tilesets/               # Tile maps for world
│   ├── audio/
│   └── fonts/
└── protocol/
    └── *.proto  # ProtoBuf definitions
```

### 3.2 Network Manager (Godot)

```gdscript
# scripts/network/NetworkManager.gd
class_name NetworkManager
extends Node

signal connected()
signal disconnected()
signal state_updated(state: Dictionary)
signal command_sent(command: ClientCommand)

var ws_peer: WebSocketPeer
var server_url: String
var player_id: String
var is_connected_to_server: bool = false

func connect_to_server(url: String) -> void:
    server_url = url
    ws_peer = WebSocketPeer.new()
    ws_peer.connect_to_url(url)

func _process(delta: float) -> void:
    ws_peer.poll()
    
    if ws_peer.get_ready_state() == WebSocketPeer.STATE_OPEN:
        while ws_peer.get_available_packet_count() > 0:
            var packet = ws_peer.get_packet()
            _handle_server_message(packet)

func send_command(command: ClientCommand) -> void:
    var data = _encode_command(command)
    ws_peer.send_text(data)
    command_sent.emit(command)
```

---

## 4. Protocole Réseau

### 4.1 Message Types

| Type | Direction | Usage |
|------|-----------|-------|
| `AUTH_REQUEST` | C→S | Connexion joueur |
| `AUTH_RESPONSE` | S→C | Token validation |
| `MOVE_COMMAND_2D` | C→S | Déplacement (x, y) |
| `ATTACK_COMMAND` | C→S | Attaque |
| `SKILL_USE` | C→S | Utilisation compétence |
| `CHAT_MESSAGE` | C→S | Chat texte |
| `STATE_UPDATE` | S→C | État du monde |
| `ENTITY_UPDATE_2D` | S→C | Entities (2D positions) |
| `COMBAT_EVENT` | S→C | Dégâts, healing |
| `QUEST_UPDATE` | S→C | Progression quête |

### 4.2 Synchronisation

**Approche: Client-side Prediction + Server Reconciliation**

```
Client Tick 100 ──┐
                  ├─► Send MOVE_2D(100) ──► Server
Client Tick 101 ──┤                      │
                  │                      ▼
Client Tick 102 ──┘              Process at Tick 100
                                  │
                                  ▼
                          Send STATE_2D(100) ──► Client
                                                  │
                                                  ▼
                                          Reconcile state
                                          (rollback + replay)
```

### 4.3 Lag Compensation

- **Interpolation**: Délai 100ms pour affichage entities
- **Extrapolation**: Prédiction mouvement NPCs
- **Server Authority**: Validation toutes les actions

---

## 5. Design Base de Données

### 5.1 Schéma PostgreSQL

```sql
-- Users & Authentication
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_banned BOOLEAN DEFAULT FALSE
);

-- Characters (2D positions)
CREATE TABLE characters (
    character_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id),
    name VARCHAR(50) NOT NULL,
    class_type VARCHAR(20) NOT NULL,  -- 'warrior', 'rogue', 'mage'
    level INTEGER DEFAULT 1,
    exp BIGINT DEFAULT 0,
    current_zone VARCHAR(50),
    position_x FLOAT,    -- 2D X
    position_y FLOAT,    -- 2D Y (was: position_x, _y, _z)
    orientation FLOAT,   -- Rotation 0-360 degrees
    is_online BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Inventory
CREATE TABLE inventory (
    inventory_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID REFERENCES characters(character_id),
    item_id INTEGER NOT NULL,
    quantity INTEGER DEFAULT 1,
    slot_index INTEGER,  -- NULL = bag
    equipment_slot VARCHAR(20),  -- 'weapon', 'armor', etc
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Skills
CREATE TABLE character_skills (
    character_id UUID REFERENCES characters(character_id),
    skill_id INTEGER NOT NULL,
    level INTEGER DEFAULT 1,
    PRIMARY KEY (character_id, skill_id)
);

-- Quest Progress
CREATE TABLE quest_progress (
    character_id UUID REFERENCES characters(character_id),
    quest_id INTEGER NOT NULL,
    status VARCHAR(20),  -- 'active', 'completed'
    progress_data JSONB,  -- {kill_count: 5, items_collected: [1,2,3]}
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (character_id, quest_id)
);

-- Guilds
CREATE TABLE guilds (
    guild_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    leader_id UUID REFERENCES characters(character_id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE guild_members (
    guild_id UUID REFERENCES guilds(guild_id),
    character_id UUID REFERENCES characters(character_id),
    role VARCHAR(20),  -- 'leader', 'officer', 'member'
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (guild_id, character_id)
);
```

### 5.2 Redis Cache

```
Key Patterns:
- session:{token} → {user_id, expires_at}
- player:{character_id} → {position_x, position_y, state, last_seen}
- world:zone:{zone_id} → {entities[], timestamp}
- guild:{guild_id} → {members[], data}
- rate_limit:{user_id} → {count, reset_time}
```

---

## 6. Sécurité

### 6.1 Authentification

```
1. Client envoie: username + password
2. Server vérifie bcrypt hash
3. Server génère JWT (expiry: 24h)
4. Client inclut JWT dans tous les messages
5. Server valide JWT via middleware
```

```go
// internal/auth/jwt.go
func GenerateJWT(userID string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.JWTSecret))
}
```

### 6.2 Validation Inputs

```go
// Toutes les commandes client doivent être validées
func ValidateMoveCommand2D(cmd *MoveCommand2D) error {
    // Position within bounds
    if cmd.X < -10000 || cmd.X > 10000 {
        return ErrInvalidPosition
    }
    if cmd.Y < -10000 || cmd.Y > 10000 {
        return ErrInvalidPosition
    }
    
    // Timestamp not too far in future (anti-cheat)
    if cmd.Timestamp > time.Now().UnixMilli() + 5000 {
        return ErrTimeAnomaly
    }
    
    // Rate limiting
    if !rateLimiter.Allow(cmd.UserID) {
        return ErrRateLimited
    }
    
    return nil
}
```

### 6.3 Anti-Cheat

| Mécanisme | Description |
|-----------|-------------|
| **Server Authority** | Toutes les actions validées serveur |
| **Rate Limiting** | Max X commandes/seconde |
| **Time Validation** | Timestamps cohérents |
| **Position Check** | Speed hack detection |
| **Resource Validation** | Or/exp cohérents avec actions |
| **Anomaly Detection** | ML sur patterns de jeu |

### 6.4 Rate Limiting

```go
// Redis-based rate limiter
func NewRateLimiter(redis *redis.Client) *RateLimiter {
    return &RateLimiter{
        redis: redis,
        maxRequests: 50,  // requests per second
        window: time.Second,
    }
}

func (rl *RateLimiter) Allow(userID string) bool {
    key := fmt.Sprintf("ratelimit:%s", userID)
    count, _ := rl.redis.Incr(key).Result()
    
    if count == 1 {
        rl.redis.Expire(key, rl.window)
    }
    
    return count <= rl.maxRequests
}
```

---

## 7. Infrastructure

### 7.1 Déploiement

```
┌─────────────────────────────────────────────┐
│              Load Balancer                  │
│              (nginx/HAProxy)                │
└──────────────────┬──────────────────────────┘
                   │
     ┌─────────────┼─────────────┐
     │             │             │
┌────▼────┐  ┌────▼────┐  ┌────▼────┐
│ Game    │  │ Game    │  │ Game    │
│ Server 1│  │ Server 2│  │ Server N│
│ (Go)    │  │ (Go)    │  │ (Go)    │
└────┬────┘  └────┬────┘  └────┬────┘
     │             │             │
     └─────────────┼─────────────┘
                   │
     ┌─────────────┼─────────────┐
     │             │             │
┌────▼────┐  ┌────▼────┐  ┌────▼────┐
│ Redis   │  │Postgres │  │  Backup │
│ Cluster │  │ Primary │  │  Replic │
└─────────┘  └─────────┘  └─────────┘
```

### 7.2 Scaling

- **Horizontal**: Plus de game servers (sharding par zone)
- **Vertical**: DB upgrade, Redis cluster
- **Auto-scaling**: Basé sur CPU/memory usage

---

## 8. ADRs (Architecture Decision Records)

### ADR-001: Choix PostgreSQL vs MongoDB

**Decision**: PostgreSQL

**Rationale**:
- Données relationnelles (characters ↔ inventory ↔ guilds)
- ACID compliance pour transactions critiques
- Requêtes complexes pour leaderboards, stats
- Scalabilité verticale suffisante pour V1

### ADR-002: Protocol Buffers vs JSON

**Decision**: Protocol Buffers

**Rationale**:
- 60-70% plus petit que JSON
- Parsing plus rapide
- Schema évolution native
- Typage fort

### ADR-003: Monolith vs Microservices

**Decision**: Monolith Modulaire

**Rationale**:
- Complexité réduite pour V1
- Déploiement simplifié
- Debugging plus facile
- Peut migrer vers microservices si besoin

### ADR-004: 3D vers 2D Top-Down

**Decision**: Transition vers 2D Top-Down pour Sprint 1

**Rationale**:
- **Développement plus rapide**: Sprites vs modèles 3D
- **Assets accessibles**: Plus facile trouver/créer sprites 2D
- **Gameplay clair**: Meilleure visibilité tactique en top-down
- **Performance**: Moins gourmand, plus de joueurs simultanés
- **Scope réduit**: MVP plus rapide à valider

**Impact**:
- Database: position_z supprimé
- Client: Node3D → Node2D/Control
- Scenes: Camera3D → Camera2D
- PlayerController: CharacterBody3D → CharacterBody2D

---

## 9. Roadmap Technique

### Phase 1 (Mois 1-3)
- [ ] Setup projet Go + Godot
- [ ] Auth system (JWT)
- [ ] WebSocket connection
- [ ] Database schema v1
- [ ] Basic movement sync (2D)

### Phase 2 (Mois 4-6)
- [ ] Combat system serveur
- [ ] Inventory system
- [ ] Quest system
- [ ] Redis caching
- [ ] Rate limiting

### Phase 3 (Mois 7-9)
- [ ] Guild system
- [ ] Chat system
- [ ] Anti-cheat v1
- [ ] Load testing
- [ ] Monitoring

### Phase 4 (Mois 10-12)
- [ ] Sharding support
- [ ] Auto-scaling
- [ ] Backup/restore
- [ ] Security audit
- [ ] Performance optimization

---

*Dernière mise à jour: 17 Mars 2026*  
*Propriétaire: @cto*  
*Transition 2D: Voir [DECISION_2D.md](./DECISION_2D.md)*
