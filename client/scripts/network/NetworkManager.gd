extends Node

signal connected()
signal disconnected()
signal authenticated()
signal connection_established()
signal character_list_received(characters: Array)
signal character_created(character: Dictionary)
signal move_updated(player_id: String, position: Vector2)
signal state_updated(state: Dictionary)
signal enter_world_response(player_data: Dictionary)

var ws_peer: WebSocketPeer = WebSocketPeer.new()
var server_url: String = "ws://localhost:8080/ws"
var auth_token: String = ""
var player_id: String = ""
var is_connected_to_server: bool = false
var is_authenticated: bool = false

var _pending_requests: Dictionary = {}
var _request_id: int = 0

func _exit_tree() -> void:
	ws_peer.close()

func _process(delta: float) -> void:
	ws_peer.poll()
	
	var state = ws_peer.get_ready_state()
	
	if state == WebSocketPeer.STATE_CONNECTING:
		return
	
	if state == WebSocketPeer.STATE_OPEN:
		if not is_connected_to_server:
			is_connected_to_server = true
			connection_established.emit()
		
		while ws_peer.get_available_packet_count() > 0:
			var packet: PackedByteArray = ws_peer.get_packet()
			_parse_packet(packet)
	
	if state == WebSocketPeer.STATE_CLOSED:
		if is_connected_to_server:
			is_connected_to_server = false
			disconnected.emit()

func connect_to_server(url: String = "") -> void:
	if url != "":
		server_url = url
	
	var err = ws_peer.connect_to_url(server_url)
	if err != OK:
		push_error("Failed to connect to server: %s" % error_string(err))

func authenticate(username: String, password: String) -> bool:
	if not is_connected_to_server:
		push_error("Not connected to server")
		return false
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id,
		"username": username,
		"password": password
	}
	
	_pending_requests[request_id] = "auth"
	_send_packet("auth_login", payload)
	return true

func register(username: String, email: String, password: String) -> bool:
	if not is_connected_to_server:
		push_error("Not connected to server")
		return false
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id,
		"username": username,
		"email": email,
		"password": password
	}
	
	_pending_requests[request_id] = "register"
	_send_packet("auth_register", payload)
	return true

func create_character(name: String, class_type: String) -> bool:
	if not is_authenticated:
		push_error("Not authenticated")
		return false
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id,
		"name": name,
		"class_type": class_type
	}
	
	_pending_requests[request_id] = "create_character"
	_send_packet("character_create", payload)
	return true

func get_characters() -> bool:
	if not is_authenticated:
		push_error("Not authenticated")
		return false
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id
	}
	
	_pending_requests[request_id] = "get_characters"
	_send_packet("character_list", payload)
	return true

func send_move(target_x: float, target_y: float, target_z: float) -> void:
	if not is_authenticated:
		return
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id,
		"timestamp": Time.get_unix_time_from_system() * 1000,
		"target_x": target_x,
		"target_y": target_y,
		"target_z": target_z
	}
	
	_send_packet("move_command", payload)

func send_move_2d(target_x: float, target_y: float) -> void:
	if not is_authenticated:
		return
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id,
		"timestamp": Time.get_unix_time_from_system() * 1000,
		"target_x": target_x,
		"target_y": target_y
	}
	
	_send_packet("move_command_2d", payload)

func send_move_2d_delta(delta_x: float, delta_y: float) -> void:
	if not is_authenticated:
		return
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id,
		"timestamp": Time.get_unix_time_from_system() * 1000,
		"delta_x": delta_x,
		"delta_y": delta_y
	}
	
	_send_packet("move_command_2d_delta", payload)

func enter_world(character_id: String) -> bool:
	if not is_authenticated:
		push_error("Not authenticated")
		return false
	
	var request_id = _generate_request_id()
	var payload = {
		"request_id": request_id,
		"character_id": character_id
	}
	
	_pending_requests[request_id] = "enter_world"
	_send_packet("enter_world", payload)
	return true

func send_chat_message(message: String) -> void:
	if not is_authenticated:
		return
	
	var payload = {
		"message": message,
		"timestamp": Time.get_unix_time_from_system() * 1000
	}
	
	_send_packet("chat_message", payload)

func _generate_request_id() -> int:
	_request_id += 1
	return _request_id

