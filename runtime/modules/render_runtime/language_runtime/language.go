package language_runtime

import (
	"go.starlark.net/starlark"
	"golang.org/x/text/language"
)

const threadKey = "github.com/tronbyt/pixlet/runtime/$language"

var DefaultLanguage = language.English

func AttachToThread(t *starlark.Thread, lang language.Tag) {
	t.SetLocal(threadKey, lang)
}

func FromThread(thread *starlark.Thread) language.Tag {
	if thread != nil {
		if lang, ok := thread.Local(threadKey).(language.Tag); ok {
			return lang
		}
	}
	return DefaultLanguage
}
