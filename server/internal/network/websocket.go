package network

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/ffreville/mmo-team-test/server/internal/auth"
	"github.com/ffreville/mmo-team-test/server/internal/database"
	"github.com/ffreville/mmo-team-test/server/internal/game/world"
	"github.com/gorilla/websocket"
)

const (
	WriteWait      = 10 * time.Second
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 512
)

type Gateway struct {
	clients   map[string]*Client
	clientsMu sync.RWMutex
	auth      *auth.AuthService
	world     *world.World
	redis     *database.RedisClient
	upgrader  websocket.Upgrader
}

type Client struct {
	conn        *websocket.Conn
	gateway     *Gateway
	userID      string
	username    string
	characterID string
	send        chan []byte
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type AuthPayload struct {
	Token string `json:"token"`
}

type MovePayload struct {
	Timestamp int64   `json:"timestamp"`
	TargetX   float64 `json:"target_x"`
	TargetY   float64 `json:"target_y"`
	TargetZ   float64 `json:"target_z"`
}

type MoveDeltaPayload struct {
	Timestamp int64   `json:"timestamp"`
	DeltaX    float64 `json:"delta_x"`
	DeltaY    float64 `json:"delta_y"`
}

func NewGateway(authService *auth.AuthService, gameWorld *world.World, redis *database.RedisClient) *Gateway {
	return &Gateway{
		clients: make(map[string]*Client),
		auth:    authService,
		world:   gameWorld,
		redis:   redis,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

func (g *Gateway) HandleConnection(conn *websocket.Conn) {
	client := &Client{
		conn:    conn,
		gateway: g,
		send:    make(chan []byte, 256),
	}

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.gateway.unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.gateway.handleMessage(c, message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (g *Gateway) handleMessage(client *Client, message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	switch msg.Type {
	case "auth_login":
		g.handleAuthLogin(client, msg)
	case "auth_register":
		g.handleAuthRegister(client, msg)
	case "auth":
		g.handleAuth(client, msg)
	case "enter_world":
		g.handleEnterWorld(client, msg)
	case "move_command":
		g.handleMove(client, msg)
	case "move_command_2d":
		g.handleMove2D(client, msg)
	case "move_command_2d_delta":
		g.handleMove2DDelta(client, msg)
	case "character_create":
		g.handleCharacterCreate(client, msg)
	case "character_list":
		g.handleCharacterList(client, msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (g *Gateway) handleAuth(client *Client, msg Message) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	token, ok := payload["token"].(string)
	if !ok {
		g.sendError(client, "missing token")
		return
	}

	claims, err := g.auth.ValidateToken(token)
	if err != nil {
		g.sendError(client, "invalid token")
		return
	}

	client.userID = claims.UserID
	client.username = claims.Username

	g.register(client)

	g.sendMessage(client, "auth_response", map[string]interface{}{
		"success":   true,
		"player_id": claims.UserID,
		"message":   "Authenticated successfully",
		"username":  claims.Username,
	})
}

func (g *Gateway) handleAuthLogin(client *Client, msg Message) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	requestID := float64(0)
	if rid, ok := payload["request_id"].(float64); ok {
		requestID = rid
	}

	username, ok := payload["username"].(string)
	if !ok {
		g.sendError(client, "missing username")
		return
	}

	password, ok := payload["password"].(string)
	if !ok {
		g.sendError(client, "missing password")
		return
	}

	token, err := g.auth.Login(username, password)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
		if requestID != 0 {
			response["request_id"] = requestID
		}
		g.sendMessage(client, "auth_response", response)
		return
	}

	claims, _ := g.auth.ValidateToken(token)
	client.userID = claims.UserID
	client.username = claims.Username

	g.register(client)

	response := map[string]interface{}{
		"success":   true,
		"token":     token,
		"player_id": claims.UserID,
		"message":   "Login successful",
		"username":  claims.Username,
	}
	if requestID != 0 {
		response["request_id"] = requestID
	}

	g.sendMessage(client, "auth_response", response)
}

func (g *Gateway) handleAuthRegister(client *Client, msg Message) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	requestID := float64(0)
	if rid, ok := payload["request_id"].(float64); ok {
		requestID = rid
	}

	username, ok := payload["username"].(string)
	if !ok {
		g.sendError(client, "missing username")
		return
	}

	email, ok := payload["email"].(string)
	if !ok {
		g.sendError(client, "missing email")
		return
	}

	password, ok := payload["password"].(string)
	if !ok {
		g.sendError(client, "missing password")
		return
	}

	user, err := g.auth.Register(username, email, password)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
		if requestID != 0 {
			response["request_id"] = requestID
		}
		g.sendMessage(client, "auth_response", response)
		return
	}
	_ = user

	token, err := g.auth.Login(username, password)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "registration succeeded but login failed",
		}
		if requestID != 0 {
			response["request_id"] = requestID
		}
		g.sendMessage(client, "auth_response", response)
		return
	}

	claims, _ := g.auth.ValidateToken(token)
	client.userID = claims.UserID
	client.username = claims.Username

	g.register(client)

	response := map[string]interface{}{
		"success":   true,
		"token":     token,
		"player_id": claims.UserID,
		"message":   "Registration successful",
		"username":  claims.Username,
	}
	if requestID != 0 {
		response["request_id"] = requestID
	}

	g.sendMessage(client, "auth_response", response)
}

// handleEnterWorld creates a player entry on the server when a client enters the world
func (g *Gateway) handleEnterWorld(client *Client, msg Message) {
	if client.userID == "" {
		g.sendError(client, "not authenticated")
		return
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	characterID := ""
	if cid, ok := payload["character_id"].(string); ok {
		characterID = cid
	}

	if characterID == "" {
		g.sendError(client, "missing character_id")
		return
	}

	// Store character ID in client
	client.characterID = characterID

	// Get character data from world
	character := g.world.GetCharacter(characterID)
	if character == nil {
		g.sendError(client, "character not found")
		return
	}

	// Create player entry in World.players
	err := g.world.CreatePlayerEntry(client.userID, characterID, character.Name, character.PositionX, character.PositionY, character.PositionZ)
	if err != nil {
		log.Printf("Error creating player entry: %v", err)
		g.sendError(client, "failed to create player entry")
		return
	}

	// Get player position (should be the character's position)
	pos := g.world.GetPlayerPosition(client.userID)

	// Send success response with initial position
	g.sendMessage(client, "enter_world_response", map[string]interface{}{
		"success":      true,
		"player_id":    client.userID,
		"character_id": characterID,
		"server_x":     pos.X,
		"server_y":     pos.Y,
		"server_z":     pos.Z,
		"message":      "Welcome to the world!",
	})

	log.Printf("Player %s (%s) entered world at position (%.2f, %.2f, %.2f)",
		client.username, characterID, pos.X, pos.Y, pos.Z)
}

func (g *Gateway) handleMove(client *Client, msg Message) {
	if client.userID == "" {
		g.sendError(client, "not authenticated")
		return
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	timestamp := int64(0)
	if t, ok := payload["timestamp"].(float64); ok {
		timestamp = int64(t)
	}

	targetX := 0.0
	if x, ok := payload["target_x"].(float64); ok {
		targetX = x
	}

	targetY := 0.0
	if y, ok := payload["target_y"].(float64); ok {
		targetY = y
	}

	targetZ := 0.0
	if z, ok := payload["target_z"].(float64); ok {
		targetZ = z
	}

	err := g.world.MovePlayer(client.userID, targetX, targetY, targetZ)
	if err != nil {
		g.sendError(client, err.Error())
		return
	}

	pos := g.world.GetPlayerPosition(client.userID)

	g.sendMessage(client, "move_response", map[string]interface{}{
		"success":   true,
		"server_x":  pos.X,
		"server_y":  pos.Y,
		"server_z":  pos.Z,
		"timestamp": timestamp,
	})

	g.broadcastToZone(client, "player_moved", map[string]interface{}{
		"player_id": client.userID,
		"x":         pos.X,
		"y":         pos.Y,
		"z":         pos.Z,
	})
}

func (g *Gateway) handleMove2D(client *Client, msg Message) {
	if client.userID == "" {
		g.sendError(client, "not authenticated")
		return
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	timestamp := int64(0)
	if t, ok := payload["timestamp"].(float64); ok {
		timestamp = int64(t)
	}

	targetX := 0.0
	if x, ok := payload["target_x"].(float64); ok {
		targetX = x
	}

	targetY := 0.0
	if y, ok := payload["target_y"].(float64); ok {
		targetY = y
	}

	err := g.world.MovePlayer(client.userID, targetX, targetY, 0)
	if err != nil {
		g.sendError(client, err.Error())
		return
	}

	pos := g.world.GetPlayerPosition(client.userID)

	g.sendMessage(client, "move_response", map[string]interface{}{
		"success":   true,
		"server_x":  pos.X,
		"server_y":  pos.Y,
		"timestamp": timestamp,
	})

	g.broadcastToZone(client, "player_moved", map[string]interface{}{
		"player_id": client.userID,
		"x":         pos.X,
		"y":         pos.Y,
	})
}

func (g *Gateway) handleMove2DDelta(client *Client, msg Message) {
	log.Printf("[DEBUG] handleMove2DDelta called for client: %s", client.username)

	if client.userID == "" {
		g.sendError(client, "not authenticated")
		return
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	timestamp := int64(0)
	if t, ok := payload["timestamp"].(float64); ok {
		timestamp = int64(t)
	}

	deltaX := 0.0
	if dx, ok := payload["delta_x"].(float64); ok {
		deltaX = dx
	}

	deltaY := 0.0
	if dy, ok := payload["delta_y"].(float64); ok {
		deltaY = dy
	}

	// Debug: Log received delta values
	log.Printf("[DEBUG] Received delta movement - deltaX: %.4f, deltaY: %.4f, timestamp: %d", deltaX, deltaY, timestamp)

	// Calculate delta distance for logging
	deltaDistance := math.Sqrt(deltaX*deltaX + deltaY*deltaY)
	log.Printf("[DEBUG] Delta distance: %.4f", deltaDistance)

	err := g.world.MovePlayerByDelta(client.userID, deltaX, deltaY)
	if err != nil {
		log.Printf("[ERROR] MovePlayerByDelta failed: %v", err)
		g.sendError(client, err.Error())
		return
	}

	log.Printf("[DEBUG] MovePlayerByDelta succeeded for client: %s", client.username)

	pos := g.world.GetPlayerPosition(client.userID)

	g.sendMessage(client, "move_response", map[string]interface{}{
		"success":   true,
		"server_x":  pos.X,
		"server_y":  pos.Y,
		"timestamp": timestamp,
	})

	g.broadcastToZone(client, "player_moved", map[string]interface{}{
		"player_id": client.userID,
		"x":         pos.X,
		"y":         pos.Y,
	})
}

func (g *Gateway) handleCharacterCreate(client *Client, msg Message) {
	if client.userID == "" {
		g.sendError(client, "not authenticated")
		return
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		g.sendError(client, "invalid payload")
		return
	}

	name := ""
	if n, ok := payload["name"].(string); ok {
		name = n
	}

	classType := "warrior"
	if c, ok := payload["class_type"].(string); ok {
		classType = c
	}

	character, err := g.world.CreateCharacter(client.userID, name, classType)
	if err != nil {
		g.sendError(client, err.Error())
		return
	}

	client.characterID = character.CharacterID

	g.sendMessage(client, "character_create_response", map[string]interface{}{
		"success": true,
		"character": map[string]interface{}{
			"character_id": character.CharacterID,
			"name":         character.Name,
			"class_type":   character.ClassType,
			"level":        character.Level,
			"position_x":   character.PositionX,
			"position_y":   character.PositionY,
			"position_z":   character.PositionZ,
		},
		"message": "Character created successfully",
	})
}

func (g *Gateway) handleCharacterList(client *Client, msg Message) {
	log.Printf("[DEBUG] handleCharacterList called for client userID: %s", client.userID)
	if client.userID == "" {
		log.Printf("[ERROR] handleCharacterList: client not authenticated, userID is empty")
		g.sendError(client, "not authenticated")
		return
	}

	log.Printf("[DEBUG] Calling ListCharacters with userID: %s", client.userID)
	characters := g.world.ListCharacters(client.userID)
	log.Printf("[DEBUG] ListCharacters returned %d characters", len(characters))

	charactersData := make([]map[string]interface{}, len(characters))
	for i, char := range characters {
		charactersData[i] = map[string]interface{}{
			"character_id": char.CharacterID,
			"name":         char.Name,
			"class_type":   char.ClassType,
			"level":        char.Level,
			"position_x":   char.PositionX,
			"position_y":   char.PositionY,
			"position_z":   char.PositionZ,
		}
	}

	g.sendMessage(client, "character_list_response", map[string]interface{}{
		"success":    true,
		"characters": charactersData,
	})
}

func (g *Gateway) register(client *Client) {
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()
	g.clients[client.userID] = client
	log.Printf("Client connected: %s", client.username)
}

func (g *Gateway) unregister(client *Client) {
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()
	if _, ok := g.clients[client.userID]; ok {
		// Remove player from world if exists
		if client.userID != "" {
			g.world.RemovePlayer(client.userID)
		}
		delete(g.clients, client.userID)
		close(client.send)
		log.Printf("Client disconnected: %s", client.username)
	}
}

func (g *Gateway) sendMessage(client *Client, msgType string, payload interface{}) {
	message := Message{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		log.Printf("Client send buffer full")
	}
}

func (g *Gateway) sendError(client *Client, message string) {
	g.sendMessage(client, "error", map[string]interface{}{
		"message": message,
	})
}

func (g *Gateway) broadcastToZone(exclude *Client, msgType string, payload interface{}) {
	g.clientsMu.RLock()
	defer g.clientsMu.RUnlock()

	for _, client := range g.clients {
		if client.userID != exclude.userID {
			g.sendMessage(client, msgType, payload)
		}
	}
}

func (g *Gateway) GetClient(userID string) *Client {
	g.clientsMu.RLock()
	defer g.clientsMu.RUnlock()
	return g.clients[userID]
}
