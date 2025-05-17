package api

import (
	"context"
	"net/http"
	"strings"
)

type langContextKey string

const (
	LangContextKey langContextKey = "lang"
	DefaultLang    string         = "en"
)

// Language middleware to extract language preference from request
func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get language from query parameter
		lang := r.URL.Query().Get("lang")

		// If not in query, try to get from Accept-Language header
		if lang == "" {
			acceptLang := r.Header.Get("Accept-Language")
			if strings.Contains(strings.ToLower(acceptLang), "ru") {
				lang = "ru"
			} else {
				lang = DefaultLang
			}
		}

		// Validate language
		if lang != "ru" && lang != "en" {
			lang = DefaultLang
		}

		// Add language to context
		ctx := context.WithValue(r.Context(), LangContextKey, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetLanguage gets the language from context
func GetLanguage(ctx context.Context) string {
	if lang, ok := ctx.Value(LangContextKey).(string); ok {
		return lang
	}
	return DefaultLang
}

// GetLocalizedContent returns content in the requested language
func GetLocalizedContent(content map[string]string, lang string) string {
	if content == nil {
		return ""
	}

	if val, ok := content[lang]; ok && val != "" {
		return val
	}

	// Fallback to default language if requested language is not available
	if val, ok := content[DefaultLang]; ok {
		return val
	}

	// If no content is available in any language
	return ""
}
