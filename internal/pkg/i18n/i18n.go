// Package i18n holds the small, hand-translated message catalog used for
// request-validation field errors — the one piece of API-level text
// localized in this phase.
//
// This is deliberately narrow in scope: apperror's ~60 hand-written
// service-level messages ("invalid credentials", "could not create
// account", etc.) are NOT localized here. Doing that properly would mean
// apperror.Error carrying message *keys* instead of literal strings and
// touching nearly every service file — a real architectural change, not
// a catalog addition, and out of scope for this phase. See the README's
// "what's intentionally not here yet" section.
package i18n

// Key identifies one validator tag's message, independent of wording in
// any language.
type Key string

const (
	KeyRequired     Key = "required"
	KeyEmail        Key = "email"
	KeyTooShort     Key = "too_short"
	KeyTooLong      Key = "too_long"
	KeyInvalidPhone Key = "invalid_phone"
	KeyInvalid      Key = "invalid" // fallback for any validator tag with no specific message
)

// catalog[locale][key] -> message. Every locale here mirrors
// learning.SupportedLocales. Draft translations (same caveat as the
// learning-content seed migration): not reviewed by a native speaker.
var catalog = map[string]map[Key]string{
	"en": {
		KeyRequired:     "is required",
		KeyEmail:        "must be a valid email address",
		KeyTooShort:     "is too short",
		KeyTooLong:      "is too long",
		KeyInvalidPhone: "must be a valid phone number in international format, e.g. +2348012345678",
		KeyInvalid:      "is invalid",
	},
	"pcm": {
		KeyRequired:     "no fit dey empty",
		KeyEmail:        "must be correct email address",
		KeyTooShort:     "too short",
		KeyTooLong:      "too long",
		KeyInvalidPhone: "must be correct phone number for international format, e.g. +2348012345678",
		KeyInvalid:      "no correct",
	},
	"yo": {
		KeyRequired:     "ó ṣe pàtàkì",
		KeyEmail:        "gbọ́dọ̀ jẹ́ àdírẹ́sì ìmẹ́lì tó tọ́",
		KeyTooShort:     "kúrú jù",
		KeyTooLong:      "gùn jù",
		KeyInvalidPhone: "gbọ́dọ̀ jẹ́ nọ́mbà fóònù tó tọ́ ní ọ̀nà àgbáyé, f.a. +2348012345678",
		KeyInvalid:      "kò tọ́",
	},
	"ha": {
		KeyRequired:     "ana bukata",
		KeyEmail:        "dole ne ya zama adireshin imel mai inganci",
		KeyTooShort:     "ya yi gajarta sosai",
		KeyTooLong:      "ya yi tsawo sosai",
		KeyInvalidPhone: "dole ne ya zama lambar waya mai inganci a tsarin ƙasa da ƙasa, misali +2348012345678",
		KeyInvalid:      "ba daidai ba ne",
	},
	"ig": {
		KeyRequired:     "achọrọ ya",
		KeyEmail:        "ga abụrịrị akwụkwọ ozi email ziri ezi",
		KeyTooShort:     "dị mkpụmkpụ",
		KeyTooLong:      "dị ogologo",
		KeyInvalidPhone: "ga abụrịrị nọmba ekwentị ziri ezi n'usoro mba ụwa, dị ka +2348012345678",
		KeyInvalid:      "ezighị ezi",
	},
	"fr": {
		KeyRequired:     "est requis",
		KeyEmail:        "doit être une adresse e-mail valide",
		KeyTooShort:     "est trop court",
		KeyTooLong:      "est trop long",
		KeyInvalidPhone: "doit être un numéro de téléphone valide au format international, par ex. +2348012345678",
		KeyInvalid:      "est invalide",
	},
	"sw": {
		KeyRequired:     "inahitajika",
		KeyEmail:        "lazima iwe anwani sahihi ya barua pepe",
		KeyTooShort:     "ni fupi mno",
		KeyTooLong:      "ni ndefu mno",
		KeyInvalidPhone: "lazima iwe nambari sahihi ya simu katika muundo wa kimataifa, mf. +2348012345678",
		KeyInvalid:      "si sahihi",
	},
	"ar": {
		KeyRequired:     "مطلوب",
		KeyEmail:        "يجب أن يكون عنوان بريد إلكتروني صالحًا",
		KeyTooShort:     "قصير جدًا",
		KeyTooLong:      "طويل جدًا",
		KeyInvalidPhone: "يجب أن يكون رقم هاتف صالحًا بالصيغة الدولية، مثل +2348012345678",
		KeyInvalid:      "غير صالح",
	},
}

// T returns the message for key in locale, falling back to English if
// the locale is unsupported or missing that specific key — the same
// "never fail, just degrade to English" fallback philosophy as the
// learning content API.
func T(locale string, key Key) string {
	if msgs, ok := catalog[locale]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	return catalog["en"][key]
}
