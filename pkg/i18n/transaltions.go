package i18n

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:generate gotext -srclang=en-GB update -out=catalog.go -lang=en-GB,ru-RU,es-ES,fr-FR,de-DE,ko-KR,tr-TR,it-IT,pl-PL,uk-UA,be-BY,nl-NL,ms-MY,pt-PT,ca-ES,fa-IR github.com/ksysoev/help-my-pet/pkg/bot

const DefaultLanguage = "en"

var supportedLanguages = map[string]language.Tag{
	"en": language.MustParse("en-GB"),
	"ru": language.MustParse("ru-RU"),
	"es": language.MustParse("es-ES"),
	"fr": language.MustParse("fr-FR"),
	"de": language.MustParse("de-DE"),
	"ko": language.MustParse("ko-KR"),
	"tr": language.MustParse("tr-TR"),
	"it": language.MustParse("it-IT"),
	"pl": language.MustParse("pl-PL"),
	"uk": language.MustParse("uk-UA"),
	"be": language.MustParse("be-BY"),
	"nl": language.MustParse("nl-NL"),
	"ms": language.MustParse("ms-MY"),
	"pt": language.MustParse("pt-PT"),
	"ca": language.MustParse("ca-ES"),
	"fa": language.MustParse("fa-IR"),
}

// Localizer manages message printers for multiple languages, enabling localization support and language-specific formatting.
type Localizer struct {
	printers map[string]*message.Printer
}

// NewLocalizer creates and initializes a Localizer with printers for all supported languages.
// It configures a printer for each language in the supportedLanguages map and returns a Localizer instance.
func NewLocalizer() *Localizer {
	printers := make(map[string]*message.Printer, len(supportedLanguages))

	for lang, tag := range supportedLanguages {
		printers[lang] = message.NewPrinter(tag)
	}

	return &Localizer{
		printers: printers,
	}
}

// GetPrinter retrieves the message printer for the specified language.
// It returns the printer instance for the given language if available or falls back to the default language printer.
func (l *Localizer) GetPrinter(lang string) *message.Printer {
	if p, ok := l.printers[lang]; ok {
		return p
	}

	return l.printers[DefaultLanguage]
}
