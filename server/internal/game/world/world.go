package world

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/ffreville/mmo-team-test/server/internal/database"
	"github.com/ffreville/mmo-team-test/server/pkg/models"
	"github.com/google/uuid"
)

type World struct {
	db        *database.PostgresDB
	players   map[string]*Player
	playersMu sync.RWMutex

	characters   map[string]*models.Character
	charactersMu sync.RWMutex

	zones  map[string]*Zone
	bounds Bounds
}

type Player struct {
	ID          string
	UserID      string
	Username    string
	CharacterID string
	Position    Vector3
	Orientation float64
	IsMoving    bool
	LastUpdate  time.Time
	ZoneID      string
}

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

type Bounds struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
	MinZ float64
	MaxZ float64
}

type Zone struct {
	ID      string
	Players map[string]*Player
	Bounds  Bounds
}

func NewWorld(db *database.PostgresDB) *World {
	return &World{
		db:         db,
		players:    make(map[string]*Player),
		characters: make(map[string]*models.Character),
		zones: map[string]*Zone{
			"starter_zone": {
				ID:      "starter_zone",
				Players: make(map[string]*Player),
				Bounds: Bounds{
					MinX: -1000, MaxX: 1000,
					MinY: -1000, MaxY: 1000,
					MinZ: -100, MaxZ: 100,
				},
			},
		},
		bounds: Bounds{
			MinX: -1000, MaxX: 1000,
			MinY: -1000, MaxY: 1000,
			MinZ: -100, MaxZ: 100,
		},
	}
}

func (w *World) CreateCharacter(userID, name, classType string) (*models.Character, error) {
	if name == "" {
		return nil, errors.New("character name is required")
	}

	if len(name) > 50 {
		return nil, errors.New("character name too long (max 50 characters)")
	}

	validClasses := map[string]bool{
		"warrior": true,
		"rogue":   true,
		"mage":    true,
	}

	if !validClasses[classType] {
		return nil, fmt.Errorf("invalid class type: %s", classType)
	}

	w.charactersMu.Lock()
	defer w.charactersMu.Unlock()

	characterID := uuid.New().String()

	character := &models.Character{
		CharacterID: characterID,
		UserID:      userID,
		Name:        name,
		ClassType:   classType,
		Level:       1,
		Exp:         0,
		CurrentZone: "starter_zone",
		PositionX:   0,
		PositionY:   0,
		PositionZ:   0,
		Orientation: 0,
		IsOnline:    false,
		CreatedAt:   time.Now(),
	}

	w.characters[characterID] = character

	// Persist character to PostgreSQL if database is available
	if w.db != nil {
		ctx := context.Background()
		var dbCharacterID string
		err := w.db.Pool().QueryRow(ctx,
			"INSERT INTO characters (character_id, user_id, name, class_type, level, exp, current_zone, position_x, position_y, position_z, orientation, is_online, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING character_id",
			characterID, userID, name, classType, 1, 0, "starter_zone", 0.0, 0.0, 0.0, 0.0, false, time.Now()).Scan(&dbCharacterID)
		if err != nil {
			fmt.Printf("Warning: failed to persist character to database: %v\n", err)
			// Continue with in-memory storage even if DB insert fails
		}
	}

	player := &Player{
		ID:          characterID,
		UserID:      userID,
		CharacterID: characterID,
		Username:    name,
		Position:    Vector3{X: 0, Y: 0, Z: 0},
		ZoneID:      "starter_zone",
		LastUpdate:  time.Now(),
	}

	w.players[userID] = player
	w.zones["starter_zone"].Players[characterID] = player

	return character, nil
}

