extends Control
class_name CharacterCreation

@onready var name_input: LineEdit = $MainContainer/NameInput
@onready var warrior_button: Button = $MainContainer/ClassGrid/WarriorButton
@onready var thief_button: Button = $MainContainer/ClassGrid/ThiefButton
@onready var mage_button: Button = $MainContainer/ClassGrid/MageButton
@onready var class_description: Label = $MainContainer/ClassDescription
@onready var create_button: Button = $MainContainer/CreateButton
@onready var cancel_button: Button = $MainContainer/CancelButton
@onready var status_label: Label = $MainContainer/StatusLabel

var network_manager: NetworkManager
var selected_class: String = ""
var class_descriptions: Dictionary = {
	"warrior": "Maître du combat au corps à corps. Fortes statistiques de vie et de dégâts.",
	"rogue": "Expert en furtivité et en coups critiques. Rapide et létal.",
	"mage": "Maître des arcanes. Dégâts à distance et sorts puissants."
}

func _ready() -> void:
	network_manager = get_node_or_null("/root/NetworkManager")
	if not network_manager:
		push_error("NetworkManager not found in scene tree")
		return
	
	warrior_button.pressed.connect(_on_warrior_pressed)
	thief_button.pressed.connect(_on_thief_pressed)
	mage_button.pressed.connect(_on_mage_pressed)
	create_button.pressed.connect(_on_create_pressed)
	cancel_button.pressed.connect(_on_cancel_pressed)
	
	name_input.text_changed.connect(_on_name_changed)

func _on_warrior_pressed() -> void:
	_select_class("warrior")
	warrior_button.add_theme_color_override("font_color", Color(1.0, 0.4, 0.4, 1))
	thief_button.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))
	mage_button.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))

func _on_thief_pressed() -> void:
	_select_class("rogue")
	warrior_button.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))
	thief_button.add_theme_color_override("font_color", Color(1.0, 0.8, 0.2, 1))
	mage_button.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))

func _on_mage_pressed() -> void:
	_select_class("mage")
	warrior_button.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))
	thief_button.add_theme_color_override("font_color", Color(0.7, 0.7, 0.7, 1))
	mage_button.add_theme_color_override("font_color", Color(0.4, 0.6, 1.0, 1))

func _select_class(classname: String) -> void:
	selected_class = classname
	class_description.text = class_descriptions.get(classname, "")
	_validate_form()

func _on_name_changed(newtext: String) -> void:
	_validate_form()

func _validate_form() -> void:
	var name_valid = name_input.text.strip_edges().length() >= 3
	var class_valid = not selected_class.is_empty()
	create_button.disabled = not (name_valid and class_valid)

func _on_create_pressed() -> void:
	var character_name: String = name_input.text.strip_edges()
	
	if character_name.length() < 3:
		status_label.text = "Le nom doit contenir au moins 3 caractères!"
		status_label.modulate = Color(1, 0.3, 0.3, 1)
		return
	
	if selected_class.is_empty():
		status_label.text = "Veuillez choisir une classe!"
		status_label.modulate = Color(1, 0.3, 0.3, 1)
		return
	
	status_label.text = "Création en cours..."
	status_label.modulate = Color(0.9, 0.7, 0.3, 1)
	create_button.disabled = true
	
	network_manager.create_character(character_name, selected_class)
	await network_manager.character_created
	
	status_label.text = "Personnage créé avec succès!"
	status_label.modulate = Color(0.3, 1.0, 0.3, 1)
	
	await get_tree().create_timer(1.5).timeout
	get_tree().change_scene_to_file("res://scenes/ui/menus/CharacterSelection.tscn")

func _on_cancel_pressed() -> void:
	get_tree().change_scene_to_file("res://scenes/ui/menus/CharacterSelection.tscn")
