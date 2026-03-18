# Sprint 2 Kickoff Meeting - Système de Combat & Contenu de Base

**Date**: March 18, 2026  
**Duration**: 1h30  
**Participants**: @ceo, @cto, @dev-team, @design-team  
**Facilitator**: @meetings

---

## 1. Sprint 1 Retrospective (10 min)

### ✅ Achievements
- **5/5 objectifs atteints** : Infrastructure, Auth, WebSocket, Database, Movement
- **Flow complet fonctionnel** : Login → Character Creation → World → Movement
- **37 tests unitaires** passants
- **Delta movement** avec validation anti-cheat
- **Decorations** (arbres, rochers, buissons) pour visualisation

### 🎯 Lessons Learned
**What went well:**
- Parallel development (client + backend)
- Rapid iteration on movement system
- Database persistence from day 1
- Clear communication between teams

**Areas for improvement:**
- More upfront testing strategy
- Earlier CI/CD setup
- Better documentation during development

**Technical debt:**
- Integration tests needed
- API documentation incomplete
- No performance benchmarking yet

---

## 2. Sprint 2 Objectives (15 min)

### 🎯 Primary Goal
Étendre le prototype avec **combat**, **inventory**, **zones**, **PNJ**, et **chat**.

### 📋 5 Main Objectives

| # | Objective | Owner | Estimation | Status |
|---|-----------|-------|------------|--------|
| 1 | **Combat System** (HP, dégâts, cooldown) | @dev-team | 18h | Planned |
| 2 | **Inventory System** (slots, items, equip) | @dev-team | 16h | Planned |
| 3 | **3 Zones** (Village, Forêt, Grotte) | @dev-team + @design-team | 10h | Planned |
| 4 | **NPCs avec dialogue** | @dev-team + @design-team | 6h | Planned |
| 5 | **Chat System** (local + global) | @dev-team | 8h | Planned |
| 6 | **Sprite & Animation** (NEW) | @dev-team | 8h | Planned |

### 📅 Timeline : 12 Jours

```
Semaine 3 (Jours 1-6)
├── J1: Design combat + Enemy spawn architecture
├── J2: Enemy spawn system backend
├── J3: Combat backend + integration
├── J4: Sprite/Animation system client
├── J5: Combat client + Enemy display
└── J6: Inventory backend

Semaine 4 (Jours 7-12)
├── J7: Inventory client
├── J8: Multiple zones (TileMaps)
├── J9: NPCs dialogue
├── J10: Chat system
├── J11: Integration testing
└── J12: Sprint review
```

---

## 3. Task Breakdown (20 min)

### 3.1 Combat System (18h)

**Backend (10h)** - @dev-team
- `CombatManager` : StartCombat, Attack, CalculateDamage, Heal
- HP system avec max HP configurable
- Damage basé sur classe (warrior/rogue/mage)
- Cooldown attaque (1s min)
- Acceptance: Attack fonctionne, cooldown respecté, HP correct

**Client UI (8h)** - @dev-team
- Health bar (vert → jaune → rouge)
- Attack button avec cooldown visuel
- Combat log (dernières 5 actions)
- Target info panel
- Acceptance: UI réactive, feedback visuel clair

### 3.2 Enemy Spawn System (6h)

**Backend** - @dev-team
- `EnemyManager` : SpawnEnemy, DespawnEnemy, GetNearbyEnemies, RespawnEnemy
- Spawn points par zone
- Types: Slime, Goblin, Wolf, Skeleton
- Aggro range detection
- Acceptance: Ennemis spawn, aggro joueur, respawn auto

### 3.3 Sprite & Animation System (8h) ⭐ NEW

**Client** - @dev-team
- Sprite sheets pour chaque classe (4 directions × 4 states)
- States: IDLE, RUN, ATTACK, HIT, DEAD
- Transition fluide entre états
- Enemy sprites différents par type
- **Assets disponibles** : `/sprites/Player/`, `/sprites/Enemies/`
- Acceptance: Animations fluides, sprites corrects

### 3.4 Inventory System (16h)

**Backend (8h)** - @dev-team
- Inventory avec 20 slots
- Types: weapon, armor, potion
- Add/Remove/Equip items
- Persistance DB
- Items seedés: sword, shield, health potion

