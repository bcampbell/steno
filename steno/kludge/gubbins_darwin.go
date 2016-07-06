package kludge

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation

#import <Foundation/Foundation.h>

#include <stdio.h>

const char* gimme()
{
    char* buf = malloc(256);
    if (buf==0) {
        return buf;
    }

    NSAutoreleasePool * pool = [[NSAutoreleasePool alloc] init];
    bool result = false;
    NSBundle * bundle = [NSBundle mainBundle];
    printf("bundle: %p\n",bundle);
    NSString * s = [bundle resourcePath];
    printf("s: %p\n",s);
    result = ( [s getFileSystemRepresentation: buf maxLength: 256] == YES );
    printf("buf: '%s'\n",buf);
    [pool drain];
    if (!result) {
        return 0;
    }
    return buf;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func DataPath() (string, error) {
	s := C.gimme()
	if s == nil {
		return "", fmt.Errorf("Poop. bundle path kludge thingy failed.\n")
	}
	defer C.free(unsafe.Pointer(s))

	gs := C.GoString(s)

	return gs, nil
}

// get (or create) per-user directory (eg "$HOME/.steno")
// TODO: use /System/Library/Steno or whatever it is instead of
// generic unix version
func PerUserPath() (string, error) {
	home := os.GetEnv("HOME")
	if home == "" {
		return "", fmt.Errorf("$HOME not set")
	}
	dir := filepath.Join(home, ".steno")
	// create dir if if doesn't already exist
	err := os.MkDirAll(dir, 0755)
	if err != nil {
		return "", err
	}
	return dir, nil
}
