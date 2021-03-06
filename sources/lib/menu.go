

package zrun


import "bufio"
import "fmt"
import "io"
import "os"
import "os/exec"
import "path/filepath"
import "strings"
import "sync"


import isatty "github.com/mattn/go-isatty"
import "github.com/eiannone/keyboard"




func menuMain (_executable string, _arguments []string, _environment map[string]string) (*Error) {
	
	if len (_arguments) != 1 {
		return errorf (0x6b439ede, "invalid arguments")
	}
	
	_inputs := make ([]string, 0, 1024)
	if _stream, _error := os.Open (_arguments[0]); _error == nil {
		defer _stream.Close ()
		_reader := bufio.NewReader (_stream)
		for {
			if _line, _error := _reader.ReadString ('\n'); _error == nil {
				_input := strings.TrimRight (_line, "\n")
				_inputs = append (_inputs, _input)
			} else if _error == io.EOF {
				if _line == "" {
					break
				} else {
					errorf (0x1f57b1db, "expected proper line")
				}
			} else {
				return errorw (0x3dd692c8, _error)
			}
		}
	}
	
	_context := & Context {
			selfExecutable : _executable,
			cleanEnvironment : _environment,
		}
	
	if _paths, _ok := _environment["PATH"]; _ok {
		_context.executablePaths = filepath.SplitList (_paths)
	}
	if _terminal, _ok := _environment["TERM"]; _ok {
		_context.terminal = _terminal
	}
	
	if _outputs, _error := menuSelect (_inputs, _context); _error == nil {
		for _, _output := range _outputs {
			if _, _error := io.WriteString (os.Stdout, _output + "\n"); _error != nil {
				return errorw (0xeb4af0b7, _error)
			}
		}
		os.Exit (0)
		panic (0x1ad77faa)
	} else {
		return _error
	}
}




func menuSelect (_inputs []string, _context *Context) ([]string, *Error) {
	
	_inputsChannel := make (chan string, 1024)
	_outputsChannel := make (chan string, 1024)
	_outputs := make ([]string, 0, 1024)
	
	_waiter := & sync.WaitGroup {}
	
	_waiter.Add (1)
	go func () () {
		for _, _input := range _inputs {
			_inputsChannel <- _input
		}
		close (_inputsChannel)
		_waiter.Done ()
	} ()
	
	_waiter.Add (1)
	go func () () {
		for {
			_output, _ok := <- _outputsChannel
			if _ok {
				_outputs = append (_outputs, _output)
			} else {
				break
			}
		}
		_waiter.Done ()
	} ()
	
	_error := menuSelect_0 (_inputsChannel, _outputsChannel, _context);
	
	close (_outputsChannel)
	_waiter.Wait ()
	
	if _error == nil {
		return _outputs, nil
	} else {
		return nil, _error
	}
}


func menuSelect_0 (_inputsChannel <-chan string, _outputsChannel chan<- string, _context *Context) (*Error) {
	
	_hasTerminal := (_context.terminal != "") && (_context.terminal != "dumb")
	
	if _hasTerminal {
		if ! isatty.IsTerminal (os.Stderr.Fd ()) {
			return errorf (0xfc026596, "stderr is not a TTY")
		}
	}
	
	_command := & exec.Cmd {
			Stdin : nil,
			Stdout : nil,
			Stderr : os.Stderr,
			Dir : _context.workspace,
		}
	
	_commandFzf := false
	if _hasTerminal {
		_commandFzf = true
		_command.Path = _context.selfExecutable
		_command.Args = []string {
				"[z-run:select]",
			}
	} else if _path, _error := resolveExecutable ("z-run--select", _context.executablePaths); _error == nil {
		_command.Path = _path
		_command.Args = []string {
				"[z-run:select]",
			}
	} else if _path, _error := resolveExecutable ("rofi", _context.executablePaths); _error == nil {
		_command.Path = _path
		_command.Args = []string {
				"[z-run:select]",
				"-dmenu",
				"-p", "",
				"-i",
				"-no-custom",
				"-matching-negate-char", "\\x0",
			}
	} else if _path, _error := resolveExecutable ("dmenu", _context.executablePaths); _error == nil {
		_command.Path = _path
		_command.Args = []string {
				"[z-run:select]",
				"-p", "",
				"-l", "16",
				"-i",
			}
	} else if _path, _error := resolveExecutable ("choose", _context.executablePaths); _error == nil {
		_command.Path = _path
		_command.Args = []string {
				"[z-run:select]",
				"-n", "16",
				"-w", "40",
				"-s", "14",
			}
	} else {
		return errorf (0xb91714f7, "unresolved `z-run--select`")
	}
	
	if _command.Env == nil {
		_command.Env = prepareEnvironment (_context)
	}
	
//	logf ('d', 0x5cbde167, "%v", _command.Path)
//	logf ('d', 0x44b3328a, "%v", _command.Args[0])
//	logf ('d', 0x3cc16861, "%v", _command.Args[1:])
//	logf ('d', 0x8f4e574f, "%v", _command.Env)
	
	if _exitCode, _, _outputsCount, _error := processExecuteAndPipe (_command, _inputsChannel, _outputsChannel, true); _error == nil {
		if _commandFzf {
			switch _exitCode {
				case 0 :
					if _outputsCount == 0 {
						return errorf (0xbb7ff442, "invalid outputs")
					}
				case 1 :
					if _outputsCount != 0 {
						return errorf (0x6bd364da, "invalid outputs")
					}
				case 130 :
					if _outputsCount != 0 {
						return errorf (0xac4b1681, "invalid outputs")
					}
				case 2 :
					return errorf (0x85cabb2a, "failed")
				default :
					return errorf (0xef9908df, "failed")
			}
		} else {
			switch _exitCode {
				case 0 :
					if _outputsCount == 0 {
						return errorf (0x4e0abce6, "invalid outputs")
					}
				case 1 :
					if _outputsCount != 0 {
						return errorf (0x6ad0fdcd, "invalid outputs")
					}
				default :
					return errorf (0xb156b11d, "failed")
			}
		}
	} else {
		return _error
	}
	
	return nil
}




func menuQuit (_context *Context) (bool, *Error) {
	
	if _outputs, _error := menuSelect ([]string { "quit?" }, _context); _error == nil {
		if len (_outputs) == 0 {
			return false, nil
		} else if (len (_outputs) == 1) && (_outputs[0] == "quit?") {
			return true, nil
		} else {
			return false, errorf (0x272fb981, "invalid outputs")
		}
	} else {
		return false, _error
	}
}




func menuPause (_context *Context) (bool, *Error) {
	
	// FIMXE:  Find a more proper implementation for this!
	
	_term := os.Getenv ("TERM")
	if (_term == "dumb") || (_term == "") {
		return false, nil
	}
	
	if ! isatty.IsTerminal (os.Stderr.Fd ()) {
		return false, nil
	}
	
	fmt.Fprintf (os.Stderr, "\n---- << press return to continue... >>")
	os.Stderr.Sync ()
	
	_loop : for {
		
		var _key keyboard.Key
		if _, _key_0, _error := keyboard.GetSingleKey (); _error == nil {
			_key = _key_0
		} else {
			fmt.Fprintf (os.Stderr, "\n\n")
			os.Stderr.Sync ()
			return false, errorw (0x933cd3f2, _error)
		}
		
		switch _key {
			case keyboard.KeyEnter :
				break _loop;
			case keyboard.KeyEsc :
				break _loop;
		}
	}
	
	fmt.Fprintf (os.Stderr, "\r------------------------------------------------------------------------------\n\n")
	os.Stderr.Sync ()
	
	return false, nil
}

