package web

import (
	"embed"
	"html/template"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/Masterminds/sprig/v3"
	"github.com/reddec/view"
)

//go:embed all:views
var views embed.FS

//go:embed all:static
var Static embed.FS

func View[Args any](path string) *view.View[Args] {
	funcs := sprig.HtmlFuncMap()
	funcs["isoDateTime"] = func(t time.Time) string {
		return t.Format(time.RFC3339)
	}
	funcs["prettyDateTime"] = func(t time.Time) string {
		return t.Format(time.DateTime)
	}
	funcs["seconds"] = func(t float64) time.Duration {
		return time.Duration(float64(time.Second) * t)
	}
	funcs["isValidString"] = func(data []byte) bool {
		return utf8.Valid(data)
	}
	funcs["urlpath"] = func(v string) string {
		return url.PathEscape(v)
	}
	return view.Must(view.NewTemplate[Args](template.New("").Funcs(funcs), views, path))
}
