extends Node2D
class_name Main

@onready var status_label: Label = $HUD/StatusLabel
@onready var player_info_label: Label = $HUD/PlayerInfoLabel
@onready var position_label: Label = $HUD/PositionLabel

var network_manager: NetworkManager
var player: CharacterBody2D
var _world_ready: bool = false

# Décorations
var _decorations_layer: Node2D
var _tree_texture: Texture2D
var _rock_texture: Texture2D
var _bush_texture: Texture2D

func _ready() -> void:
	network_manager = get_node_or_null("/root/NetworkManager")
	if not network_manager:
		push_error("NetworkManager not found in scene tree")
		return
	
	if not network_manager.is_authenticated:
		push_warning("Not authenticated, redirecting to auth menu")
		get_tree().change_scene_to_file.call_deferred("res://scenes/ui/menus/AuthMenu.tscn")
		return
	
	# Connect to enter_world_response signal
	network_manager.enter_world_response.connect(_on_enter_world_response)
	
	# Setup decorations before entering world
	_setup_decorations()
	
	_setup_player_info()
	_enter_world()
	status_label.text = "Connecté - %s" % Time.get_datetime_string_from_system(false, true)

func _setup_player_info() -> void:
	var current_character = _get_current_character()
	if current_character.is_empty():
		player_info_label.text = "Joueur: Anonymous"
	else:
		var char_name = current_character.get("name", "Unknown")
		var char_class = _format_class_name(current_character.get("class_type", ""))
		player_info_label.text = "Joueur: %s (%s)" % [char_name, char_class]

func _get_current_character() -> Dictionary:
	var user_data = get_tree().get_meta("current_character", {})
	return user_data if user_data else {}

func _format_class_name(class_type: String) -> String:
	match class_type.to_lower():
		"guerrier": return "Guerrier"
		"voleur": return "Voleur"
		"mage": return "Mage"
		_: return class_type.capitalize()

func _enter_world() -> void:
	# Get current character data
	var current_character = _get_current_character()
	if current_character.is_empty():
		push_error("No character data found, redirecting to character selection")
		get_tree().change_scene_to_file.call_deferred("res://scenes/ui/menus/CharacterSelection.tscn")
		return
	
	var character_id = current_character.get("character_id", "")
	if character_id.is_empty():
		push_error("Character ID is empty")
		return
	
	print("Entering world with character: ", current_character.get("name"))
	
	# Tell server to create player entry
	var success = network_manager.enter_world(character_id)
	if not success:
		push_error("Failed to enter world")
		return
	
	# Wait for server response before spawning player
	# Player will be spawned in _on_enter_world_response()

func _on_enter_world_response(player_data: Dictionary) -> void:
	print("Enter world response received: ", player_data)
	
	# Spawn player at server position
	_spawn_player_at_position(player_data.get("server_x", 0.0), 
	                          player_data.get("server_y", 0.0))
	
	# Setup camera to follow player
	_setup_camera()
	
	_world_ready = true
	status_label.text = "Monde chargé - Position: (%.1f, %.1f)" % [player_data.get("server_x", 0.0), player_data.get("server_y", 0.0)]



func _spawn_player_at_position(pos_x: float, pos_y: float) -> void:
	# Load player scene
	var player_scene = preload("res://scenes/character/Player2D.tscn")
	if not player_scene:
		push_error("Failed to load Player2D.tscn")
		return
	
	# Instantiate player
	player = player_scene.instantiate()
	player.name = "LocalPlayer"
	
	# Set initial position from server
	player.global_position = Vector2(pos_x, pos_y)
	
	# Set network reference if player has it
	if player.has_method("set_network_manager"):
		player.set_network_manager(network_manager)
	elif "network" in player:
		player.network = network_manager
	
	add_child(player)
	print("Player spawned at server position: ", player.global_position)

func _spawn_player() -> void:
	# Deprecated: use _spawn_player_at_position instead
	_spawn_player_at_position(0.0, 0.0)

func _setup_camera() -> void:
	# Create camera and attach to player
	var camera = Camera2D.new()
	camera.name = "MainCamera"
	camera.position = Vector2.ZERO  # Camera at player's center
	
	if player:
		player.add_child(camera)
		camera.set_as_top_level(false)  # Camera moves with player
	
	print("Camera attached to player")

func _process(delta: float) -> void:
	# Update position label if player exists
	if player and position_label:
		position_label.text = "Position: (%.0f, %.0f)" % [player.global_position.x, player.global_position.y]

# ============================================================================
# DÉCORATIONS
# ============================================================================

func _setup_decorations() -> void:
	"""Crée la couche de décorations et génère les textures."""
	# Créer le node DecorationsLayer
	_decorations_layer = Node2D.new()
	_decorations_layer.name = "DecorationsLayer"
	add_child(_decorations_layer)
	
	# Générer les textures
	_tree_texture = _create_tree_texture()
	_rock_texture = _create_rock_texture()
	_bush_texture = _create_bush_texture()
	
	# Ajouter les décorations
	_add_trees()
	_add_rocks()
	_add_bushes()
	
	print("Décorations chargées: 7 arbres, 5 rochers, 3 buissons")

