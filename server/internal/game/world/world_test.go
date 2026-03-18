package world

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// newTestWorld creates a new World instance for testing with nil database
func newTestWorld() *World {
	return NewWorld(nil)
}

func TestNewWorld(t *testing.T) {
	world := newTestWorld()
	assert.NotNil(t, world)
	assert.NotNil(t, world.players)
	assert.NotNil(t, world.characters)
	assert.NotNil(t, world.zones)
}

func TestCreateCharacter_Success(t *testing.T) {
	world := newTestWorld()

	character, err := world.CreateCharacter("user-123", "TestChar", "warrior")
	assert.NoError(t, err)
	assert.Equal(t, "user-123", character.UserID)
	assert.Equal(t, "TestChar", character.Name)
	assert.Equal(t, "warrior", character.ClassType)
	assert.Equal(t, 1, character.Level)
	assert.Equal(t, int64(0), character.Exp)
	assert.Equal(t, "starter_zone", character.CurrentZone)
}

func TestCreateCharacter_EmptyName(t *testing.T) {
	world := newTestWorld()
	_, err := world.CreateCharacter("user-123", "", "warrior")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "character name is required")
}

func TestCreateCharacter_NameTooLong(t *testing.T) {
	world := newTestWorld()
	longName := "ThisCharacterNameIsWayTooLongAndShouldFailValidation"
	_, err := world.CreateCharacter("user-123", longName, "warrior")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "character name too long")
}

func TestCreateCharacter_InvalidClass(t *testing.T) {
	world := newTestWorld()
	_, err := world.CreateCharacter("user-123", "TestChar", "ninja")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid class type")
}

func TestCreateCharacter_ValidClasses(t *testing.T) {
	world := newTestWorld()

	classes := []string{"warrior", "rogue", "mage"}
	for _, class := range classes {
		char, err := world.CreateCharacter("user-123", "TestChar", class)
		assert.NoError(t, err)
		assert.Equal(t, class, char.ClassType)
	}
}

func TestListCharacters(t *testing.T) {
	world := newTestWorld()

	world.CreateCharacter("user-123", "Char1", "warrior")
	world.CreateCharacter("user-123", "Char2", "mage")
	world.CreateCharacter("user-456", "Char3", "rogue")

	characters := world.ListCharacters("user-123")
	assert.Len(t, characters, 2)
}

func TestListCharacters_Empty(t *testing.T) {
	world := newTestWorld()
	characters := world.ListCharacters("non-existent-user")
	assert.Empty(t, characters)
}

func TestMovePlayer_Success(t *testing.T) {
	world := newTestWorld()

	world.CreateCharacter("user-123", "TestChar", "warrior")

	err := world.MovePlayer("user-123", 5, 3, 0)
	assert.NoError(t, err)

	pos := world.GetPlayerPosition("user-123")
	assert.Equal(t, 5.0, pos.X)
	assert.Equal(t, 3.0, pos.Y)
	assert.Equal(t, 0.0, pos.Z)
}

