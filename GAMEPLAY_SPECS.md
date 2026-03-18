# Gameplay Core Specifications (2D Top-Down)

## Version
- **Date**: 2026-03-17
- **Status**: Draft v2.0 (2D)
- **Author**: Design Team

---

## 1. Système de Combat

### 1.1 Vue d'ensemble
- **Type**: Combat action en temps réel avec éléments tactiques
- **Perspective**: **2D Top-Down** (vue de dessus)
- **Pace**: Rapide mais stratégique (pas de "button mashing")

### 1.2 Vue 2D Top-Down

```
        Nord (Y-)
           ↑
    ↖      |      ↗
西 ———+———  Joueur  ——— Est
    ↙      |      ↘
           ↓
        Sud (Y+)
```

- **Déplacement**: 8 directions (WASD + diagonales)
- **Orientation**: Joueur regarde dans la direction du mouvement ou de la souris
- **Combat**: Targeting automatique ou manuel (clic souris)

### 1.3 Mécaniques de Base

#### Attaques
- **Attaque légère**: Rapide, faible dégât, faible cooldown
- **Attaque lourde**: Lent, fort dégât, knockback, cooldown moyen
- **Attaque spéciale**: Unique par classe, coût en ressource, cooldown long

#### Défense
- **Blocage**: Réduit dégâts de 70%, consomme stamina
- **Parade**: Timing précis (fenêtre 0.3s), annule attaque ennemi, open pour contre-attaque
- **Esquive**: Invincibilité courte (0.5s), consomme stamina, distance fixe

#### Ressources
- **HP**: Points de vie (régénération passive lente en combat, rapide hors combat)
- **Stamina**: Esquive/blocage (régénération 10/s hors combat, 3/s en combat)
- **Ressource de classe**: Mana/Énergie/Rage (variant par classe)

### 1.4 Dégâts & Calculs

```
Dégâts finals = (Dégâts base × Compétence multiplier × Level factor) × (1 - Résistance %) × Crit multiplier
```

- **Critical Hit**: 150% dégâts, taux base 5%
- **Weak Points**: +50% dégâts sur zones vulnérables
- **Elemental Weakness**: +25% dégâts si élément ennemi faible

### 1.5 Status Effects

| Effect | Duration | Description |
|--------|----------|-------------|
| Burn | 5s | 10% dégâts max HP/s |
| Freeze | 2s | Immobilité complète |
| Poison | 8s | 5% dégâts max HP/s + réduit regen de 50% |
| Stun | 1.5s | Incapacité d'agir |
| Slow | 4s | Réduit vitesse de 40% |

---

## 2. Classes de Personnages

### 2.1 Guerrier (Tank/Melee)

**Rôle**: Résister aux dégâts, contrôler les ennemis
**Ressource**: Rage (générée par dégâts reçus/donnés)

#### Stats de Base
- HP: 120%
- Dégâts: 100%
- Vitesse: 90%
- Armor: 150%

#### Compétences

| Nom | Type | Coût | Cooldown | Description |
|-----|------|------|----------|-------------|
| Frappe de Bouclier | Attaque légère | 0 Rage | 0s | Attaque rapide avec knockback léger |
| Fendre | Attaque lourde | 0 Rage | 3s | Dégâts élevés, brise 20% armor ennemi |
| Barricade | Défense | 0 Rage | 15s | Blocage permanent 5s, +50% armor, immunité knockback |
| Vague de Choc | Spéciale | 40 Rage | 30s | AOE 5m, 150% dégâts, stun 1s, knockback fort |
| Rugissement | Buff | 20 Rage | 45s | +30% dégâts équipe, 10s, rayon 10m |

#### Passive
- **Forteresse**: Réduit dégâts critiques reçus de 25%

---

### 2.2 Voleur (DPS/Assassin)

**Rôle**: Dégâts élevés, mobilité, assassinat ciblée
**Ressource**: Énergie (régénération 20/s)

#### Stats de Base
- HP: 80%
- Dégâts: 130%
- Vitesse: 120%
- Armor: 70%

#### Compétences

| Nom | Type | Coût | Cooldown | Description |
|-----|------|------|----------|-------------|
| Frappe Rapide | Attaque légère | 10 Énergie | 0s | 3 coups rapides, +10% dégâts par coup |
| Frappe Arrière | Attaque lourde | 20 Énergie | 4s | Dégâts derrière cible, +50% si backstab |
| Furtivité | Défense | 30 Énergie | 20s | Invisible 6s, premier coup +100% dégâts |
| Téléportation | Mobilité | 25 Énergie | 12s | Teleport 8m, invincible pendant déplacement |
| Poison Mortel | Spéciale | 50 Énergie | 40s | DoT 30% HP max sur 8s, réduit healing de 50% |

#### Passive
- **Prédateur**: +25% dégâts sur cibles avec moins de 30% HP

---

### 2.3 Mage (DPS/Ranged/Control)

**Rôle**: Dégâts à distance, contrôle de foule, burst
**Ressource**: Mana (régénération 15/s, +5/s avec gear)

