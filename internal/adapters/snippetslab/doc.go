// Package snippetslab reads a SnippetsLab JSON export into a domain.Library with fresh
// IDs, so that SnippetsLab's format and its mapping onto TuiSnip stay in one place and
// the Import Action only ever sees a domain.Library and what the Import leaves behind.
package snippetslab
