package config

//go:generate go-enum --marshal --names

// ENUM(auto, light, dark).
type Theme string
