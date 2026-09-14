package i18n

import (
	"net/http"
	"strings"
)

const (
	LangEN = "en"
	LangTR = "tr"
)

func Normalize(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if i := strings.IndexAny(lang, "-_"); i > 0 {
		lang = lang[:i]
	}
	switch lang {
	case LangTR:
		return LangTR
	default:
		return LangEN
	}
}

func Resolve(r *http.Request, tenantDefault string) string {
	if r != nil {
		if q := strings.TrimSpace(r.URL.Query().Get("lang")); q != "" {
			return Normalize(q)
		}
		if al := firstAcceptLanguage(r.Header.Get("Accept-Language")); al != "" {
			return Normalize(al)
		}
	}
	if strings.TrimSpace(tenantDefault) != "" {
		return Normalize(tenantDefault)
	}
	return LangEN
}

func firstAcceptLanguage(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	part := strings.Split(header, ",")[0]
	part = strings.TrimSpace(strings.Split(part, ";")[0])
	return part
}

func T(lang, key string) string {
	lang = Normalize(lang)
	if m, ok := catalogs[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if lang != LangEN {
		if v, ok := catalogs[LangEN][key]; ok {
			return v
		}
	}
	return key
}

func Dict(lang string) map[string]string {
	lang = Normalize(lang)
	src := catalogs[lang]
	if src == nil {
		src = catalogs[LangEN]
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	if lang != LangEN {
		for k, v := range catalogs[LangEN] {
			if _, ok := out[k]; !ok {
				out[k] = v
			}
		}
	}
	return out
}

func ClientDict(lang string) map[string]string {
	lang = Normalize(lang)
	out := make(map[string]string)
	for k, v := range catalogs[LangEN] {
		if strings.HasPrefix(k, "js.") || strings.HasPrefix(k, "error.") {
			out[k] = v
		}
	}
	if lang != LangEN {
		for k, v := range catalogs[lang] {
			if strings.HasPrefix(k, "js.") || strings.HasPrefix(k, "error.") {
				out[k] = v
			}
		}
	}
	return out
}