#### Stats de Base
- HP: 70%
- Dégâts: 150%
- Vitesse: 85%
- Armor: 60%

#### Compétences

| Nom | Type | Coût | Cooldown | Description |
|-----|------|------|----------|-------------|
| Orbe de Feu | Attaque légère | 15 Mana | 0s | Projectile, 100% dégâts feu |
| Glace Nova | Attaque lourde | 35 Mana | 5s | AOE 4m, 120% dégâts glace, slow 40% 3s |
| Miroir Magique | Défense | 40 Mana | 25s | Réfléchit 50% dégâts magiques 4s, 2 charges |
| Tempête | Spéciale | 80 Mana | 50s | Channel 4s, AOE 8m, 200% dégâts, stun 0.5s par tick |
| Polymorphie | CC | 50 Mana | 35s | Transforme ennemi en mouton 6s |

#### Passive
- **Maître des Éléments**: Chaque élément augmente dégâts du suivant de 15% (feu > glace > arcane > feu)

---

## 3. Boucles de Jeu Principales

### 3.1 Boucle Principale (Core Loop)

```
[Combat] → [Loot/XP] → [Upgrade] → [Combat Plus Difficile]
    ↑                                      ↓
    └─────────────── Progression ──────────┘
```

#### Détails de la boucle
1. **Combat**: Affronte ennemis (monsters, PvP, bosses)
2. **Récompense**: Gain XP, loot équipement, ressources
3. **Progression**: Upgrade équipement, apprendre compétences, stats
4. **Défi accru**: Contenu plus difficile avec meilleures récompenses

### 3.2 Boucle de Progression

#### Short-term (minutes)
- Combat ennemis normaux
- Farm ressources crafting
- Compléter quêtes mineures

#### Mid-term (heures)
- Level up personnage (tous les 5-10 min de jeu)
- Upgrader équipement
- Débloquer nouvelles compétences

#### Long-term (jours/semaines)
- Raid bosses
- PvP endgame
- Collection achievements
- Guild wars

### 3.3 Boucle Sociale

```
[Rejoindre Groupe] → [Content Co-op] → [Partage Loot] → [Renforcer Liens]
       ↑                                                        ↓
       └─────────────── Guild Benefits ────────────────────────┘
```

### 3.4 Boucle Économique

- **Crafting**: Collecte ressources → Crafting items → Vente/Équipement
- **Marché**: Achat/vente joueur à joueur
- **Daily**: Quêtes quotidiennes → Ressources premium → Time-savers

---

## 4. Balancing Guidelines

### 4.1 Échelle de Difficulté

| Contenu | HP Ennemi | Dégâts Ennemi | XP Gain | Loot Quality |
|---------|-----------|---------------|---------|--------------|
| Débutant | 100% | 100% | 100% | Common |
| Normal | 150% | 120% | 150% | Uncommon |
| Difficile | 250% | 150% | 250% | Rare |
| Expert | 400% | 200% | 400% | Epic |
| Legendaire | 600% | 250% | 600% | Legendary |

### 4.2 Échelle de Niveau

```
XP requis = Base × (Level ^ 1.5)
Dégâts ennemi = Base × (1 + Level × 0.1)
HP ennemi = Base × (1 + Level × 0.15)
```

### 4.3 Équilibre des Classes (PvP)

- **Guerrier**: Survie élevée, DPS moyen, contrôle bon
- **Voleur**: Burst élevé, survie faible, mobilité excellente
- **Mage**: DPS AOE élevé, survie très faible, contrôle excellent

**Tiers actuel**: Équilibré (à tester en alpha)

### 4.4 Métriques de Suivi

- Win rate par classe en PvP (cible: 45-55%)
- Temps moyen pour tuer boss (cible: 3-10 min selon niveau)
- DPS par seconde par classe (cible: ±15% entre classes)
- Survie moyenne en solo (cible: 15-30 min de contenu)

---

## 5. Interface Utilisateur (Combat) - 2D

### 5.1 HUD Combat (2D Layout)

```
┌─────────────────────────────────────────────────────────────┐
│  [Target Info]                           [Minimap]          │
│  HP ████████░░ 80%                                           │
│  Name: Goblin                                               │
│                                                             │
│                                                             │
│                                                             │
│                                                             │
│                                                             │
│                                                             │
│  [Chat/Notifications]                                       │
│                                                             │
│  [HP]████████░░ [Stamina]██████░░ [Rage]████░░              │
│         [Skill Bar: 1 2 3 4 5]                              │
└─────────────────────────────────────────────────────────────┘
```

- **Barres HP/Stamina/Ressource**: Bas gauche
- **Cooldown skills**: Bas centre (1-5 keys)
- **Minimap**: Haut droite (vue radar 2D)
- **Target info**: Haut gauche
- **Chat/Notifications**: Bas droite

### 5.2 Feedback Visuel 2D

