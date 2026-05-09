package config

import (
	"strings"
)

// renderConfig serializes a parsed ssh_config into a canonical text form.
func renderConfig(cfg *Config) []byte {
	if cfg == nil || len(cfg.Blocks) == 0 {
		return nil
	}

	var builder strings.Builder
	for index, block := range cfg.Blocks {
		if index > 0 {
			builder.WriteString(newLineStr)
		}
		renderBlock(&builder, block)
	}

	return []byte(builder.String())
}

// renderBlock writes one Host or Match block into the builder.
func renderBlock(builder *strings.Builder, block *Block) {
	builder.WriteString(renderBlockHeader(block))
	builder.WriteString(newLineStr)

	for _, directive := range block.Directives {
		builder.WriteString(renderDirective(directive))
		builder.WriteString(newLineStr)
	}
}

// renderBlockHeader formats the leading Host or Match line for a block.
func renderBlockHeader(block *Block) string {
	if block.Type == BlockMatch {
		parts := make([]string, 0, len(block.Conditions)*2)
		for _, condition := range block.Conditions {
			parts = append(parts, condition.Keyword)
			if condition.Value != emptyString {
				parts = append(parts, renderValue(condition.Value))
			}
		}
		return matchHeader + strings.Join(parts, blankStr)
	}

	patterns := make([]string, 0, len(block.Patterns))
	for _, pattern := range block.Patterns {
		patterns = append(patterns, renderValue(pattern))
	}
	return hostHeader + strings.Join(patterns, blankStr)
}

// renderDirective formats a directive line using a stable two-space indent.
func renderDirective(directive *Directive) string {
	return "  " + directive.Key + blankStr + joinRenderedValues(directive.Values)
}

// joinRenderedValues renders and joins directive values for one line.
func joinRenderedValues(values []string) string {
	rendered := make([]string, 0, len(values))
	for _, value := range values {
		rendered = append(rendered, renderValue(value))
	}
	return strings.Join(rendered, blankStr)
}

// renderValue quotes values that contain whitespace or quotes.
func renderValue(value string) string {
	if value == emptyString {
		return doubleQuoteStr + doubleQuoteStr
	}
	if !strings.ContainsAny(value, quotableChars) {
		return value
	}
	replacer := strings.NewReplacer(backslashStr, escapedBackslashStr, doubleQuoteStr, escapedQuoteStr)
	return doubleQuoteStr + replacer.Replace(value) + doubleQuoteStr
}
