# Décision Architecture: Transition vers 2D Top-Down

**Decision**: ADR-004  
**Date**: 2026-03-17  
**Status**: Approved  
**Author**: CEO + CTO + Design Team  
**Approbation**: @ceo

---

## Contexte

Le projet MMO Fantastique avait initialement été conçu en 3D (3ème personne). Après analyse des contraintes techniques et du scope pour le MVP, nous décidons de passer en **2D Top-Down** pour le Sprint 1 et la phase Alpha.

---

## Décision

**Transition vers 2D Top-Down** pour le client Godot.

### Ce qui change:
- **Perspective**: 3D → 2D Top-Down (vue de dessus)
- **Nodes Godot**: Node3D → Node2D/Control
- **Camera**: Camera3D → Camera2D (ou orthographique)
- **Player**: CharacterBody3D → CharacterBody2D
- **Positions**: Vector3 (x, y, z) → Vector2 (x, y) + orientation
- **Assets**: Modèles 3D → Sprites 2D + TileMaps

### Ce qui reste identique:
- **Backend Go**: Architecture serveur inchangée
- **Database**: Schema adapté (position_z supprimé)
- **Protocol**: ProtoBuf definitions adaptées (2D messages)
- **Gameplay**: Classes, compétences, combat restent les mêmes

---

## Rationales

### 1. Développement plus rapide ✅

**3D**:
- Modélisation 3D complexe
- Rigging + animations 3D
- Textures PBR
- Temps: 3-4 semaines par personnage

**2D**:
- Sprites pixel art ou vectoriels
- Animations 2D (sprite sheets)
- Temps: 3-5 jours par personnage

**Gain**: **~70% de temps économisé** sur les assets

### 2. Assets plus accessibles ✅

**3D**:
- Modèles 3D coûteux ($100-500/unité)
- Peu de assets gratuits de qualité
- Nécessite artiste 3D spécialisé

**2D**:
- Sprites 2D abordables ($10-50/unité)
- Beaucoup de assets gratuits/cheap
- Asset stores riches (Itch.io, OpenGameArt)
- Peut être fait par artiste 2D ou pixel artist

**Gain**: **Budget divisé par 10**, plus de choix

### 3. Gameplay plus clair ✅

**3D**:
- Caméra peut obstruer la vue
- Difficile voir les ennemis autour
- Navigation parfois confuse

**2D Top-Down**:
- Vue complète du champ de bataille
- Tous les ennemis visibles
- Navigation intuitive
- Meilleur pour combat tactique

**Gain**: **Expérience joueur améliorée** pour le combat

### 4. Performance ✅

**3D**:
- GPU intensif
- Limite ~50-100 joueurs/zone
- Nécessite hardware puissant

**2D**:
- GPU léger (sprites 2D)
- Peut supporter ~500+ joueurs/zone
- Fonctionne sur machines modestes

**Gain**: **Scalabilité améliorée**, audience plus large

### 5. Scope réduit pour MVP ✅

**3D**:
- Beaucoup de contenu nécessaire pour paraître "fini"
- Environnements 3D complexes
- Animations nombreuses

**2D**:
- MVP plus petit acceptable
- TileMaps rapides à créer
- Animations 2D plus simples

**Gain**: **Time-to-market réduit de 3-4 mois**

---

## Conséquences

### Positive

- ✅ Développement 70% plus rapide
- ✅ Budget assets divisé par 10
- ✅ Meilleure scalabilité serveur
- ✅ Gameplay tactique amélioré
- ✅ MVP plus rapide à valider
- ✅ Audience plus large (PC faibles)

### Negative

- ⚠️ Perte de l'immersion 3D
- ⚠️ Moins "wow factor" visuel
- ⚠️ Nécessite réapprendre certains concepts
- ⚠️ Certains joueurs préfèrent 3D

### Neutral

- ↔️ Backend inchangé (Go)
- ↔️ Database schema adapté (minime)
- ↔️ Gameplay core identique

---

## Impact Technique

### Client Godot

**Fichiers à modifier/créer**:
```
client/
├── scenes/
│   ├── main/Main.tscn           # Node3D → Node2D
│   └── ui/menus/*.tscn          # Control nodes (déjà 2D)
├── scripts/
│   ├── gameplay/
│   │   ├── PlayerController.gd  # → PlayerController2D.gd
│   │   └── CharacterBody3D → CharacterBody2D
│   └── network/NetworkManager.gd # Pas de changement
└── assets/
    ├── sprites/                 # Nouveau (2D)
    ├── tilesets/                # Nouveau
    └── models/                  # Supprimé (3D)
```

**Changements code**:
```gdscript
# AVANT (3D)
extends CharacterBody3D
var velocity: Vector3
func _physics_process(delta):
    move_and_slide()

# APRÈS (2D)
extends CharacterBody2D
var velocity: Vector2
func _physics_process(delta):
    move_and_slide()
```

### Database

**Schema change**:
```sql
-- AVANT (3D)
position_x FLOAT,
position_y FLOAT,
position_z FLOAT,
orientation FLOAT

-- APRÈS (2D)
position_x FLOAT,
position_y FLOAT,
orientation FLOAT  -- Rotation 0-360°
```

**Migration requise**:
```sql
ALTER TABLE characters DROP COLUMN position_z;
```

### Protocol Buffers

**Messages à mettre à jour**:
```protobuf
// AVANT
message Position3D {
  float x = 1;
  float y = 2;
  float z = 3;
}

// APRÈS
message Position2D {
  float x = 1;
  float y = 2;
}
```

---

## Plan de Migration

### Phase 1: Infrastructure (Jour 1-2)
- [ ] Mettre à jour ARCHITECTURE.md
- [ ] Mettre à jour GAMEPLAY_SPECS.md
- [ ] Mettre à jour SPRINT_1_PLAN.md
- [ ] Créer ce document (ADR-004)

### Phase 2: Database (Jour 2-3)
- [ ] Créer migration SQL (drop position_z)
- [ ] Appliquer migration locale
- [ ] Mettre à jour models Go

### Phase 3: Client (Jour 3-7)
- [ ] Convertir Main.tscn (Node3D → Node2D)
- [ ] Créer PlayerController2D.gd
- [ ] Convertir zones de test (3D → TileMap 2D)
- [ ] Ajouter sprites placeholders

### Phase 4: Testing (Jour 8-10)
- [ ] Tester movement 2D
- [ ] Tester UI 2D
- [ ] Valider gameplay core
- [ ] Demo sprint 1

---

## Validation

**CEO**: ✅ Approuvé  
**CTO**: ✅ Approuvé  
**Design Team**: ✅ Approuvé  
**Dev Team**: ✅ Notifié

---

## Références

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Architecture mise à jour 2D
- [GAMEPLAY_SPECS.md](./GAMEPLAY_SPECS.md) - Specs gameplay 2D
- [SPRINT_1_PLAN.md](./SPRINT_1_PLAN.md) - Plan sprint ajusté
- [Godot 2D Documentation](https://docs.godotengine.org/en/stable/tutorials/2d/index.html)

---

*Dernière mise à jour: 17 Mars 2026*  
*Propriétaire: @cto*
