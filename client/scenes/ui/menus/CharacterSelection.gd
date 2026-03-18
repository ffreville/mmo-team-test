extends Control
class_name CharacterSelection

@onready var title_label: Label = $VBoxContainer/TitleLabel
@onready var characters_container: VBoxContainer = $VBoxContainer/CharactersContainer
@onready var create_button: Button = $VBoxContainer/CreateButton
@onready var play_button: Button = $VBoxContainer/PlayButton
@onready var back_button: Button = $VBoxContainer/BackButton
@onready var status_label: Label = $VBoxContainer/StatusLabel

var network_manager: NetworkManager
var characters: Array = []
var selected_character_index: int = -1

func _ready() -> void:
	network_manager = get_node_or_null("/root/NetworkManager")
	if not network_manager:
		push_error("NetworkManager not found in scene tree")
		return
	
	create_button.pressed.connect(_on_create_pressed)
	play_button.pressed.connect(_on_play_pressed)
	back_button.pressed.connect(_on_back_pressed)
	network_manager.character_list_received.connect(_on_character_list_received)
	
	if network_manager.is_authenticated:
		network_manager.get_characters()
		await network_manager.character_list_received
		_update_character_list()

func _on_character_list_received(characters_data: Array) -> void:
	characters = characters_data
	print("CharacterSelection: Received ", characters.size(), " characters")
	_update_character_list()

func _update_character_list() -> void:
	# Clear existing character buttons
	for child in characters_container.get_children():
		characters_container.remove_child(child)
		child.queue_free()
	
	if characters.is_empty():
		title_label.text = "No Characters Found"
		play_button.disabled = true
	else:
		title_label.text = "Select Your Character (%d)" % characters.size()
		
		# Create a button for each character
		for i in range(characters.size()):
			var char_data: Dictionary = characters[i]
			var button = Button.new()
			button.text = "%s (Level %d %s)" % [
				char_data.get("name", "Unknown"),
				char_data.get("level", 1),
				char_data.get("class_type", "").capitalize()
			]
			button.custom_minimum_size = Vector2(250, 50)
			button.pressed.connect(_on_character_selected.bind(i))
			characters_container.add_child(button)
		
		if selected_character_index >= 0 and selected_character_index < characters.size():
			play_button.disabled = false
		else:
			play_button.disabled = true

func _on_character_selected(index: int) -> void:
	selected_character_index = index
	play_button.disabled = false
	print("Character selected: ", characters[index].get("name"))

func _on_create_pressed() -> void:
	get_tree().change_scene_to_file("res://scenes/ui/menus/CharacterCreation.tscn")

func _on_play_pressed() -> void:
	if selected_character_index == -1 or selected_character_index >= characters.size():
		status_label.text = "No character selected!"
		status_label.modulate = Color(1, 0.3, 0.3, 1)
		return
	
	var char_data: Dictionary = characters[selected_character_index]
	status_label.text = "Loading world..."
	status_label.modulate = Color(0.9, 0.7, 0.3, 1)
	
	get_tree().set_meta("current_character", char_data)
	get_tree().change_scene_to_file("res://scenes/main/Main.tscn")

func _on_back_pressed() -> void:
	get_tree().change_scene_to_file("res://scenes/ui/menus/AuthMenu.tscn")
