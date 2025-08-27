package main

/*
#include <stdint.h>
typedef void (*DnsCallback)(void*, char*);
static void callDnsCallback(DnsCallback cb, void* userData, char* result) {
    cb(userData, result);
}
*/
import "C"
import (
	dns "nettest/pkg/dns"
	utils "nettest/pkg/utils"
	"unsafe"
)

//export DnsRequest
func DnsRequest(server, qname, qtype, qclass, sni, clientSubnet *C.char) *C.char {
	goServer := C.GoString(server)
	goQname := C.GoString(qname)
	goQtype := C.GoString(qtype)
	goQclass := C.GoString(qclass)
	goSNI := C.GoString(sni)
	goClientSubnet := C.GoString(clientSubnet)
	result := dns.DnsRequest(goServer, goQname, goQtype, goQclass, goSNI, goClientSubnet)
	return C.CString(result)
}

//export DnsRequestJson
func DnsRequestJson(json *C.char) *C.char {
	goJSON := C.GoString(json)
	result := dns.DnsRequestJson(goJSON)
	return C.CString(result)
}

//export DnsRequestOverSocks5
func DnsRequestOverSocks5(proxy, server, qname, qtype, qclass, sni, clientSubnet *C.char) *C.char {
	goProxy := C.GoString(proxy)
	goServer := C.GoString(server)
	goQname := C.GoString(qname)
	goQtype := C.GoString(qtype)
	goQclass := C.GoString(qclass)
	goSNI := C.GoString(sni)
	goClientSubnet := C.GoString(clientSubnet)
	result := dns.DnsRequestOverSocks5(goProxy, goServer, goQname, goQtype, goQclass, goSNI, goClientSubnet)
	return C.CString(result)
}

//export DnsRequestAsync
func DnsRequestAsync(server, qname, qtype, qclass, sni, clientSubnet *C.char, cb C.DnsCallback, userData unsafe.Pointer) {
	goServer := C.GoString(server)
	goQname := C.GoString(qname)
	goQtype := C.GoString(qtype)
	goQclass := C.GoString(qclass)
	goSNI := C.GoString(sni)
	goClientSubnet := C.GoString(clientSubnet)
	go func() {
		result := dns.DnsRequest(goServer, goQname, goQtype, goQclass, goSNI, goClientSubnet)
		cResult := C.CString(result)
		C.callDnsCallback(cb, userData, cResult)
		utils.FreeCString(unsafe.Pointer(cResult))
	}()
}

//export DnsRequestOverSocks5Async
func DnsRequestOverSocks5Async(proxy, server, qname, qtype, qclass, sni, clientSubnet *C.char, cb C.DnsCallback, userData unsafe.Pointer) {
	goProxy := C.GoString(proxy)
	goServer := C.GoString(server)
	goQname := C.GoString(qname)
	goQtype := C.GoString(qtype)
	goQclass := C.GoString(qclass)
	goSNI := C.GoString(sni)
	goClientSubnet := C.GoString(clientSubnet)
	go func() {
		result := dns.DnsRequestOverSocks5(goProxy, goServer, goQname, goQtype, goQclass, goSNI, goClientSubnet)
		cResult := C.CString(result)
		C.callDnsCallback(cb, userData, cResult)
		utils.FreeCString(unsafe.Pointer(cResult))
	}()
}

//export FreeCString
func FreeCString(s *C.char) {
	utils.FreeCString(unsafe.Pointer(s))
}

func main() {}
