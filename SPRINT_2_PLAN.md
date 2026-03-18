# Sprint 2 Plan - Système de Combat & Contenu de Base

**Sprint**: 2/12  
**Période**: Semaines 3-4 (Mois 1)  
**Phase**: 1 - Fondations (Mois 1-6)  
**Statut**: Planification  
**Perspective**: **2D Top-Down**

---

## Participants

- **@ceo** - Validation scope et priorités
- **@cto** - Architecture technique et validation
- **@dev-team** - Implémentation client/serveur
- **@design-team** - Spécifications gameplay combat/inventory

---

## Objectifs du Sprint

### Objectif Principal
Étendre le prototype avec un **système de combat basique**, un **inventaire minimal**, et **plus de contenu** (zones, PNJ, chat).

### Objectifs Spécifiques
1. **Combat System**: Combat tour par tour ou temps réel basique avec HP et dégâts
2. **Inventory System**: Inventaire minimal avec slots et items de base
3. **Multiple Zones**: 3 zones distinctes avec TileMaps (Village, Forêt, Grotte)
4. **NPCs**: PNJ statiques avec dialogue simple
5. **Chat System**: Chat local et/ou global fonctionnel

### Definition of Done
- [ ] Code review complétée pour toutes les tâches
- [ ] Tests unitaires > 70% coverage sur nouveaux systèmes
- [ ] Documentation des systèmes de combat et inventory
- [ ] Déploiement local fonctionnel avec tous les nouveaux features
- [ ] **Flow complet**: Login → Combat → Inventory → Zones → PNJ → Chat

---

## Timeline du Sprint

```
Semaine 3 (Jours 1-6)
├── J1: Design combat system + Enemy spawn architecture
├── J2: Enemy spawn system backend (spawns, AI basique)
├── J3: Combat backend (HP, dégâts, validation) + integration with enemies
├── J4: Sprite/Animation system client (sprite sheets, idle/run/attack)
├── J5: Combat client (UI, animations) + Enemy display
├── J6: Inventory backend (slots, items, DB)

Semaine 4 (Jours 7-12)
├── J7: Inventory client (UI, drag-drop)
├── J8: Multiple zones (TileMaps, transitions)
├── J9: NPCs avec dialogue system
├── J10: Chat system (local + global)
├── J11: Integration testing + Bug fixes
└── J12: Sprint review + Demo preparation
```

---

## Tâches Détaillées

### 1. Système de Combat (Équipe: Dev Team, Design Team)

#### 1.1 Combat Backend - **Estimation: 10h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 3

**Tâches**:
```go
// internal/game/combat/combat.go
type CombatManager struct {
    players map[string]*CombatState
}

type CombatState struct {
    PlayerID     string
    TargetID     string
    CurrentHP    int
    MaxHP        int
    Damage       int
    IsInCombat   bool
    LastAttack   time.Time
}

func (cm *CombatManager) StartCombat(playerID, targetID string) error
func (cm *CombatManager) Attack(attackerID, defenderID string) error
func (cm *CombatManager) CalculateDamage(attacker, defender *CombatState) int
func (cm *CombatManager) Heal(playerID string, amount int) error
func (cm *CombatManager) GetCombatState(playerID string) *CombatState
```

**Features**:
- Système HP (Health Points) avec max HP configurable
- Calcul de dégâts basé sur la classe (warrior/rogue/mage)
- Cooldown d'attaque (anti-spam)
- Gestion état combat (in combat / out of combat)
- Système de heal basique

**Acceptance Criteria**:
- [ ] `StartCombat` initialise l'état combat
- [ ] `Attack` calcule et applique les dégâts
- [ ] Cooldown d'attaque respecté (1s min)
- [ ] HP ne descend pas en dessous de 0
- [ ] `GetCombatState` retourne l'état actuel

#### 1.2 Combat Client UI - **Estimation: 8h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 3

**Tâches**:
```gdscript
# scenes/ui/combat/CombatHUD.tscn
- Health bar (current/max HP)
- Attack button
- Heal button (if available)
- Combat log (last 5 actions)
- Target info panel (name, HP, class)

# scripts/gameplay/CombatController.gd
class_name CombatController
extends Control

var network: NetworkManager
var current_hp: int
var max_hp: int
var is_in_combat: bool

func _ready() -> void
func _on_attack_pressed() -> void
func _on_heal_pressed() -> void
func update_combat_state(state: Dictionary) -> void
func add_combat_log(message: String) -> void
```

