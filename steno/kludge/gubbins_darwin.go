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


func DataPath() (string,error) {
	s := C.gimme()
    if s==nil {
        return "", fmt.Errorf("Poop. bundle path kludge thingy failed.\n")
    }
    defer C.free(unsafe.Pointer(s) )

    gs := C.GoString(s)

    return gs,nil
}
