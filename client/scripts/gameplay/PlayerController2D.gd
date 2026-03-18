extends CharacterBody2D
class_name PlayerController2D

@export var network: NetworkManager
@export var move_speed: float = 200.0

var current_animation: String = "idle"

# Keyboard movement variables
var keyboard_velocity: Vector2 = Vector2.ZERO
var last_move_time: float = 0.0
var move_command_interval: float = 0.1  # Send move command every 100ms
var last_sent_position: Vector2 = Vector2.ZERO  # Track last sent position for delta calculation

@onready var sprite: Sprite2D = $Sprite2D
@onready var collision: CollisionShape2D = $CollisionShape2D

func _physics_process(delta: float) -> void:
	# Handle keyboard movement only
	_handle_keyboard_movement(delta)

func _handle_keyboard_movement(delta: float) -> void:
	# Calculate velocity from keyboard input
	keyboard_velocity = Vector2.ZERO
	
	if Input.is_key_pressed(KEY_W) or Input.is_key_pressed(KEY_UP):
		keyboard_velocity.y -= 1
	if Input.is_key_pressed(KEY_S) or Input.is_key_pressed(KEY_DOWN):
		keyboard_velocity.y += 1
	if Input.is_key_pressed(KEY_A) or Input.is_key_pressed(KEY_LEFT):
		keyboard_velocity.x -= 1
	if Input.is_key_pressed(KEY_D) or Input.is_key_pressed(KEY_RIGHT):
		keyboard_velocity.x += 1
	
	# Normalize to maintain consistent speed in diagonal directions
	if keyboard_velocity.length() > 0:
		keyboard_velocity = keyboard_velocity.normalized()
	
	# Apply movement
	if keyboard_velocity.length() > 0:
		global_position += keyboard_velocity * move_speed * delta
		_update_animation("run")
		
		# Face direction (only horizontal flipping for 2D top-down)
		if keyboard_velocity.x < 0:
			sprite.flip_h = true
		elif keyboard_velocity.x > 0:
			sprite.flip_h = false
		
		# Send move command periodically
		var current_time = Time.get_unix_time_from_system()
		if current_time - last_move_time >= move_command_interval:
			send_keyboard_move_command()
			last_move_time = current_time
	else:
		_update_animation("idle")

func _ready() -> void:
	if sprite:
		sprite.visible = true
		# Create a placeholder texture if none is set
		if sprite.texture == null:
			_create_placeholder_texture()
	if collision:
		collision.visible = false
	# Initialize last_sent_position to current position for delta movement tracking
	last_sent_position = global_position

func send_keyboard_move_command() -> void:
	if network and network.is_authenticated:
		# Calculate delta from last sent position
		var delta_x = global_position.x - last_sent_position.x
		var delta_y = global_position.y - last_sent_position.y
		
		# Debug: Log delta values being sent
		print("[DEBUG] Sending delta movement - delta_x: %.4f, delta_y: %.4f" % [delta_x, delta_y])
		var delta_length = Vector2(delta_x, delta_y).length()
		print("[DEBUG] Delta length: %.4f" % delta_length)
		print("[DEBUG] Current position: (%.2f, %.2f), Last sent: (%.2f, %.2f)" % [global_position.x, global_position.y, last_sent_position.x, last_sent_position.y])
		
		# Send delta movement instead of absolute position
		network.send_move_2d_delta(delta_x, delta_y)
		
		# Update last sent position
		last_sent_position = global_position



func _update_animation(anim_name: String) -> void:
	if anim_name != current_animation:
		current_animation = anim_name
		# Animation logic would go here with AnimationPlayer

func set_player_visibility(visible: bool) -> void:
	if sprite:
		sprite.visible = visible

func _create_placeholder_texture() -> void:
	# Create a simple 32x32 blue square texture
	var image = Image.create(32, 32, false, Image.FORMAT_RGBA8)
	image.fill(Color(0.2, 0.6, 1.0, 1.0))  # Blue color
	
	var texture = ImageTexture.create_from_image(image)
	sprite.texture = texture
	sprite.modulate = Color(1, 1, 1, 1)  # Reset modulate since texture has color
	print("Created placeholder texture for player")
