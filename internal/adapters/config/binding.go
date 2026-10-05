package config

//go:generate go-enum --marshal --names

// ENUM(quit, help, search, zoom, new_snippet, capture, focus_next, focus_prev, focus_right, focus_left, focus_folders, focus_tags, focus_list, focus_snippet, open, back, down, up, top, bottom, page_down, page_up, new_folder, new_tag, rename, delete, move, collapse, language, copy, edit, open_in_editor, duplicate, cycle_sort, wrap, save, cancel, next_field, prev_field, open_field, pick_language, edit_tags, leave, indent, dedent, accept, show_all_languages, yes, no).
type Binding string
