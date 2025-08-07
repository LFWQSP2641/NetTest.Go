package utils

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

func FreeCString(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}
