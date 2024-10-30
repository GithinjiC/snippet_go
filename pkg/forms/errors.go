package forms

import ()

type errors map[string][]string // hold validation errors from forms. Name of form==key

// Add error msgs to map
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// retriueve first error msg for a field
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