**Features**:
- Health bar visuelle (vert → jaune → rouge)
- Bouton Attack avec cooldown visuel
- Bouton Heal (optionnel selon classe)
- Combat log texte
- Target info panel

**Acceptance Criteria**:
- [ ] Health bar affiche HP correct
- [ ] Attack button envoie commande au serveur
- [ ] Cooldown visuel sur boutons
- [ ] Combat log montre actions récentes
- [ ] Target info montre nom et HP

---

### 2. Système d'Inventory (Équipe: Dev Team)

#### 2.1 Inventory Backend - **Estimation: 8h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 4

**Tâches**:
```go
// internal/game/inventory/inventory.go
type Inventory struct {
    UserID string
    Slots  []InventorySlot
}

type InventorySlot struct {
    SlotID   int
    ItemID   string
    Quantity int
    ItemType string  // weapon, armor, potion, etc.
}

type Item struct {
    ItemID    string
    Name      string
    ItemType  string
    Stats     ItemStats
    Rarity    string  // common, uncommon, rare, epic, legendary
}

type ItemStats struct {
    Damage     int
    Defense    int
    HealAmount int
}

func (i *Inventory) AddItem(itemID string, quantity int) error
func (i *Inventory) RemoveItem(itemID string, quantity int) error
func (i *Inventory) GetSlot(slotID int) *InventorySlot
func (i *Inventory) EquipItem(slotID int) error
```

**Features**:
- Inventory avec slots fixes (ex: 20 slots)
- Types d'items: weapon, armor, potion
- Système de quantité (stacking pour potions)
- Equipement d'items (weapon/armor)
- Persistance en DB

**Acceptance Criteria**:
- [ ] `AddItem` ajoute item ou stack existant
- [ ] `RemoveItem` retire item correctement
- [ ] `EquipItem` applique les stats
- [ ] Inventory sauvegardé en DB
- [ ] Items de base seedés (sword, shield, health potion)

#### 2.2 Inventory Client UI - **Estimation: 8h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 5

**Tâches**:
```gdscript
# scenes/ui/inventory/InventoryScreen.tscn
- Inventory grid (4x5 = 20 slots)
- Item tooltip on hover
- Equip/Unequip buttons
- Character stats panel
- Close button

# scripts/ui/InventoryManager.gd
class_name InventoryManager
extends Control

var network: NetworkManager
var inventory_slots: Array = []

func _ready() -> void
func _on_slot_clicked(slot_id: int) -> void
func _on_equip_pressed(slot_id: int) -> void
func update_inventory(items: Array) -> void
func show_tooltip(item: Dictionary) -> void
```

**Features**:
- Grille 4x5 pour les items
- Tooltip au hover sur les items
- Bouton Equip/Unequip
- Stats du personnage affichées
- Drag-drop (optionnel si temps)

**Acceptance Criteria**:
- [ ] Grille affiche tous les slots
- [ ] Items visibles dans les slots
- [ ] Tooltip montre info item
- [ ] Equip met à jour les stats
- [ ] Close button retourne au jeu

---

### 3. Zones & Contenu (Équipe: Dev Team, Design Team)

#### 3.1 Multiple Zones - **Estimation: 10h**
**Owner**: @dev-team  
**Priorité**: Haute  
**Jalon**: Jour 6

**Tâches**:
```gdscript
# scenes/world/ZoneManager.gd
class_name ZoneManager
extends Node2D

var current_zone: String
var zones: Dictionary = {
    "village": {
        "scene": "res://scenes/world/village.tscn",
        "bounds": Rect2(-500, -500, 1000, 1000),
        "spawn_point": Vector2(0, 0)
    },
    "forest": {
        "scene": "res://scenes/world/forest.tscn",
        "bounds": Rect2(-800, -800, 1600, 1600),
        "spawn_point": Vector2(0, 0)
    },
    "cave": {
        "scene": "res://scenes/world/cave.tscn",
        "bounds": Rect2(-400, -400, 800, 800),
        "spawn_point": Vector2(0, 0)
    }
}

func switch_zone(zone_name: String) -> void
func get_current_zone() -> String
```

