package config

//go:generate go-enum --marshal --names

// ENUM(auto, native, osc52).
type ClipboardBackend string
