package main

import (
	dns "nettest/pkg/dns"
	utils "nettest/pkg/utils"
	"unsafe"
)
import "C"

//export DnsRequest
func DnsRequest(server *C.char, qname *C.char, qtype *C.char, qclass *C.char) *C.char {
	goServer := C.GoString(server)
	goQname := C.GoString(qname)
	goQtype := C.GoString(qtype)
	goQclass := C.GoString(qclass)
	result := dns.DnsRequest(goServer, goQname, goQtype, goQclass)
	return C.CString(result)
}

//export DnsRequestOverSocks5
func DnsRequestOverSocks5(proxy *C.char, server *C.char, qname *C.char, qtype *C.char, qclass *C.char) *C.char {
	goProxy := C.GoString(proxy)
	goServer := C.GoString(server)
	goQname := C.GoString(qname)
	goQtype := C.GoString(qtype)
	goQclass := C.GoString(qclass)
	result := dns.DnsRequestOverSocks5(goProxy, goServer, goQname, goQtype, goQclass)
	return C.CString(result)
}

//export FreeCString
func FreeCString(s *C.char) {
	utils.FreeCString(unsafe.Pointer(s))
}

func main() {}