**Features**:
- 3 zones avec TileMaps:
  - **Village**: Sol vert clair, maisons, PNJ
  - **Forêt**: Sol vert foncé, arbres denses
  - **Grotte**: Sol gris, rochers, ambiance sombre
- Transitions entre zones (portails/entrées)
- Spawn point par zone
- Bounds par zone

**Acceptance Criteria**:
- [ ] 3 zones créées avec TileMaps
- [ ] Transition zone fonctionne
- [ ] Spawn point correct par zone
- [ ] Bounds respectés par zone
- [ ] Current zone sauvegardée

#### 3.2 NPC System - **Estimation: 6h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 7

**Tâches**:
```gdscript
# scenes/world/npc/NPC2D.tscn
- Sprite2D (placeholder)
- DialogueBox (CanvasLayer)
- Interaction range (CircleShape2D)

# scripts/world/NPCController.gd
class_name NPCController
extends CharacterBody2D

var npc_id: String
var name: String
var dialogue_lines: Array = []
var interaction_range: float = 50.0
var is_interactable: bool = false

func _ready() -> void
func _on_player_entered_range() -> void
func _on_player_exited_range() -> void
func show_dialogue() -> void
func hide_dialogue() -> void
```

**Features**:
- PNJ statiques avec nom affiché
- Interaction range (50 units)
- Dialogue box avec plusieurs lignes
- Fermeture dialogue (E ou click)
- PNJ de base:
  - **Village**: Guard, Merchant, Quest giver

**Acceptance Criteria**:
- [ ] PNJ visible avec nom
- [ ] Interaction range détectée
- [ ] Dialogue box s'affiche
- [ ] Multiple dialogue lines
- [ ] Close dialogue fonctionne

---

### 4. Chat System (Équipe: Dev Team)

#### 4.1 Chat Backend - **Estimation: 4h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 8

**Tâches**:
```go
// internal/network/chat.go
type ChatManager struct {
    messages []ChatMessage
    maxMessages int
}

type ChatMessage struct {
    SenderID   string
    SenderName string
    Message    string
    Timestamp  time.Time
    Channel    string  // "local", "global", "party"
}

func (cm *ChatManager) SendMessage(senderID, senderName, message, channel string) error
func (cm *ChatManager) GetRecentMessages(channel string, count int) []ChatMessage
func (cm *ChatManager) BroadcastToZone(zoneID, msgType string, payload interface{})
```

**Features**:
- Canaux: local, global, party (local par défaut)
- Message history (derniers 50 messages)
- Rate limiting (5 messages/seconde)
- Flood protection
- Anti-spam (mots interdits optionnel)

**Acceptance Criteria**:
- [ ] `SendMessage` accepte les messages
- [ ] Rate limiting fonctionne
- [ ] `GetRecentMessages` retourne l'historique
- [ ] Broadcast aux joueurs de la zone
- [ ] Channel local/global fonctionnel

#### 4.2 Chat Client UI - **Estimation: 4h**
**Owner**: @dev-team  
**Priorité**: Moyenne  
**Jalon**: Jour 8

**Tâches**:
```gdscript
# scenes/ui/chat/ChatPanel.tscn
- Chat log (scrollable)
- Input line
- Send button
- Channel selector (local/global)

# scripts/ui/ChatManager.gd
class_name ChatManager
extends Control

var network: NetworkManager
var chat_history: Array = []

func _ready() -> void
func _on_send_pressed() -> void
func _on_channel_changed(channel: String) -> void
func add_message(sender: String, message: String, channel: String) -> void
```

**Features**:
- Chat log scrollable (derniers 50 messages)
- Input line pour taper messages
- Send button + Enter pour envoyer
- Channel selector (local/global)
- Couleurs par channel

**Acceptance Criteria**:
- [ ] Chat log affiche messages
- [ ] Input line fonctionne
- [ ] Send button envoie message
- [ ] Channel selector change channel
- [ ] Messages reçus affichés

---

## Répartition par Équipe

### @ceo
- **Validation initiale du scope**: 2h (Jour 1)
- **Revue sprint plan**: 1h (Jour 1)
- **Sprint review/demo**: 1h (Jour 10)
- **Total**: 4h