func _send_packet(packet_type: String, payload: Dictionary) -> void:
	var packet = {
		"type": packet_type,
		"payload": payload
	}
	
	var json = JSON.stringify(packet)
	ws_peer.send_text(json)

func _parse_packet(data: PackedByteArray) -> void:
	var json_string = data.get_string_from_utf8()
	var json = JSON.new()
	var err = json.parse(json_string)
	
	if err != OK:
		push_error("Failed to parse JSON: %s" % json_string)
		return
	
	var packet = json.get_data()
	var packet_type = packet.get("type", "")
	var payload = packet.get("payload", {})
	
	match packet_type:
		"auth_response":
			_handle_auth_response(payload)
		"character_list_response":
			_handle_character_list(payload)
		"character_create_response":
			_handle_character_create(payload)
		"move_response":
			_handle_move_response(payload)
		"state_update":
			state_updated.emit(payload)
		"enter_world_response":
			_handle_enter_world_response(payload)
		"error":
			_handle_error(payload)

func _handle_auth_response(payload: Dictionary) -> void:
	# Handle both string and int request_id (JSON number can be float64 in Go)
	var request_id_raw = payload.get("request_id")
	var request_id: int = 0
	
	if request_id_raw is int:
		request_id = request_id_raw
	elif request_id_raw is float:
		request_id = int(request_id_raw)
	elif request_id_raw is String:
		request_id = int(request_id_raw)
	
	var success = payload.get("success", false)
	
	if success:
		auth_token = payload.get("token", "")
		player_id = payload.get("player_id", "")
		is_authenticated = true
		authenticated.emit()
	else:
		push_error("Auth failed: %s" % payload.get("message", "Unknown error"))
	
	# Clean up pending request if it exists
	if request_id in _pending_requests:
		_pending_requests.erase(request_id)

func _handle_character_list(payload: Dictionary) -> void:
	var request_id = payload.get("request_id", 0)
	var success = payload.get("success", false)
	
	if success:
		var characters = payload.get("characters", [])
		character_list_received.emit(characters)
	else:
		push_error("Failed to get characters: %s" % payload.get("message", "Unknown error"))
	
	_pending_requests.erase(request_id)

func _handle_character_create(payload: Dictionary) -> void:
	var request_id = payload.get("request_id", 0)
	var success = payload.get("success", false)
	
	if success:
		var character = payload.get("character", {})
		character_created.emit(character)
	else:
		push_error("Failed to create character: %s" % payload.get("message", "Unknown error"))
	
	_pending_requests.erase(request_id)

func _handle_move_response(payload: Dictionary) -> void:
	var request_id = payload.get("request_id", 0)
	var success = payload.get("success", false)
	
	if success:
		var pos_x = payload.get("server_x", 0.0)
		var pos_y = payload.get("server_y", 0.0)
		var pos_z = payload.get("server_z", 0.0)
	else:
		push_error("Move failed: %s" % payload.get("message", "Unknown error"))
	
	_pending_requests.erase(request_id)

func _handle_enter_world_response(payload: Dictionary) -> void:
	var request_id = payload.get("request_id", 0)
	var success = payload.get("success", false)
	
	if success:
		player_id = payload.get("player_id", "")
		var server_x = payload.get("server_x", 0.0)
		var server_y = payload.get("server_y", 0.0)
		var server_z = payload.get("server_z", 0.0)
		enter_world_response.emit(payload)
		print("Successfully entered world at position: ", Vector2(server_x, server_y))
	else:
		push_error("Failed to enter world: %s" % payload.get("message", "Unknown error"))
	
	if request_id in _pending_requests:
		_pending_requests.erase(request_id)

func _handle_error(payload: Dictionary) -> void:
	var message = payload.get("message", "Unknown error")
	push_error("Server error: %s" % message)
	
	# For authentication errors, we need to ensure is_authenticated stays false
	# and the timeout in AuthMenu will trigger, showing the error to the user
	# Clean up any pending auth requests
	var keys_to_remove = []
	for key in _pending_requests:
		if _pending_requests[key] == "auth" or _pending_requests[key] == "register" or _pending_requests[key] == "enter_world":
			keys_to_remove.append(key)
	
	for key in keys_to_remove:
		_pending_requests.erase(key)
