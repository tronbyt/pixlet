package i18n_runtime

import (
	"go.starlark.net/starlark"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

const languageThreadKey = "github.com/tronbyt/pixlet/runtime/$language"

var DefaultLanguage = language.English

func AttachLanguageToThread(t *starlark.Thread, lang language.Tag) {
	t.SetLocal(languageThreadKey, lang)
}

func LanguageFromThread(thread *starlark.Thread) language.Tag {
	if thread != nil {
		if lang, ok := thread.Local(languageThreadKey).(language.Tag); ok {
			return lang
		}
	}
	return DefaultLanguage
}

const catalogThreadKey = "github.com/tronbyt/pixlet/runtime/$catalog"

func AttachCatalogToThread(t *starlark.Thread, b *catalog.Builder) {
	t.SetLocal(catalogThreadKey, b)
}

func CatalogFromThread(thread *starlark.Thread) *catalog.Builder {
	if thread != nil {
		if b, ok := thread.Local(catalogThreadKey).(*catalog.Builder); ok {
			return b
		}
	}
	return catalog.NewBuilder()
}
