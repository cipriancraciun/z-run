
// +build !linux


package zrun


import "fmt"
import "os"
import "path"
import "runtime"
import "syscall"




// FIXME:  Merge with Linux variant!

func createPipe (_size int, _cacheRoot string) (int, *os.File, *Error) {
	
	var _interpreterScriptInput int
	var _interpreterScriptOutput *os.File
	var _interpreterScriptDescriptors [2]int
	
	_maxPipeSize := 0
	switch runtime.GOOS {
		case "darwin" :
			_maxPipeSize = 16 * 1024
		case "freebsd" :
			_maxPipeSize = 512
		case "openbsd" :
			_maxPipeSize = 16 * 1024
		default :
			_maxPipeSize = 0
	}
	
	if _size <= _maxPipeSize {
		if _error := syscall.Pipe (_interpreterScriptDescriptors[:]); _error == nil {
			_interpreterScriptInput = _interpreterScriptDescriptors[0]
			_interpreterScriptOutput = os.NewFile (uintptr (_interpreterScriptDescriptors[1]), "")
		} else {
			return -1, nil, errorw (0xece645ff, _error)
		}
	} else {
		if _cacheRoot == "" {
			// FIXME:  We should make sure that the cache path is never empty!
			panic (0xd6f17610)
		}
		_temporaryPath := path.Join (_cacheRoot, generateRandomToken () + ".buffer")
		if _descriptor, _error := syscall.Open (_temporaryPath, syscall.O_CREAT | syscall.O_EXCL | syscall.O_WRONLY, 0600); _error == nil {
			_interpreterScriptOutput = os.NewFile (uintptr (_descriptor), "")
		} else {
			return -1, nil, errorw (0x2b19feaa, _error)
		}
		if _descriptor, _error := syscall.Open (_temporaryPath, syscall.O_RDONLY, 0600); _error == nil {
			_interpreterScriptInput = _descriptor
		} else {
			// FIXME:  Here we leak the first descriptor!
			return -1, nil, errorw (0x694ce572, _error)
		}
		if _error := syscall.Unlink (_temporaryPath); _error != nil {
			// FIXME:  Here we leak both descriptors!
			return -1, nil, errorw (0xc5afd6fd, _error)
		}
	}
	
	if _, _error := os.Stat (fmt.Sprintf ("/dev/fd/%d", _interpreterScriptInput)); _error != nil {
		// FIXME:  Here we leak both descriptors!
		return -1, nil, errorw (0x5ea72831, _error)
	}
	
	return _interpreterScriptInput, _interpreterScriptOutput, nil
}