func _create_tree_texture() -> Texture2D:
	"""Crée une texture d'arbre (feuillage vert + tronc brun)."""
	var img = Image.create(64, 64, false, Image.FORMAT_RGBA8)
	img.fill(Color(0, 0, 0, 0))  # Transparent
	
	# Tronc brun (rectangle en bas)
	var trunk_rect = Rect2(24, 40, 16, 20)
	img.fill_rect(trunk_rect, Color(0.5, 0.3, 0.1, 1))
	
	# Feuillage vert (cercle au-dessus)
	var foliage_center = Vector2(32, 28)
	var foliage_radius = 20
	for y in range(64):
		for x in range(64):
			var dist = Vector2(x, y).distance_to(foliage_center)
			if dist < foliage_radius:
				# Légère variation de vert
				var shade = 0.8 + (randf() * 0.2)
				img.set_pixel(x, y, Color(0.2, shade, 0.2, 1))
	
	var tex = ImageTexture.create_from_image(img)
	return tex

func _create_rock_texture() -> Texture2D:
	"""Crée une texture de rocher (forme irrégulière grise)."""
	var img = Image.create(48, 48, false, Image.FORMAT_RGBA8)
	img.fill(Color(0, 0, 0, 0))  # Transparent
	
	# Forme irrégulière de rocher
	var rock_points = [
		Vector2(10, 24), Vector2(18, 12), Vector2(30, 10),
		Vector2(40, 18), Vector2(42, 30), Vector2(35, 40),
		Vector2(20, 42), Vector2(8, 35)
	]
	
	# Dessiner le polygone
	var color_base = Color(0.4, 0.4, 0.45, 1)
	for point in rock_points:
		var x = int(point.x)
		var y = int(point.y)
		# Dessiner un petit cercle pour chaque point
		for dy in range(-3, 4):
			for dx in range(-3, 4):
				if dx*dx + dy*dy <= 9:
					if x+dx >= 0 and x+dx < 48 and y+dy >= 0 and y+dy < 48:
						var shade = 0.9 + (randf() * 0.2)
						img.set_pixel(x+dx, y+dy, Color(
							color_base.r * shade,
							color_base.g * shade,
							color_base.b * shade,
							1
						))
	
	# Remplir l'intérieur
	for y in range(48):
		for x in range(48):
			if _is_point_in_polygon(Vector2(x, y), rock_points):
				if img.get_pixel(x, y).a < 0.5:
					var shade = 0.85 + (randf() * 0.15)
					img.set_pixel(x, y, Color(
						color_base.r * shade,
						color_base.g * shade,
						color_base.b * shade,
						1
					))
	
	var tex = ImageTexture.create_from_image(img)
	return tex

func _is_point_in_polygon(point: Vector2, polygon: Array) -> bool:
	"""Vérifie si un point est à l'intérieur d'un polygone (algorithme du rayon)."""
	var inside = false
	var j = polygon.size() - 1
	
	for i in range(polygon.size()):
		var pi = polygon[i]
		var pj = polygon[j]
		
		if ((pi.y > point.y) != (pj.y > point.y)) and \
		   (point.x < (pj.x - pi.x) * (point.y - pi.y) / (pj.y - pi.y) + pi.x):
			inside = !inside
		
		j = i
	
	return inside

func _create_bush_texture() -> Texture2D:
	"""Crée une texture de buisson (cercle vert foncé)."""
	var img = Image.create(40, 40, false, Image.FORMAT_RGBA8)
	img.fill(Color(0, 0, 0, 0))  # Transparent
	
	var center = Vector2(20, 20)
	var radius = 16
	
	for y in range(40):
		for x in range(40):
			var dist = Vector2(x, y).distance_to(center)
			if dist < radius:
				# Vert foncé avec variations
				var shade = 0.3 + (randf() * 0.2)
				img.set_pixel(x, y, Color(0.1, shade, 0.1, 1))
	
	var tex = ImageTexture.create_from_image(img)
	return tex

func _add_trees() -> void:
	"""Ajoute 7 arbres à différentes positions."""
	var tree_positions = [
		Vector2(-600, -400),
		Vector2(-300, 200),
		Vector2(100, -500),
		Vector2(400, -200),
		Vector2(600, 300),
		Vector2(-500, 400),
		Vector2(200, 500)
	]
	
	for i in range(tree_positions.size()):
		var sprite = Sprite2D.new()
		sprite.texture = _tree_texture
		sprite.global_position = tree_positions[i]
		# Légère variation d'échelle
		var scale = 0.8 + (randf() * 0.4)
		sprite.scale = Vector2(scale, scale)
		# Rotation aléatoire légère
		sprite.rotation = randf_range(-0.1, 0.1)
		_decorations_layer.add_child(sprite)

func _add_rocks() -> void:
	"""Ajoute 5 rochers à différentes positions."""
	var rock_positions = [
		Vector2(-700, 100),
		Vector2(-100, -300),
		Vector2(300, 400),
		Vector2(500, -400),
		Vector2(-400, -100)
	]
	
	for i in range(rock_positions.size()):
		var sprite = Sprite2D.new()
		sprite.texture = _rock_texture
		sprite.global_position = rock_positions[i]
		# Variation d'échelle
		var scale = 0.7 + (randf() * 0.5)
		sprite.scale = Vector2(scale, scale)
		# Rotation aléatoire
		sprite.rotation = randf_range(-0.3, 0.3)
		_decorations_layer.add_child(sprite)

func _add_bushes() -> void:
	"""Ajoute 3 buissons à différentes positions."""
	var bush_positions = [
		Vector2(-200, -600),
		Vector2(0, 0),
		Vector2(700, 500)
	]
	
	for i in range(bush_positions.size()):
		var sprite = Sprite2D.new()
		sprite.texture = _bush_texture
		sprite.global_position = bush_positions[i]
		# Variation d'échelle
		var scale = 0.9 + (randf() * 0.3)
		sprite.scale = Vector2(scale, scale)
		_decorations_layer.add_child(sprite)