**Client UI (8h)** - @dev-team
- Grille 4x5 slots
- Tooltip au hover
- Equip/Unequip buttons
- Character stats panel
- Acceptance: Inventory fonctionnel, equip met à jour stats

### 3.5 Multiple Zones (10h)

**Backend + Client** - @dev-team + @design-team
- 3 zones: Village (vert clair), Forêt (vert foncé), Grotte (gris)
- TileMaps pour chaque zone
- Transitions entre zones (portails)
- Spawn point par zone
- **Assets disponibles** : `/sprites/Tiles/`, `/sprites/Outdoor decoration/`
- Acceptance: 3 zones fonctionnelles, transitions OK

### 3.6 NPCs with Dialogue (6h)

**Backend + Client** - @dev-team + @design-team
- PNJ statiques avec nom affiché
- Interaction range (50 units)
- Dialogue box avec plusieurs lignes
- PNJ: Guard, Merchant, Quest giver (Village)
- Acceptance: Dialogue s'affiche, multiple lines, close OK

### 3.7 Chat System (8h)

**Backend (4h)** - @dev-team
- Canaux: local, global, party
- Message history (50 messages)
- Rate limiting (5 msg/sec)
- Broadcast to zone

**Client UI (4h)** - @dev-team
- Chat log scrollable
- Input line + Send button
- Channel selector
- Couleurs par channel
- Acceptance: Local + global fonctionnel, rate limiting OK

---

## 4. Technical Architecture (10 min) - @cto

### Combat Architecture
```go
// server/internal/game/combat/combat.go
type CombatManager struct {
    players map[string]*CombatState
    enemies map[string]*CombatState
}

func (cm *CombatManager) Attack(attackerID, defenderID string) error
func (cm *CombatManager) CalculateDamage(attacker, defender *CombatState) int
```

### Enemy Spawn Architecture
```go
// server/internal/game/enemies/enemy.go
type EnemyManager struct {
    enemies      map[string]*Enemy
    spawnPoints  []SpawnPoint
    respawnTimer *time.Ticker
}

func (em *EnemyManager) SpawnEnemy(zoneID, enemyType string, position Vector2) *Enemy
func (em *EnemyManager) UpdateEnemies() // Called every second
```

### Sprite Integration
```gdscript
# client/scripts/gameplay/CharacterSprite.gd
class_name CharacterSprite
extends Sprite2D

enum AnimationState { IDLE, RUN, ATTACK, HIT, DEAD }

func set_state(new_state: AnimationState) -> void
func play_animation(anim_name: String) -> void
```

### Database Schema Updates
```sql
-- New table: items
CREATE TABLE items (
    item_id UUID PRIMARY KEY,
    name VARCHAR(50),
    item_type VARCHAR(20),  -- weapon, armor, potion
    stats JSONB,  -- {damage, defense, heal_amount}
    rarity VARCHAR(20)
);

-- New table: inventory
CREATE TABLE inventory (
    inventory_id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    slot_id INTEGER,
    item_id UUID REFERENCES items(item_id),
    quantity INTEGER
);

-- Update: characters table
ALTER TABLE characters ADD COLUMN current_hp INTEGER DEFAULT 100;
ALTER TABLE characters ADD COLUMN max_hp INTEGER DEFAULT 100;
ALTER TABLE characters ADD COLUMN damage INTEGER DEFAULT 10;
```

---

## 5. Design Specifications (10 min) - @design-team

### Combat Mechanics
- **Warrior**: HP 120, Damage 15, Speed 1.0s
- **Rogue**: HP 80, Damage 20, Speed 0.8s
- **Mage**: HP 70, Damage 25, Speed 1.2s

**Enemy Stats:**
- **Slime**: HP 30, Damage 5
- **Goblin**: HP 50, Damage 8
- **Wolf**: HP 60, Damage 10
- **Skeleton**: HP 80, Damage 12

### Inventory Items
- **Weapon**: Sword (+10 damage), Dagger (+8 damage, faster)
- **Armor**: Shield (+5 defense), Helmet (+3 defense)
- **Potion**: Health Potion (heal 20 HP)

