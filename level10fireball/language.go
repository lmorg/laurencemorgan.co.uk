// language.go
package main

var lang_strings = map[string]string{
	"char_counter_desktop_en":   " characters remaining.",
	"char_counter_mobile_en":    " char's left.",
	"char_limit_desktop_en":     " character limit.",
	"char_limit_mobile_en":      " char limit.",
	"required_field_desktop_en": "Required field",
	"optional_field_desktop_en": "Optional field",
	"required_field_mobile_en":  "Required",
	"optional_field_mobile_en":  "Optional",
}

func lang(session *Session, phrase string) DisplayText {
	return DisplayText{lang_strings[phrase+"_"+session.Theme+"_"+session.Language]}
}

const (
	URL_PM = "pm"
)