func TestMovePlayer_NotFound(t *testing.T) {
	world := newTestWorld()
	err := world.MovePlayer("non-existent-user", 10, 5, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player not found")
}

func TestMovePlayer_OutOfBounds_X(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	err := world.MovePlayer("user-123", 200, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")
}

func TestMovePlayer_OutOfBounds_Y(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	err := world.MovePlayer("user-123", 0, 200, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")
}

func TestMovePlayer_OutOfBounds_Z(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	err := world.MovePlayer("user-123", 0, 0, 200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")
}

func TestMovePlayer_SpeedHack(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	err := world.MovePlayer("user-123", 100, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "movement too fast")
}

func TestMovePlayer_MultipleMoves(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	world.MovePlayer("user-123", 3, 0, 0)
	world.MovePlayer("user-123", 5, 0, 0)
	world.MovePlayer("user-123", 5, 3, 0)

	pos := world.GetPlayerPosition("user-123")
	assert.Equal(t, 5.0, pos.X)
	assert.Equal(t, 3.0, pos.Y)
}

func TestValidateMove_Valid(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	player := world.GetPlayer("user-123")
	err := world.ValidateMove(player, 2, 2, 0)
	assert.NoError(t, err)
}

func TestValidateMove_OutOfBounds(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	player := world.GetPlayer("user-123")
	err := world.ValidateMove(player, 200, 0, 0)
	assert.Error(t, err)
}

func TestValidateMove_SpeedHack(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	player := world.GetPlayer("user-123")
	err := world.ValidateMove(player, 100, 0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "movement too fast")
}

func TestGetPlayerPosition(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	pos := world.GetPlayerPosition("user-123")
	assert.Equal(t, 0.0, pos.X)
	assert.Equal(t, 0.0, pos.Y)
	assert.Equal(t, 0.0, pos.Z)
}

func TestGetPlayerPosition_NotFound(t *testing.T) {
	world := newTestWorld()
	pos := world.GetPlayerPosition("non-existent")
	assert.Equal(t, 0.0, pos.X)
	assert.Equal(t, 0.0, pos.Y)
	assert.Equal(t, 0.0, pos.Z)
}

func TestGetPlayer(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	player := world.GetPlayer("user-123")
	assert.NotNil(t, player)
	assert.Equal(t, "user-123", player.UserID)
	assert.Equal(t, "TestChar", player.Username)
}

func TestGetPlayer_NotFound(t *testing.T) {
	world := newTestWorld()
	player := world.GetPlayer("non-existent")
	assert.Nil(t, player)
}

func TestGetZone(t *testing.T) {
	world := newTestWorld()

	zone := world.GetZone("starter_zone")
	assert.NotNil(t, zone)
	assert.Equal(t, "starter_zone", zone.ID)
}

func TestGetZone_NotFound(t *testing.T) {
	world := newTestWorld()
	zone := world.GetZone("non_existent_zone")
	assert.Nil(t, zone)
}

func TestDistanceTo(t *testing.T) {
	v := Vector3{X: 0, Y: 0, Z: 0}
	dist := v.DistanceTo(3, 4, 0)
	assert.Equal(t, 5.0, dist)
}

func TestDistanceTo_ZAxis(t *testing.T) {
	v := Vector3{X: 0, Y: 0, Z: 0}
	dist := v.DistanceTo(0, 0, 5)
	assert.Equal(t, 5.0, dist)
}

func TestPlayerInZone(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	zone := world.GetZone("starter_zone")
	assert.NotNil(t, zone)
	assert.Len(t, zone.Players, 1)
}

func TestMovePlayerByDelta_Success(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	// Move by delta (3, 4) should result in position (3, 4) from origin
	err := world.MovePlayerByDelta("user-123", 3, 4)
	assert.NoError(t, err)

	pos := world.GetPlayerPosition("user-123")
	assert.Equal(t, 3.0, pos.X)
	assert.Equal(t, 4.0, pos.Y)
	assert.Equal(t, 0.0, pos.Z)
}

func TestMovePlayerByDelta_MultipleMoves(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	// First move: (0,0) + (3,0) = (3,0)
	err := world.MovePlayerByDelta("user-123", 3, 0)
	assert.NoError(t, err)

	// Second move: (3,0) + (2,0) = (5,0)
	err = world.MovePlayerByDelta("user-123", 2, 0)
	assert.NoError(t, err)

	// Third move: (5,0) + (0,3) = (5,3)
	err = world.MovePlayerByDelta("user-123", 0, 3)
	assert.NoError(t, err)

	pos := world.GetPlayerPosition("user-123")
	assert.Equal(t, 5.0, pos.X)
	assert.Equal(t, 3.0, pos.Y)
}

func TestMovePlayerByDelta_SpeedHack(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	// Try to move by delta (15, 0) which exceeds max distance of 10
	err := world.MovePlayerByDelta("user-123", 15, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "movement too fast")
}

func TestMovePlayerByDelta_DiagonalSpeedHack(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	// Try to move by delta (8, 8) which has distance ~11.31, exceeds max of 10
	err := world.MovePlayerByDelta("user-123", 8, 8)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "movement too fast")
}

func TestMovePlayerByDelta_MaxValidDelta(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	// Move by delta (6, 8) which has distance exactly 10 (max allowed)
	err := world.MovePlayerByDelta("user-123", 6, 8)
	assert.NoError(t, err)

	pos := world.GetPlayerPosition("user-123")
	assert.Equal(t, 6.0, pos.X)
	assert.Equal(t, 8.0, pos.Y)
}

func TestMovePlayerByDelta_NotFound(t *testing.T) {
	world := newTestWorld()
	err := world.MovePlayerByDelta("non-existent-user", 5, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player not found")
}

func TestMovePlayerByDelta_OutOfBounds(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	// Move to valid position near boundary using multiple small moves
	// Start at (0,0), move to (95, 0) in steps of 10
	for i := 0; i < 9; i++ {
		err := world.MovePlayerByDelta("user-123", 10, 0)
		assert.NoError(t, err)
	}
	// Now at (90, 0)
	err := world.MovePlayerByDelta("user-123", 5, 0)
	assert.NoError(t, err)
	// Now at (95, 0)

	// Try to move further which would go out of bounds (max X is 100)
	err = world.MovePlayerByDelta("user-123", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")
}

func TestMovePlayerByDelta_NegativeDelta(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	// Move positive first
	err := world.MovePlayerByDelta("user-123", 5, 5)
	assert.NoError(t, err)

	// Move negative delta
	err = world.MovePlayerByDelta("user-123", -2, -3)
	assert.NoError(t, err)

	pos := world.GetPlayerPosition("user-123")
	assert.Equal(t, 3.0, pos.X)
	assert.Equal(t, 2.0, pos.Y)
}

func TestValidateMoveDelta_Valid(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	player := world.GetPlayer("user-123")
	err := world.ValidateMoveDelta(player, 3, 4)
	assert.NoError(t, err)
}

func TestValidateMoveDelta_SpeedHack(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	player := world.GetPlayer("user-123")
	err := world.ValidateMoveDelta(player, 15, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "movement too fast")
}

func TestValidateMoveDelta_ZeroDelta(t *testing.T) {
	world := newTestWorld()
	world.CreateCharacter("user-123", "TestChar", "warrior")

	player := world.GetPlayer("user-123")
	err := world.ValidateMoveDelta(player, 0, 0)
	assert.NoError(t, err)
}