- **Dégâts**: Nombres flottants (blanc normal, jaune crit, rouge weak point)
- **Status Effects**: Icons au-dessus de la tête (2D sprites)
- **Warning**: Cercle rouge au sol pour AOE imminente
- **Hit stop**: Freeze écran 0.1s sur hits critiques
- **Direction Indicators**: Flèches pour ennemis hors écran

### 5.3 Contrôles 2D

| Action | Clavier | Souris |
|--------|---------|--------|
| Déplacement | WASD / Flèches | - |
| Ciblage | - | Clic gauche |
| Attaque | Espace | Clic gauche (cible) |
| Skill 1-5 | 1-5 | - |
| Esquive | Shift | - |
| Blocage | Ctrl | - |
| Camera (pan) | - | Clic droit + drag |
| Menu pause | ESC | - |

---

## 6. Monde 2D

### 6.1 Système de Zones

```
┌─────────────┬─────────────┬─────────────┐
│   Forêt   │   Village   │   Montagne  │
│   Zone 1  │   Zone 2    │   Zone 3    │
│  Lv 1-10  │   Lv 10-20  │   Lv 20-30  │
├─────────────┼─────────────┼─────────────┤
│    Marais  │    Donjon   │    Château  │
│   Zone 4   │   Zone 5    │   Zone 6    │
│  Lv 30-40  │   Lv 40-50  │   Lv 50-60  │
└─────────────┴─────────────┴─────────────┘
```

- **Tile-based**: Chaque zone = TileMap Godot
- **Transitions**: Portails/Portes entre zones
- **Loading**: Chargement instantané ou fade transition

### 6.2 Tile System

| Tile Type | Usage | Collision |
|-----------|-------|-----------|
| Grass | Sol normal | Non |
| Water | Rivière/lac | Oui (impassable) |
| Wall | Bâtiments | Oui |
| Tree | Décor | Oui (optionnel) |
| Door | Transition | Oui (sauf ouvert) |
| Spawn | Point début | Non |

### 6.3 Navigation

- **Pathfinding**: A* sur grid (Godot Navigation2D)
- **Click to move**: Clic souris → path → mouvement auto
- **WASD**: Mouvement direct 8 directions

---

## 7. Notes pour l'Implémentation 2D

### 7.1 Priorités Alpha (2D)
1. **Système de combat de base** (attaques, esquive, blocage)
2. **1 classe jouable** (Guerrier recommandé)
3. **Boucle combat → loot → upgrade**
4. **1 zone de test 2D** (TileMap simple)
5. **Sprite player** (placeholder ou pixel art)

### 7.2 Assets 2D Requis

| Asset | Format | Taille | Priority |
|-------|--------|--------|----------|
| Player sprite sheet | PNG | 32x32 ou 64x64 | Haute |
| Enemy sprite sheets | PNG | 32x32 - 64x64 | Haute |
| Tileset (grass, wall, water) | PNG | 16x16 ou 32x32 | Haute |
| UI elements | PNG | Variable | Haute |
| Particle effects | PNG/GPUParticles | Variable | Moyenne |
| Background music | OGG/MP3 | - | Basse |
| SFX (combat, UI) | WAV/OGG | - | Moyenne |

### 7.3 Godot 2D Setup

```gdscript
# PlayerController2D.gd
extends CharacterBody2D

@onready var sprite: Sprite2D = $Sprite2D
@onready var collision: CollisionShape2D = $CollisionShape2D

var move_speed: float = 200.0
var target_position: Vector2 = Vector2.ZERO

func _physics_process(delta: float) -> void:
    # Get input direction
    var input_dir = Input.get_vector("move_left", "move_right", 
                                      "move_up", "move_down")
    
    # Move character
    velocity = input_dir * move_speed
    move_and_slide()
    
    # Update sprite direction
    if input_dir != Vector2.ZERO:
        sprite.flip_h = input_dir.x < 0
```

### 7.4 Priorités Beta
1. Ajouter 2 classes restantes
2. Système de compétences complet
3. PvP 1v1 (arena 2D)
4. Équilibrage basé sur données

### 7.5 Métriques à Tracker
- Temps moyen entre deaths
- Utilisation des compétences (% time)
- Classe la plus jouée
- Taux de rétention jour 1/7/30

---

## 8. Glossaire

| Terme | Définition |
|-------|------------|
| AOE | Area of Effect - dégâts sur zone |
| CC | Crowd Control - contrôle d'ennemi |
| DoT | Damage over Time - dégâts dans le temps |
| DPS | Damage per Second |
| Knockback | Repousser l'ennemi |
| Backstab | Attaque par derrière |
| Channel | Canaliser une compétence |
| Top-Down | Vue de dessus (2D) |
| TileMap | Grille de tiles pour le monde |
| Sprite | Image 2D pour personnages/objets |

---

## Historique des Versions

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-03-16 | Version initiale 3D |
| 2.0 | 2026-03-17 | **Transition 2D Top-Down** |

---

*Dernière mise à jour: 17 Mars 2026*  
*Propriétaire: @design-team*  
*Perspective: 2D Top-Down*
