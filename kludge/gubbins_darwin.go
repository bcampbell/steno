package kludge

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation

#import <Foundation/Foundation.h>

#include <stdio.h>

bool bundle_path(char* buf, size_t bufsize)
{
    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];
    bool result = false;
    NSBundle * bundle = [NSBundle mainBundle];
    //printf("bundle: %p\n",bundle);
    NSString * s = [bundle resourcePath];
    //printf("s: %p\n",s);
    bool success = ( [s getFileSystemRepresentation: buf maxLength: bufsize] == YES );
    //printf("buf: '%s'\n",buf);
    [pool drain];
    return success;
}

bool app_support_path( char * buf, size_t bufsize )
{
    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init] ;
    NSArray *paths = NSSearchPathForDirectoriesInDomains(NSApplicationSupportDirectory ,NSUserDomainMask, YES);
    NSString *path = [paths objectAtIndex:0];
    bool success = ( [path getFileSystemRepresentation:buf maxLength: bufsize] == YES );
    [pool drain] ;
    return success;
}

*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
)

func DataPath() (string, error) {
	bufsize := C.size_t(512)
	buf := C.malloc(bufsize)
	if buf == nil {
		return "", fmt.Errorf("malloc failed")
	}
	defer C.free(buf)
	//	defer C.free(unsafe.Pointer(buf))

	ok := C.bundle_path((*C.char)(buf), bufsize)
	if !ok {
		return "", fmt.Errorf("bundle_path() failed")
	}

	dir := C.GoString((*C.char)(buf))
	return dir, nil
}

// get (or create) per-user directory (eg "~/Library/Application Support/Steno")
func PerUserPath() (string, error) {
	bufsize := C.size_t(512)
	buf := C.malloc(bufsize)
	if buf == nil {
		return "", fmt.Errorf("malloc failed")
	}
	defer C.free(buf)
	//	defer C.free(unsafe.Pointer(buf))

	ok := C.app_support_path((*C.char)(buf), bufsize)
	if !ok {
		return "", fmt.Errorf("app_support_path() failed")
	}

	dir := C.GoString((*C.char)(buf))

	dir = filepath.Join(dir, "Steno")
	// create dir if if doesn't already exist
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return dir, nil
}

// path to any external tool binaries (eg fasttext)
// TODO: should be alongside steno binary in bundle?
func BinPath() (string, error) {
	datPath, err := DataPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(datPath, "bin"), nil
}
