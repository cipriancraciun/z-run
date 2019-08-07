

package lib


import "strings"




type Scriptlet struct {
	Index uint `json:"id"`
	Label string `json:"label"`
	Kind string `json:"kind"`
	Interpreter string `json:"interpreter"`
	Body string `json:"body,omitempty"`
	Fingerprint string `json:"fingerprint"`
	Source ScriptletSource `json:"source"`
	Hidden bool `json:"hidden"`
}

type ScriptletSource struct {
	Path string `json:"path"`
	LineStart uint `json:"line_start"`
	LineEnd uint `json:"line_end"`
	Fingerprint string `json:"fingerprint"`
}




type Library struct {
	
	Scriptlets []*Scriptlet `json:"scriptlets"`
	
	ScriptletFingerprints []string `json:"fingerprints"`
	ScriptletsByFingerprint map[string]uint `json:"index_by_fingerprint"`
	
	ScriptletLabels []string `json:"labels"`
	ScriptletsByLabel map[string]uint `json:"index_by_label"`
	
	Sources []*Source `json:"sources"`
	SourcesFingerprint string `json:"sources_fingerprint"`
	
	url string
}




type Source struct {
	Path string `json:"path"`
	Executable bool `json:"executable"`
	FingerprintMeta string `json:"fingerprint_meta"`
	FingerprintData string `json:"fingerprint_data"`
}




func NewLibrary () (*Library) {
	return & Library {
			Scriptlets : make ([]*Scriptlet, 0, 1024),
			ScriptletFingerprints : make ([]string, 0, 1024),
			ScriptletsByFingerprint : make (map[string]uint, 1024),
			ScriptletLabels : make ([]string, 0, 1024),
			ScriptletsByLabel : make (map[string]uint, 1024),
		}
}


func (_library *Library) SelectFingerprints () ([]string, error) {
	return _library.ScriptletFingerprints, nil
}

func (_library *Library) SelectLabels () ([]string, error) {
	return _library.ScriptletLabels, nil
}


func (_library *Library) ResolveFullByFingerprint (_fingerprint string) (*Scriptlet, error) {
	if _index, _exists := _library.ScriptletsByFingerprint[_fingerprint]; _exists {
		_scriptlet := _library.Scriptlets[_index]
		return _scriptlet, nil
	} else {
		return nil, nil
	}
}

func (_library *Library) ResolveMetaByFingerprint (_fingerprint string) (*Scriptlet, error) {
	if _scriptlet, _error := _library.ResolveFullByFingerprint (_fingerprint); (_error == nil) && (_scriptlet != nil) {
		_meta := & Scriptlet {}
		*_meta = *_scriptlet
		_meta.Body = ""
		return _meta, nil
	} else {
		return nil, _error
	}
}

func (_library *Library) ResolveBodyByFingerprint (_fingerprint string) (string, bool, error) {
	if _scriptlet, _error := _library.ResolveFullByFingerprint (_fingerprint); (_error == nil) && (_scriptlet != nil) {
		return _scriptlet.Body, true, nil
	} else {
		return "", false, _error
	}
}


func (_library *Library) ResolveFullByLabel (_label string) (*Scriptlet, error) {
	if _index, _exists := _library.ScriptletsByLabel[_label]; _exists {
		_scriptlet := _library.Scriptlets[_index]
		return _scriptlet, nil
	} else {
		return nil, nil
	}
}

func (_library *Library) ResolveMetaByLabel (_label string) (*Scriptlet, error) {
	if _scriptlet, _error := _library.ResolveFullByLabel (_label); (_error == nil) && (_scriptlet != nil) {
		_meta := & Scriptlet {}
		*_meta = *_scriptlet
		_meta.Body = ""
		return _meta, nil
	} else {
		return nil, _error
	}
}

func (_library *Library) ResolveBodyByLabel (_label string) (string, bool, error) {
	if _scriptlet, _error := _library.ResolveFullByLabel (_label); (_error == nil) && (_scriptlet != nil) {
		return _scriptlet.Body, true, nil
	} else {
		return "", false, _error
	}
}

func (_library *Library) ResolveFingerprintByLabel (_label string) (string, bool, error) {
	if _scriptlet, _error := _library.ResolveFullByLabel (_label); (_error == nil) && (_scriptlet != nil) {
		return _scriptlet.Fingerprint, true, nil
	} else {
		return "", false, _error
	}
}


func (_library *Library) Url () (string) {
	return _library.url
}




func includeScriptlet (_library *Library, _scriptlet *Scriptlet) (error) {
	
	if _scriptlet.Label != strings.TrimSpace (_scriptlet.Label) {
		return errorf (0xd8797e9e, "invalid scriptlet label `%s`", _scriptlet.Label)
	}
	if _scriptlet.Label == "" {
		return errorf (0xaede3d8c, "invalid scriptlet label `%s`", _scriptlet.Label)
	}
	if _, _exists := _library.ScriptletsByLabel[_scriptlet.Label]; _exists {
		return errorf (0x883f9a7f, "duplicate scriptlet label `%s`", _scriptlet.Label)
	}
	if (_scriptlet.Interpreter == "<shell>") || (_scriptlet.Interpreter == "<print>") {
		// NOP
	} else {
		return errorf (0xbf289098, "invalid scriptlet interpreter `%s`", _scriptlet.Interpreter)
	}
	
	_fingerprint := NewFingerprinter () .StringWithLen (_scriptlet.Label) .StringWithLen (_scriptlet.Kind) .StringWithLen (_scriptlet.Interpreter) .StringWithLen (_scriptlet.Body) .Build ()
	
	if _, _exists := _library.ScriptletsByFingerprint[_fingerprint]; _exists {
		return nil
	}
	
	switch _scriptlet.Kind {
		case "executable" :
			// NOP
		case "generator" :
			_scriptlet.Kind = "generator-pending"
		default :
			return errorf (0x4b8aacf2, "invalid scriptlet kind `%s`", _scriptlet.Kind)
	}
	
	_scriptlet.Index = uint (len (_library.Scriptlets))
	_scriptlet.Fingerprint = _fingerprint
	
	_library.Scriptlets = append (_library.Scriptlets, _scriptlet)
	
	_library.ScriptletFingerprints = append (_library.ScriptletFingerprints, _scriptlet.Fingerprint)
	if !_scriptlet.Hidden {
		_library.ScriptletLabels = append (_library.ScriptletLabels, _scriptlet.Label)
	}
	
	_library.ScriptletsByFingerprint[_scriptlet.Fingerprint] = _scriptlet.Index
	_library.ScriptletsByLabel[_scriptlet.Label] = _scriptlet.Index
	
	return nil
}
