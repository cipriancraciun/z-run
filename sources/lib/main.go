

package lib


import "log"
import "os"
import "strings"
import "syscall"
import "unicode"
import "unicode/utf8"




type Context struct {
	selfExecutable string
	selfArgument0 string
	selfArguments []string
	selfEnvironment map[string]string
	cleanArguments []string
	cleanEnvironment map[string]string
	cacheRoot string
	cacheEnabled bool
	terminal string
}




func main_0 (_executable string, _argument0 string, _arguments []string, _environment map[string]string) (error) {
	
	var _command string
	var _scriptlet string
	var _sourcePath string
	var _cacheRoot string
	var _cachePath string
	
	var _cleanArguments []string
	var _cleanEnvironment map[string]string = make (map[string]string, len (_environment))
	var _terminal string
	
	for _name, _value := range _environment {
		
		var _nameCanonical string
		{
			_nameCanonical = strings.ToUpper (_name)
			_nameCanonical = strings.ReplaceAll (_nameCanonical, "-", "_")
			for {
				_nameCanonical_0 := strings.ReplaceAll (_nameCanonical, "__", "_")
				if _nameCanonical != _nameCanonical_0 {
					_nameCanonical = _nameCanonical_0
				} else {
					break
				}
			}
			_nameCanonical = strings.Replace (_nameCanonical, "Z_RUN", "ZRUN", 1)
		}
		
		if strings.HasPrefix (_nameCanonical, "ZRUN_") || strings.HasPrefix (_nameCanonical, "_ZRUN_") {
			
			if _name != _nameCanonical {
				logf ('w', 0x9bc8b3da, "environment variable does not have canonical name;  expected `%s`, encountered `%s`!", _nameCanonical, _name)
			}
			
			switch _nameCanonical {
				case "ZRUN_SOURCE" :
					_sourcePath = _value
				case "ZRUN_LIBRARY" :
					_cachePath = _value
				case "ZRUN_EXECUTABLE" :
					if _executable != _value {
						logf ('w', 0xfb1f0645, "environment variable mismatched:  `%s`;  expected `%s`, encountered `%s`!", _nameCanonical, _executable, _value)
					}
				case "ZRUN_CACHE" :
					_cacheRoot = _value
				case "ZRUN_TERM" :
					_terminal = _value
				default :
					logf ('w', 0xafe247b0, "environment variable unknown:  `%s` with value `%s`", _nameCanonical, _value)
			}
			
		} else if strings.HasPrefix (_nameCanonical, "X_RUN_") || strings.HasPrefix (_nameCanonical, "_X_RUN_") {
			
			if _name != _nameCanonical {
				logf ('w', 0x37850eb3, "environment variable does not have canonical name;  expected `%s`, encountered `%s`!", _nameCanonical, _name)
			}
			
			switch _nameCanonical {
				case "X_RUN_COMMANDS" :
					_sourcePath = _value
				case "X_RUN_ACTION" :
					_command = "legacy:" + _value
				case "X_RUN_TERM" :
					_terminal = _value
				default :
					logf ('w', 0xdf61b057, "environment variable unknown:  `%s` with value `%s`", _nameCanonical, _value)
			}
			
		} else {
			_cleanEnvironment[_name] = _value
		}
	}
	
	for _index, _argument := range _arguments {
		
		if _argument == "--" {
			_cleanArguments = _arguments[_index + 1:]
			break
			
		} else if strings.HasPrefix (_argument, "-") {
			if _command != "" {
				return errorf (0xae04b5ff, "unexpected argument `%s`", _argument)
			}
			if strings.HasPrefix (_argument, "--source=") {
				_sourcePath = _argument[len ("--source="):]
			} else if strings.HasPrefix (_argument, "--library=") {
				_cachePath = _argument[len ("--library="):]
			} else {
				return errorf (0x33555ffb, "invalid argument `%s`", _argument)
			}
			
		} else if strings.HasPrefix (_argument, "::") {
			if _command == "" {
				_command = "execute"
			}
			_scriptlet = _argument
			_cleanArguments = _arguments[_index + 1:]
			break
			
		} else {
			if _command == "" {
				switch _argument {
					
					case "execute" :
						_command = "execute"
						continue
					
					case "select-execute" :
						_command = "select-execute"
						_index += 1
					
					case "select-label", "select" :
						_command = "select-label"
						_index += 1
					
					case "export-script" :
						_command = "export-script"
						continue
					
					case "export-labels-list", "export-labels", "list" :
						_command = "export-labels-list"
						_index += 1
					
					case "parse-library-json", "parse-library", "parse" :
						_command = "parse-library-json"
						_index += 1
					
					case "export-library-json", "export-library" :
						_command = "export-library-json"
						_index += 1
					
					case "export-library-cdb" :
						_command = "export-library-cdb"
						_index += 1
				}
			} else {
				_scriptlet = _argument
				_index += 1
			}
			_cleanArguments = _arguments[_index:]
			break
		}
	}
	
	if _scriptlet != "" {
		if strings.HasPrefix (_scriptlet, ":: ") {
			_scriptlet = _scriptlet[3:]
		} else {
			return errorf (0x72ad17f7, "invalid scriptlet label `%s`", _scriptlet)
		}
	}
	
	if (_command == "") && (_scriptlet == "") {
		_command = "select-execute"
	}
	
	_cacheEnabled := true
	if _command == "parse-library-json" {
		_cacheEnabled = false
	}
	if _cacheEnabled {
		if _cacheRoot == "" {
			if _cacheRoot_0, _error := resolveCache (); _error == nil {
				_cacheRoot = _cacheRoot_0
			} else {
				return _error
			}
		}
	} else {
		_cacheRoot = ""
		_cachePath = ""
	}
	
	if _terminal == "" {
		_terminal, _ = _cleanEnvironment["TERM"]
	}
	if _terminal == "dumb" {
		_terminal = ""
	}
	
	if _cachePath != "" {
		if _sourcePath != "" {
			logf ('w', 0x1fe0b572, "both library and source path specified;  using library!")
			_sourcePath = ""
		}
	}
	
	_context := & Context {
			selfExecutable : _executable,
			selfArguments : _arguments,
			selfEnvironment : _environment,
			cleanArguments : _cleanArguments,
			cleanEnvironment : _cleanEnvironment,
			cacheRoot : _cacheRoot,
			cacheEnabled : _cacheEnabled,
			terminal : _terminal,
		}
	
	var _library LibraryStore
	if _cachePath != "" {
//		logf ('d', 0xeeedb7f0, "opening library...")
		if _library_0, _error := resolveLibraryCached (_cachePath); _error == nil {
			_library = _library_0
		} else {
			return _error
		}
	} else {
//		logf ('d', 0x93dbfd8c, "resolving library...")
		if _library_0, _error := resolveLibrary (_sourcePath, _context); _error == nil {
			_library = _library_0
		} else {
			return _error
		}
	}
	
	switch _command {
		
		case "execute" :
			if _scriptlet == "" {
				return errorf (0x39718e70, "execute:  expected scriptlet")
			}
			return doExecute (_library, _scriptlet, _context)
		
		case "select-execute" :
			if (_scriptlet != "") || (len (_cleanArguments) != 0) {
				return errorf (0x203e410a, "execute:  unexpected scriptlet or arguments")
			}
			return doSelectExecute (_library, _context)
		
		case "select-label" :
			if (_scriptlet != "") || (len (_cleanArguments) != 0) {
				return errorf (0x2d19b1bc, "select:  unexpected scriptlet or arguments")
			}
			return doSelectLabel (_library, os.Stdout, _context)
		
		case "export-script" :
			if _scriptlet == "" {
				return errorf (0xf24640a2, "export:  expected scriptlet")
			}
			if len (_cleanArguments) != 0 {
				return errorf (0xcf8db3c0, "export:  unexpected arguments")
			}
			return doExportScript (_library, _scriptlet, os.Stdout, _context)
		
		case "export-labels-list" :
			if (_scriptlet != "") || (len (_cleanArguments) != 0) {
				return errorf (0xf7b9c7f3, "list:  unexpected scriptlet or arguments")
			}
			return doExportLabelsList (_library, os.Stdout, _context)
		
		case "parse-library-json", "export-library-json" :
			if (_scriptlet != "") || (len (_cleanArguments) != 0) {
				return errorf (0x400ec122, "export:  unexpected scriptlet or arguments")
			}
			switch _command {
				case "parse-library-json" :
					return doExportLibraryJson (_library, os.Stdout, _context)
				case "export-library-json" :
					return doExportLibraryStore (_library, NewJsonStreamStoreOutput (os.Stdout, nil), _context)
				default :
					panic (0xda7243ef)
			}
		
		case "export-library-cdb" :
			if _scriptlet != "" {
				return errorf (0x492ac50e, "export:  unexpected scriptlet")
			}
			if len (_cleanArguments) != 1 {
				return errorf (0xf76f4459, "export:  expected database path")
			}
			return doExportLibraryCdb (_library, _cleanArguments[0], _context)
		
		case "legacy:output-selection-and-command" :
			if len (_cleanArguments) != 0 {
				return errorf (0xe4f7e6f5, "export:  unexpected arguments")
			}
			return doSelectLegacyOutput (_library, _scriptlet, os.Stdout, _context)
		
		case "" :
			return errorf (0x5d2a4326, "expected command")
		
		default :
			return errorf (0x66cf8700, "unexpected command `%s`", _command)
	}
}




