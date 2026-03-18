extends Node3D
class_name PlayerController

@export var network: NetworkManager
@export var move_speed: float = 5.0

var target_position: Vector3 = Vector3.ZERO
var is_moving: bool = false
var current_animation: String = "idle"

@onready var player_mesh: MeshInstance3D = $MeshInstance3D

func _ready() -> void:
	if player_mesh:
		player_mesh.visible = true

func _input(event: InputEvent) -> void:
	if event is InputEventMouseButton:
		if event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
			var space_state = get_world_3d().direct_space_state
			var query = PhysicsRayQueryParameters3D.create(
				get_viewport().get_camera_3d().global_position,
				event.global_position
			)
			
			var result = space_state.intersect_ray(query)
			if result.has("position"):
				target_position = result.position
				target_position.y = 0
				is_moving = true
				send_move_command()

func _physics_process(delta: float) -> void:
	if is_moving:
		var direction = (target_position - global_position).normalized()
		direction.y = 0
		
		if direction.length() > 0.1:
			global_position += direction * move_speed * delta
			_update_animation("run")
			
			if global_position.distance_to(target_position) < 0.1:
				is_moving = false
				_update_animation("idle")
		else:
			is_moving = false
			_update_animation("idle")

func send_move_command() -> void:
	if network and network.is_authenticated:
		network.send_move(target_position.x, target_position.y, target_position.z)

func _update_animation(anim_name: String) -> void:
	if anim_name != current_animation:
		current_animation = anim_name
		# Animation logic would go here with AnimationTree

func set_player_visibility(visible: bool) -> void:
	if player_mesh:
		player_mesh.visible = visible
