// Package system provides the real Clock and IDGenerator, so that the core never
// reads the clock or generates IDs itself and stays deterministic in tests.
package system