### @cto
- **Architecture review**: 4h (Jours 1-2)
- **Tech guidance**: 6h (Semaine 3-4, en parallèle)
- **Code review**: 8h (Tout le sprint)
- **Sprint review**: 1h (Jour 10)
- **Total**: 19h

### @dev-team
- **Combat system**: 18h (backend + client)
- **Inventory system**: 16h (backend + client)
- **Zones & NPCs**: 16h (zones + NPCs)
- **Chat system**: 8h (backend + client)
- **Integration testing**: 12h
- **Bug fixes**: 8h
- **Total**: 78h (~15h/personne pour 5 développeurs)

### @design-team
- **Combat gameplay specs**: 4h
- **Inventory UI/UX specs**: 4h
- **Zone design (3 zones)**: 6h
- **NPC dialogue scripts**: 4h
- **Sprint review**: 1h
- **Total**: 19h

---

## Jalons du Sprint

| Jalon | Jour | Livrable | Validateur |
|-------|------|----------|------------|
| **J1**: Combat Design | J1 | Combat specs validées | @ceo |
| **J2**: Combat Backend Ready | J3 | Combat API fonctionnelle | @cto |
| **J3**: Combat Client Ready | J3 | Combat UI fonctionnelle | @design-team |
| **J4**: Inventory Backend Ready | J4 | Inventory API fonctionnelle | @cto |
| **J5**: Sprint Mid-point | J5 | Demo combat + inventory | @ceo |
| **J6**: Zones Ready | J6 | 3 zones avec TileMaps | @design-team |
| **J7**: NPCs Ready | J7 | PNJ avec dialogue | @design-team |
| **J8**: Chat Ready | J8 | Chat system fonctionnel | @design-team |
| **J9**: Integration Complete | J9 | Tous systèmes intégrés | @cto |
| **J10**: Sprint Complete | J10 | Demo complete flow | @ceo |

---

## Risques & Atténuation

| Risque | Impact | Probabilité | Atténuation |
|--------|--------|-------------|-------------|
| **Combat trop complexe** | High | Medium | Commencer avec combat simple, itérer |
| **Inventory UI trop longue** | Medium | Medium | UI basique d'abord, améliorer après |
| **TileMaps prennent trop de temps** | Medium | High | Utiliser tilesets gratuits, simplifier |
| **NPC dialogue system complexe** | Low | Medium | Dialogue texte simple d'abord |
| **Chat flood/spam** | Low | Low | Rate limiting strict dès le début |

---

## Dépendances Externes

- [ ] Sprint 1 complètement stable
- [ ] Assets TileMap disponibles (gratuits ou placeholders)
- [ ] Sprite PNJ placeholders
- [ ] Design des items de base (weapon, armor, potion)

---

## Métriques de Succès

### Techniques
- [ ] Combat response time: <50ms
- [ ] Inventory operations: <30ms
- [ ] Chat latency: <20ms
- [ ] Zone transitions: <200ms

### Fonctionnelles
- [ ] Combat: 100% working (attack, HP, cooldown)
- [ ] Inventory: 100% working (add, remove, equip)
- [ ] Zones: 3 zones fonctionnelles
- [ ] NPCs: 3 PNJ avec dialogue
- [ ] Chat: Local + global working
- [ ] No critical bugs blocking demo

---

## Sprint Review Agenda (Jour 10, 1h)

1. **Demo** (25 min)
   - Login → Enter world
   - Combat avec PNJ/monstre
   - Ouvrir inventory, equiper item
   - Voyager entre zones
   - Interagir avec PNJ
   - Envoyer messages chat

2. **Démonstration technique** (15 min)
   - Code walkthrough combat system
   - Architecture inventory
   - Zone transition logic

3. **Rétrospective** (15 min)
   - What went well
   - What didn't
   - Improvements for Sprint 3

4. **Planning Sprint 3** (5 min)
   - Initial priorities
   - Resource allocation

---

## Notes pour Sprint 3

Si Sprint 2 est réussi, Sprint 3 se concentrera sur:
- **Système de quêtes** de base
- **Monstres/ennemis** AI simple
- **Leveling system** (XP, niveaux)
- **Plus d'items** et d'équipements
- **Sound effects** et musique

---

*Document créé: March 18, 2026*  
*Propriétaire: @meetings*  
*Status: Awaiting CEO approval*