### Zone Designs
- **Village**: Sol vert clair, maisons, fontaine, PNJ (Guard, Merchant)
- **Forêt**: Sol vert foncé, arbres denses, ennemis (Slime, Wolf)
- **Grotte**: Sol gris, rochers, ambiance sombre, ennemis (Skeleton, Goblin)

### NPC Dialogue Examples
**Guard:**
```
"Welcome to the village, traveler!
Be careful in the forest,
slimes and wolves are active today."
```

**Merchant:**
```
"Need supplies? I have
health potions and weapons
for sale!"
```

---

## 6. Risks & Mitigations (5 min)

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Combat trop complexe** | High | Medium | Commencer simple, itérer après MVP |
| **Sprite integration longue** | Medium | Medium | Utiliser assets existants, placeholder si nécessaire |
| **TileMaps prennent trop de temps** | Medium | High | Utiliser tilesets gratuits, simplifier d'abord |
| **Inventory UI trop complexe** | Low | Medium | UI basique drag-drop, améliorer après |
| **Performance avec nombreux ennemis** | Medium | Low | Limité à 10 ennemis max par zone |

---

## 7. Q&A and Clarifications (10 min)

### Questions posées:
1. **@ceo**: "Est-ce que la durée de 12 jours est réaliste?"
   - **Réponse**: Oui, avec l'ajout des sprites, c'est plus réaliste. Sprint 1 a été plus rapide que prévu (3 jours au lieu de 10), donc on a de la marge.

2. **@cto**: "Les assets sprites sont-ils suffisants?"
   - **Réponse**: Oui, tu as déjà `/sprites/Player/` et `/sprites/Enemies/` avec les animations nécessaires. Plus besoin d'assets externes pour le MVP.

3. **@dev-team**: "Quel ordre pour les tasks?"
   - **Réponse**: Enemy spawn → Combat backend → Sprite system → Combat client → Inventory → Zones → NPCs → Chat

4. **@design-team**: "Faut-il plus de PNJ?"
   - **Réponse**: Pour Sprint 2, 3 PNJ (Guard, Merchant, Quest giver) suffisent. Plus dans Sprint 3.

### Clarifications:
- **Daily standup**: Tous les jours à 10h00, 15 min
- **Tracking**: GitHub Projects avec colonne "To Do", "In Progress", "Review", "Done"
- **Code review**: Requis pour toutes les PRs
- **Testing**: Tests unitaires pour backend, tests manuels pour client

---

## 8. Commitment & Next Steps (5 min)

### ✅ Team Commitment
- **@ceo**: Validé le scope et la timeline
- **@cto**: Validé l'architecture technique
- **@dev-team**: Prêt à commencer, toutes les tâches claires
- **@design-team**: Spécifications prêtes, disponible pour questions

### 📋 Action Items

| Action | Owner | Deadline |
|--------|-------|----------|
| Setup GitHub Projects board | @dev-team | Today |
| Create branch `sprint-2` | @dev-team | Today |
| Setup enemy spawn backend | @dev-team | Day 2 |
| Integrate sprite assets | @dev-team | Day 4 |
| Finalize combat specs | @design-team | Today |
| Daily standup setup | @meetings | Today |

### 📅 Schedule
- **Daily Standup**: Tous les jours 10h00 (15 min)
- **Sprint Review**: Day 12, 14h00 (1h)
- **Sprint 3 Planning**: Day 12, 15h30 (1h)

### 🎯 Success Criteria
- Combat system fonctionnel (attack, HP, cooldown)
- Inventory fonctionnel (add, remove, equip)
- 3 zones avec TileMaps
- 3 PNJ avec dialogue
- Chat local + global
- Sprites animés pour joueur et ennemis
- **Aucun bug bloquant pour la démo**

---

## Meeting Notes

**Date**: March 18, 2026  
**Status**: ✅ **Sprint 2 LANCÉ**

**Prochaine étape**: 
1. Setup GitHub Projects board
2. Créer branch `sprint-2`
3. Premier daily standup demain 10h00

**Documents de référence:**
- [SPRINT_2_PLAN.md](./SPRINT_2_PLAN.md)
- [SPRINT_1_REVIEW.md](./SPRINT_1_REVIEW.md)
- [ARCHITECTURE.md](./ARCHITECTURE.md)

---

*Meeting recorded by: @meetings*  
*Next meeting: Daily Standup - March 19, 2026, 10:00*