func (w *World) ListCharacters(userID string) []*models.Character {
	log.Printf("[DEBUG] ListCharacters called for userID: %s", userID)
	log.Printf("[DEBUG] World db pointer: %v", w.db)

	ctx := context.Background()

	var characters []*models.Character

	if w.db != nil {
		log.Printf("[DEBUG] Database is connected, executing query for user: %s", userID)

		// Debug: Check total characters in DB
		var totalCount int
		err := w.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM characters").Scan(&totalCount)
		if err != nil {
			log.Printf("[ERROR] Failed to count characters: %v", err)
		} else {
			log.Printf("[DEBUG] Total characters in database: %d", totalCount)
		}

		// Debug: Check characters for this user
		var userCount int
		err = w.db.Pool().QueryRow(ctx, "SELECT COUNT(*) FROM characters WHERE user_id = $1", userID).Scan(&userCount)
		if err != nil {
			log.Printf("[ERROR] Failed to count user characters: %v", err)
		} else {
			log.Printf("[DEBUG] Characters for user %s: %d", userID, userCount)
		}

		rows, err := w.db.Pool().Query(ctx,
			"SELECT character_id, user_id, name, class_type, level, exp, current_zone, position_x, position_y, position_z, orientation, is_online, created_at FROM characters WHERE user_id = $1",
			userID)
		if err != nil {
			log.Printf("[ERROR] Failed to load characters from database: %v", err)
		} else {
			defer rows.Close()
			count := 0
			for rows.Next() {
				var char models.Character
				err := rows.Scan(
					&char.CharacterID, &char.UserID, &char.Name, &char.ClassType,
					&char.Level, &char.Exp, &char.CurrentZone, &char.PositionX,
					&char.PositionY, &char.PositionZ, &char.Orientation, &char.IsOnline,
					&char.CreatedAt)
				if err != nil {
					log.Printf("[ERROR] Failed to scan character: %v", err)
					continue
				}
				log.Printf("[DEBUG] Successfully loaded character: %s (%s)", char.Name, char.CharacterID)
				characters = append(characters, &char)
				count++
			}
			log.Printf("[DEBUG] Query returned %d characters from database", count)
		}
	} else {
		log.Printf("[WARN] Database is nil, skipping database query")
	}

	if len(characters) == 0 {
		log.Printf("[DEBUG] No characters from DB, checking in-memory storage for userID: %s", userID)
		w.charactersMu.RLock()
		memCount := 0
		for _, char := range w.characters {
			if char.UserID == userID {
				log.Printf("[DEBUG] Found in-memory character: %s (%s)", char.Name, char.CharacterID)
				characters = append(characters, char)
				memCount++
			}
		}
		w.charactersMu.RUnlock()
		log.Printf("[DEBUG] Found %d characters in memory", memCount)
	} else {
		log.Printf("[DEBUG] Returning %d characters from database", len(characters))
	}

	log.Printf("[DEBUG] ListCharacters returning %d total characters", len(characters))
	return characters
}

func (w *World) MovePlayer(userID string, targetX, targetY, targetZ float64) error {
	w.playersMu.RLock()
	player, ok := w.players[userID]
	w.playersMu.RUnlock()

	if !ok {
		return errors.New("player not found")
	}

	if err := w.ValidateMove(player, targetX, targetY, targetZ); err != nil {
		return err
	}

	w.playersMu.Lock()
	player.Position.X = targetX
	player.Position.Y = targetY
	player.Position.Z = targetZ
	player.IsMoving = false
	player.LastUpdate = time.Now()
	w.playersMu.Unlock()

	return nil
}

func (w *World) MovePlayerByDelta(userID string, deltaX, deltaY float64) error {
	w.playersMu.RLock()
	player, ok := w.players[userID]
	w.playersMu.RUnlock()

	if !ok {
		return errors.New("player not found")
	}

	// Calculate new position based on delta
	newX := player.Position.X + deltaX
	newY := player.Position.Y + deltaY
	newZ := player.Position.Z

	if err := w.ValidateMoveDelta(player, deltaX, deltaY); err != nil {
		return err
	}

	// Check bounds after calculating new position
	zone, ok := w.zones[player.ZoneID]
	if !ok {
		return errors.New("player zone not found")
	}

	if newX < zone.Bounds.MinX || newX > zone.Bounds.MaxX {
		return errors.New("position out of bounds (X)")
	}

	if newY < zone.Bounds.MinY || newY > zone.Bounds.MaxY {
		return errors.New("position out of bounds (Y)")
	}

	if newZ < zone.Bounds.MinZ || newZ > zone.Bounds.MaxZ {
		return errors.New("position out of bounds (Z)")
	}

	w.playersMu.Lock()
	player.Position.X = newX
	player.Position.Y = newY
	player.Position.Z = newZ
	player.IsMoving = false
	player.LastUpdate = time.Now()
	w.playersMu.Unlock()

	return nil
}

func (w *World) ValidateMove(player *Player, targetX, targetY, targetZ float64) error {
	zone, ok := w.zones[player.ZoneID]
	if !ok {
		return errors.New("player zone not found")
	}

	if targetX < zone.Bounds.MinX || targetX > zone.Bounds.MaxX {
		return errors.New("position out of bounds (X)")
	}

	if targetY < zone.Bounds.MinY || targetY > zone.Bounds.MaxY {
		return errors.New("position out of bounds (Y)")
	}

	if targetZ < zone.Bounds.MinZ || targetZ > zone.Bounds.MaxZ {
		return errors.New("position out of bounds (Z)")
	}

	distance := player.Position.DistanceTo(targetX, targetY, targetZ)
	maxDistance := 10.0

	if distance > maxDistance {
		return fmt.Errorf("movement too fast (distance: %.2f, max: %.2f)", distance, maxDistance)
	}

	return nil
}

