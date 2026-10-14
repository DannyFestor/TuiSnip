package snippetslab

import "github.com/DannyFestor/TuiSnip/internal/domain/value"

func languageNamesByLexer() map[string]string {
	return map[string]string{
		"BashLexer":          "Bash",
		"CLexer":             "C",
		"CSharpLexer":        "C#",
		"CppLexer":           "C++",
		"CssLexer":           "CSS",
		"DartLexer":          "Dart",
		"DockerLexer":        "Docker",
		"GoLexer":            "Go",
		"HtmlLexer":          "HTML",
		"HtmlPhpLexer":       "PHTML",
		"JavaLexer":          "Java",
		"JavascriptLexer":    "JavaScript",
		"JavascriptPhpLexer": "JavaScript",
		"JsonLexer":          "JSON",
		"KotlinLexer":        "Kotlin",
		"MarkdownLexer":      "markdown",
		"MySqlLexer":         "MySQL",
		"OcamlLexer":         "OCaml",
		"OdinLexer":          "Odin",
		"PhpLexer":           "PHP",
		"PostgresLexer":      "PostgreSQL SQL dialect",
		"RustLexer":          "Rust",
		"SqlLexer":           "SQL",
		"TOMLLexer":          "TOML",
		"TextLexer":          "plaintext",
		"TypeScriptLexer":    "TypeScript",
		"YamlLexer":          "YAML",
		"ZigLexer":           "Zig",
	}
}

func languageOfLexer(lexer string) (value.Language, bool) {
	language, err := value.NewLanguage(languageNamesByLexer()[lexer])
	if err != nil {
		return value.PlainText(), false
	}

	return language, true
}