func Main () () {
	
	log.SetFlags (0)
	
	var _executable string
	if _executable_0, _error := os.Executable (); _error == nil {
		_executable = _executable_0
	} else {
		panic (abortError (_error))
	}
	
	_argument0 := os.Args[0]
	switch _argument0 {
		case "[z-run:select]" :
			if _error := fzfSelectMain (); _error != nil {
				panic (abortError (_error))
			} else {
				panic (0x2346ca3f)
			}
		case "[z-run]" :
			// NOP
		default :
			_arguments := os.Args
			_arguments[0] = "[z-run]"
			_environment := os.Environ ()
			if _error := syscall.Exec (_executable, _arguments, _environment); _error != nil {
				panic (abortError (_error))
			} else {
				panic (0xe13aab5f)
			}
	}
	
	_arguments := append ([]string (nil), os.Args[1:] ...)
	
	_environment := make (map[string]string, 128)
	for _, _variable := range os.Environ () {
		if _splitIndex := strings.IndexByte (_variable, '='); _splitIndex >= 0 {
			
			_name := _variable[:_splitIndex]
			_value := _variable[_splitIndex + 1:]
			
			_nameTrimmed := strings.TrimSpace (_name)
			if _name != _nameTrimmed {
				logf ('w', 0x1d362f26, "invalid environment variable (name has spaces):  `%s`", _name)
				_name = _nameTrimmed
			}
			if strings.IndexFunc (_name, func (r rune) (bool) { return unicode.IsSpace (r) || (r > unicode.MaxASCII) }) >= 0 {
				logf ('w', 0x81ac6f2e, "invalid environment variable (name is not ASCII):  `%s`", _name)
			}
			
			if _name == "" {
				logf ('w', 0x0ffb0031, "invalid environment variable (name empty):  `%s`", _variable)
			} else if ! utf8.Valid ([]byte (_name)) {
				logf ('w', 0x54278534, "invalid environment variable (name invalid UTF-c):  `%s`", _name)
			} else if ! utf8.Valid ([]byte (_value)) {
				logf ('w', 0x785ba004, "invalid environment variable (value invalid UTF-c):  `%s`", _name)
			} else if _value == "" {
//				logf ('w', 0xfe658d34, "invalid environment variable (value empty):  `%s`", _name)
			} else if _, _exists := _environment[_name]; _exists {
				logf ('w', 0x7e7e41a5, "invalid environment variable (name duplicate):  `%s`", _name)
			} else {
				_environment[_nameTrimmed] = _value
			}
			
		} else {
			logf ('w', 0xe745517c, "invalid environment variable (missing `=`):  `%s`", _variable)
		}
	}
	
	if _error := main_0 (_executable, _argument0, _arguments, _environment); _error == nil {
		os.Exit (0)
	} else {
		panic (abortError (_error))
	}
}