func (w *World) ValidateMoveDelta(player *Player, deltaX, deltaY float64) error {
	// Calculate the distance of the delta movement
	deltaDistance := math.Sqrt(deltaX*deltaX + deltaY*deltaY)
	maxDistance := 50.0 // TEMPORARILY INCREASED FOR DEBUGGING

	log.Printf("[DEBUG] ValidateMoveDelta - deltaDistance: %.4f, maxDistance: %.4f", deltaDistance, maxDistance)

	if deltaDistance > maxDistance {
		return fmt.Errorf("movement too fast (distance: %.2f, max: %.2f)", deltaDistance, maxDistance)
	}

	return nil
}

func (w *World) GetPlayerPosition(userID string) Vector3 {
	w.playersMu.RLock()
	defer w.playersMu.RUnlock()

	if player, ok := w.players[userID]; ok {
		return player.Position
	}

	return Vector3{X: 0, Y: 0, Z: 0}
}

func (w *World) GetPlayer(userID string) *Player {
	w.playersMu.RLock()
	defer w.playersMu.RUnlock()
	return w.players[userID]
}

func (w *World) GetZone(zoneID string) *Zone {
	return w.zones[zoneID]
}

func (v Vector3) DistanceTo(x, y, z float64) float64 {
	dx := v.X - x
	dy := v.Y - y
	dz := v.Z - z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// GetCharacter retrieves a character by ID from memory or database
func (w *World) GetCharacter(characterID string) *models.Character {
	// First check in-memory storage
	w.charactersMu.RLock()
	char, ok := w.characters[characterID]
	w.charactersMu.RUnlock()

	if ok {
		return char
	}

	// If not in memory, try to load from database
	if w.db != nil {
		ctx := context.Background()
		var char models.Character
		err := w.db.Pool().QueryRow(ctx,
			"SELECT character_id, user_id, name, class_type, level, exp, current_zone, position_x, position_y, position_z, orientation, is_online, created_at FROM characters WHERE character_id = $1",
			characterID).Scan(
			&char.CharacterID, &char.UserID, &char.Name, &char.ClassType,
			&char.Level, &char.Exp, &char.CurrentZone, &char.PositionX,
			&char.PositionY, &char.PositionZ, &char.Orientation, &char.IsOnline,
			&char.CreatedAt)
		if err != nil {
			log.Printf("Warning: failed to load character from database: %v", err)
			return nil
		}
		return &char
	}

	return nil
}

// CreatePlayerEntry creates a new player entry in the World.players map
func (w *World) CreatePlayerEntry(userID, characterID, username string, posX, posY, posZ float64) error {
	w.playersMu.Lock()
	defer w.playersMu.Unlock()

	// Check if player already exists
	if _, ok := w.players[userID]; ok {
		log.Printf("Player %s already exists in world", userID)
		return nil // Already exists, not an error
	}

	// Get character to ensure it exists
	char := w.GetCharacter(characterID)
	if char == nil {
		return fmt.Errorf("character %s not found", characterID)
	}

	player := &Player{
		ID:          characterID,
		UserID:      userID,
		CharacterID: characterID,
		Username:    username,
		Position:    Vector3{X: posX, Y: posY, Z: posZ},
		ZoneID:      "starter_zone",
		LastUpdate:  time.Now(),
	}

	w.players[userID] = player

	// Add player to zone
	if zone, ok := w.zones["starter_zone"]; ok {
		zone.Players[characterID] = player
	}

	log.Printf("Player %s created in world at position (%.2f, %.2f, %.2f)",
		userID, posX, posY, posZ)
	return nil
}

// RemovePlayer removes a player from the World.players map and their zone
func (w *World) RemovePlayer(userID string) {
	w.playersMu.Lock()
	defer w.playersMu.Unlock()

	player, ok := w.players[userID]
	if !ok {
		return
	}

	// Remove from zone
	if zone, ok := w.zones[player.ZoneID]; ok {
		delete(zone.Players, player.CharacterID)
	}

	delete(w.players, userID)
	log.Printf("Player %s removed from world", userID)
}
