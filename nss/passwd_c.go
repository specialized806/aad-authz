package main

/*
#cgo LDFLAGS: -fPIC
#include <nss.h>
#include <pwd.h>
#include <errno.h>

typedef enum nss_status nss_status;
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/ubuntu/aad-auth/internal/cache"
	"github.com/ubuntu/aad-auth/internal/nss"
	"github.com/ubuntu/aad-auth/internal/passwd"
)

//export _nss_aad_getpwnam_r
func _nss_aad_getpwnam_r(name *C.char, pwd *C.struct_passwd, buf *C.char, buflen C.size_t, result **C.struct_passwd) C.nss_status {
	n := C.GoString(name)
	p, err := passwd.NewByName(n)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err) // TODO: log
		return (C.nss_status)(nss.ErrToCStatus(err))
	}
	if err = p.ToCpasswd(
		passwd.CPasswd(unsafe.Pointer(pwd)),
		(*passwd.CChar)(buf),
		passwd.CSizeT(buflen),
		(*passwd.CPasswd)(unsafe.Pointer(result))); err != nil {
		fmt.Printf("ERROR: %v\n", err) // TODO: log
		return (C.nss_status)(nss.ErrToCStatus(err))
	}

	return C.NSS_STATUS_SUCCESS
}

//export _nss_aad_getpwuid_r
func _nss_aad_getpwuid_r(uid C.uid_t, pwd *C.struct_passwd, buf *C.char, buflen C.size_t, result **C.struct_passwd) C.nss_status {
	p, err := passwd.NewByUID(uint(uid))
	if err != nil {
		fmt.Printf("ERROR: %v\n", err) // TODO: log
		return (C.nss_status)(nss.ErrToCStatus(err))
	}
	if err = p.ToCpasswd(
		passwd.CPasswd(unsafe.Pointer(pwd)),
		(*passwd.CChar)(buf),
		passwd.CSizeT(buflen),
		(*passwd.CPasswd)(unsafe.Pointer(result))); err != nil {
		fmt.Printf("ERROR: %v\n", err) // TODO: log
		return (C.nss_status)(nss.ErrToCStatus(err))
	}

	return C.NSS_STATUS_SUCCESS
}

//export _nss_aad_setpwent
func _nss_aad_setpwent() {
	// Initialization of the database is done in the read primitive
}

//export _nss_aad_endpwent
func _nss_aad_endpwent() {
	// Closing the database is done in the read primitive
}

//export _nss_aad_getpwent_r
func _nss_aad_getpwent_r(pwbuf *C.struct_passwd, buf *C.char, buflen C.size_t, pwbufp **C.struct_passwd) C.nss_status {
	p, err := passwd.NextEntry()
	if errors.Is(err, cache.ErrNoEnt) {
		return C.ENOENT
	}
	if err != nil {
		fmt.Printf("ERROR: %v\n", err) // TODO: log
		return (C.nss_status)(nss.ErrToCStatus(err))
	}

	if err = p.ToCpasswd(
		passwd.CPasswd(unsafe.Pointer(pwbuf)),
		(*passwd.CChar)(buf),
		passwd.CSizeT(buflen),
		(*passwd.CPasswd)(unsafe.Pointer(pwbufp))); err != nil {
		fmt.Printf("ERROR: %v\n", err) // TODO: log
		return (C.nss_status)(nss.ErrToCStatus(err))
	}

	return C.NSS_STATUS_SUCCESS
}
