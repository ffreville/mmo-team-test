extends Control
class_name AuthMenu

@onready var username_input: LineEdit = $VBoxContainer/UsernameInput
@onready var password_input: LineEdit = $VBoxContainer/PasswordInput
@onready var email_input: LineEdit = $VBoxContainer/EmailInput
@onready var auth_mode_label: Label = $VBoxContainer/AuthModeLabel
@onready var email_label: Label = $VBoxContainer/EmailLabel
@onready var action_button: Button = $VBoxContainer/ActionButtons/ActionButton
@onready var toggle_mode_button: Button = $VBoxContainer/ActionButtons/ToggleModeButton
@onready var status_label: Label = $VBoxContainer/StatusLabel

var network_manager: NetworkManager
var is_login_mode: bool = true

func _ready() -> void:
	network_manager = get_node_or_null("/root/NetworkManager")
	if not network_manager:
		push_error("NetworkManager not found in scene tree")
		return
	
	action_button.pressed.connect(_on_action_pressed)
	toggle_mode_button.pressed.connect(_on_toggle_mode_pressed)
	
	_update_ui()

func _update_ui() -> void:
	if is_login_mode:
		auth_mode_label.text = "Mode: Login"
		action_button.text = "Login"
		toggle_mode_button.text = "Go to Register"
		email_label.visible = false
		email_input.visible = false
	else:
		auth_mode_label.text = "Mode: Register"
		action_button.text = "Register"
		toggle_mode_button.text = "Go to Login"
		email_label.visible = true
		email_input.visible = true
	
	status_label.text = ""

func _on_action_pressed() -> void:
	var username: String = username_input.text.strip_edges()
	var password: String = password_input.text
	var email: String = email_input.text.strip_edges()
	
	if username.is_empty():
		status_label.text = "Username is required!"
		status_label.modulate = Color(1, 0.3, 0.3, 1)
		return
	
	if password.is_empty():
		status_label.text = "Password is required!"
		status_label.modulate = Color(1, 0.3, 0.3, 1)
		return
	
	if not is_login_mode and email.is_empty():
		status_label.text = "Email is required for registration!"
		status_label.modulate = Color(1, 0.3, 0.3, 1)
		return
	
	status_label.text = "Processing..."
	status_label.modulate = Color(0.9, 0.7, 0.3, 1)
	
	# Ensure connection with timeout
	if not network_manager.is_connected_to_server:
		network_manager.connect_to_server()
		# Wait for connection with a timeout
		var timeout_timer = get_tree().create_timer(5.0)
		while not network_manager.is_connected_to_server and timeout_timer.time_left > 0:
			await get_tree().process_frame
		
		if not network_manager.is_connected_to_server:
			status_label.text = "Connection timeout!"
			status_label.modulate = Color(1, 0.3, 0.3, 1)
			return
	
	if is_login_mode:
		network_manager.authenticate(username, password)
		# Wait for authentication with timeout
		var auth_timeout_timer = get_tree().create_timer(10.0)
		var auth_success = false
		while not network_manager.is_authenticated and auth_timeout_timer.time_left > 0:
			await get_tree().process_frame
			if network_manager.is_authenticated:
				auth_success = true
				break
		
		if auth_success:
			get_tree().change_scene_to_file("res://scenes/ui/menus/CharacterSelection.tscn")
		else:
			status_label.text = "Login failed or timeout!"
			status_label.modulate = Color(1, 0.3, 0.3, 1)
	else:
		network_manager.register(username, email, password)
		# Wait for authentication with timeout
		var reg_timeout_timer = get_tree().create_timer(10.0)
		var reg_success = false
		while not network_manager.is_authenticated and reg_timeout_timer.time_left > 0:
			await get_tree().process_frame
			if network_manager.is_authenticated:
				reg_success = true
				break
		
		if reg_success:
			get_tree().change_scene_to_file("res://scenes/ui/menus/CharacterSelection.tscn")
		else:
			status_label.text = "Registration failed or timeout!"
			status_label.modulate = Color(1, 0.3, 0.3, 1)

func _on_toggle_mode_pressed() -> void:
	is_login_mode = not is_login_mode
	username_input.clear()
	password_input.clear()
	email_input.clear()
	_update_ui()
