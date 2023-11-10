package schema

type _element struct {
	name    string
	enabled bool
	updated bool

	patchAllowed   *bool
	enabledPatched bool // indicates that we need to include new value in request
}

func (e *_element) isPatchAllowed() bool {
	return e.patchAllowed == nil || *e.patchAllowed
}

func (e *_element) setEnabled(value bool) {
	if e.isPatchAllowed() {
		if value != e.enabled {
			e.enabled = value
			e.updated = true
			e.enabledPatched = true
		}
	}
}
