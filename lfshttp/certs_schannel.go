//go:build windows
// +build windows

package lfshttp

import (
	"crypto"
	"io"
	"crypto/tls"
	"crypto/rsa"
	"unsafe"
	"syscall"
	"errors"
	"runtime"
	"github.com/rubyist/tracerx"
	"fmt"
	"strings"
)

//
// Some defines from wincrypt.h
//
const ALG_CLASS_HASH	= (4<<13)
const ALG_TYPE_ANY		= 0

const ALG_SID_SHA1		= 4
const ALG_SID_SHA_256	= 12
const ALG_SID_SHA_512	= 14

const CALG_SHA1 = (ALG_CLASS_HASH | ALG_TYPE_ANY | ALG_SID_SHA1)
const CALG_SHA256 = (ALG_CLASS_HASH | ALG_TYPE_ANY | ALG_SID_SHA_256)
const CALG_SHA512 = (ALG_CLASS_HASH | ALG_TYPE_ANY | ALG_SID_SHA_512)

const CERT_KEY_PROV_INFO_PROP_ID = 2

const CRYPT_ACQUIRE_CACHE_FLAG = 1

const CRYPT_E_NOT_FOUND = 0x80092004

const HP_HASHVAL = 2

//
// Load CAPI functions from crypt32.dll and advapi32.dll
//
var (
	crypt32								= syscall.NewLazyDLL("crypt32.dll")
	advapi32							= syscall.NewLazyDLL("advapi32.dll")
    certGetCertificateContextProperty	= crypt32.NewProc("CertGetCertificateContextProperty")  
    cryptAcquireCertificatePrivateKey   = crypt32.NewProc("CryptAcquireCertificatePrivateKey")  
    certDuplicateCertificateContext		= crypt32.NewProc("CertDuplicateCertificateContext")
    cryptCreateHash						= advapi32.NewProc("CryptCreateHash")
    cryptSetHashParam					= advapi32.NewProc("CryptSetHashParam")
    cryptSignHashW						= advapi32.NewProc("CryptSignHashW")
    cryptReleaseContext					= advapi32.NewProc("CryptReleaseContext")
    
)

// CertDuplicateCertificateContext from wincrypt.h
//
func CertDuplicateCertificateContext(ctx *syscall.CertContext) *syscall.CertContext {
	
	ret, _, callErr := syscall.Syscall9(certDuplicateCertificateContext.Addr(),1,
		uintptr(unsafe.Pointer(ctx)),
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0)
	if callErr != 0 {
		tracerx.Printf("http: CertDuplicateCertificateContext %s", callErr)
	}

	return ( * syscall.CertContext)(unsafe.Pointer(ret))
}

// CertGetCertificateContextProperty from wincrypt.h
//
func CertGetCertificateContextProperty(ctx *syscall.CertContext,propid uint32,pdata uintptr,ldata * uint32) int {

	ret, _, callErr := syscall.Syscall9(certGetCertificateContextProperty.Addr(),4,
		uintptr(unsafe.Pointer(ctx)),
		uintptr(propid),
		uintptr(unsafe.Pointer(pdata)),
		uintptr(unsafe.Pointer(ldata)),
		0,
		0,
		0,
		0,
		0)
	if callErr != 0 {
		tracerx.Printf("http: CertGetCertificateContextProperty %s", callErr)
	}

	return int(ret)
}

// CryptAcquireCertificatePrivateKey from wincrypt.h
//
func CryptAcquireCertificatePrivateKey(ctx *syscall.CertContext,flags uint32,res uint32,phProv * uint32,pdwKeySpec * uint32,pbFree * uint32) int {

	ret, _, callErr := syscall.Syscall9(cryptAcquireCertificatePrivateKey.Addr(),6,
		uintptr(unsafe.Pointer(ctx)),
		uintptr(flags),
		uintptr(res),
		uintptr(unsafe.Pointer(phProv)),
		uintptr(unsafe.Pointer(pdwKeySpec)),
		uintptr(unsafe.Pointer(pbFree)),
		0,
		0,
		0)
	if callErr != 0 {
		tracerx.Printf("http: CryptAcquireCertificatePrivateKey %s", callErr)
	}

	return int(ret)

}

// CryptReleaseContext from wincrypt.h
//
func CryptReleaseContext(hProv uint32,flags uint32) int {
	
	ret, _, callErr := syscall.Syscall9(cryptReleaseContext.Addr(),2,
		uintptr(hProv),
		uintptr(flags),
		0,
		0,
		0,
		0,
		0,
		0,
		0)
	if callErr != 0 {
		tracerx.Printf("http: CryptCreateHash %s", callErr)
	}

	return int(ret)
}

// CryptCreateHash from wincrypt.h
//
func CryptCreateHash(hProv uint32,algid uint32,hKey uint32,flags uint32,phHash * uint32) int {
	
	ret, _, callErr := syscall.Syscall9(cryptCreateHash.Addr(),5,
		uintptr(hProv),
		uintptr(algid),
		uintptr(hKey),
		uintptr(flags),
		uintptr(unsafe.Pointer(phHash)),
		0,
		0,
		0,
		0)
	if callErr != 0 {
		tracerx.Printf("http: CryptCreateHash %s", callErr)
	}

	return int(ret)
}

// CryptSetHashParam from wincrypt.h
//
func CryptSetHashParam(hHash uint32,paramid uint32,value []byte,flags uint32) int {

	ret, _, callErr := syscall.Syscall9(cryptSetHashParam.Addr(),4,
		uintptr(hHash),
		uintptr(paramid),
		uintptr(unsafe.Pointer(&value[0])),
		uintptr(flags),
		0,
		0,
		0,
		0,
		0)
	if callErr != 0 {
		tracerx.Printf("http: CryptCreateHash %s", callErr)
	}

	return int(ret)
}

// CryptSignHashW from wincrypt.h
//
func CryptSignHashW(hHash uint32,dwKeySpec uint32,szDescription uint32,flags uint32,pbSignature []byte,pdwSigLen * uint32) int {

	ret, _, callErr := syscall.Syscall9(cryptSignHashW.Addr(),6,
		uintptr(hHash),
		uintptr(dwKeySpec),
		uintptr(szDescription),
		uintptr(flags),
		uintptr(unsafe.Pointer(&pbSignature[0])),
		uintptr(unsafe.Pointer(pdwSigLen)),
		0,
		0,
		0)
	if callErr != 0 {
		tracerx.Printf("http: CryptSignHashW %s", callErr)
	}

	return int(ret)
}

//
// WinKey holds a certContext pointing to a system certificate and private key
//
type WinKey struct {
	certCtx * syscall.CertContext
}

// WinKeyClose is the finalizer for the WinKey certificate context
//
func WinKeyClose(t * WinKey) {
	syscall.CertFreeCertificateContext(t.certCtx)
}

// WinKey::Public return the public key for this WinKey. We fake it by returning an empty key.
// The caller of this function is mostly interested int eh type of key, RSA in our case
//
func (t WinKey) Public() crypto.PublicKey {
	var r * rsa.PublicKey
	return r
}

//
// Map of hash algorithm to CAPI ALGID
//
var hashTypes = map[crypto.Hash]uint32{
	crypto.SHA1:	CALG_SHA1,
	crypto.SHA256:  CALG_SHA256,
	crypto.SHA512:  CALG_SHA512,
}

// WinKey::Sign will sign a clientVerify digest using the private key in the system
//	
func (t WinKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	var hProv uint32
	var dwKeySpec uint32
	var bFree uint32
	
	//
	// Load the provider for this private key
	//
	rc := CryptAcquireCertificatePrivateKey(t.certCtx,CRYPT_ACQUIRE_CACHE_FLAG,0,&hProv,&dwKeySpec,&bFree);
	if rc == 0 {
		return signature , errors.New("WinKeySign: CryptAcquireCertificatePrivateKey error")
	}
	
	//
	// Release the context depending on the bFree return value
	//
	if bFree != 0 {
		defer CryptReleaseContext(hProv,0);
	}

	//
	// Map hash function to CAPI hash ALGID
	//
	hashType, ok := hashTypes[opts.HashFunc()]
	if !ok {
		return signature, errors.New("WinKey::Sign: unsupported hash function")
	}

	var	hHash uint32

	//
	// Create a CAPI hash object
	//
	rc = CryptCreateHash(hProv,hashType,0,0,&hHash)
	if rc == 0 {
		return signature , errors.New("WinKeySign: CryptCreateHash error")
	}

	//
	// Set its value
	//
	rc = CryptSetHashParam(hHash,HP_HASHVAL,digest,0);
	if rc == 0 {
		return signature , errors.New("WinKeySign: CryptSetHashParam error")
	}

	
	//
	// Sign the hash using the private key. We pre-allocate a 1024 bytes buffer to
	// make sure we have space for the signature
	var dwLen uint32;
	dwLen = 256;
	buf := make([]byte, 1024)
	
	rc = CryptSignHashW(hHash,dwKeySpec,0,0,buf,&dwLen);
	if rc == 0 {

		CryptDestroyHash(hHash);

		return signature , errors.New("WinKeySign: CryptSetHashParam error")
	}

	// Free the handle
	CryptDestroyHash(hHash);

	// Trim the signature the retuned length
	buf = buf[:dwLen];
	
	//
	// Reverse the signature. CAPI has LSB
	//
	for i := len(buf)/2-1; i >= 0; i-- {
		opp := len(buf)-1-i
		buf[i], buf[opp] = buf[opp], buf[i]
	}
	
	return buf,nil
}

const CERT_SYSTEM_STORE_CURRENT_USER = 11;		// FIX:
const CERT_STORE_PROV_SYSTEM		= 1;
const CERT_STORE_OPEN_EXISTING_FLAG = 1;

// getClientCertsForHostFromSchannel return a list of platform certificates used to client authentication
//
func getClientCertsForHostFromSchannel(c *Client, host string) ([]tls.Certificate, error) {
	var wincerts [] tls.Certificate

	configSslcert , _ := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslcert")

	certParts := strings.SplitN(configSslcert,"\\",3)
	
	if len(certParts) != 3 {
		return nil , errors.New("Invalid sslcert format for schannel")
	}

	var store;

	switch certParts[0] {
	case "CurrentUser":
		store = CERT_SYSTEM_STORE_CURRENT_USER;
		break;
	default:
		return nil , errors.New("Invalid sslcert store for schannel")
	}

	store, err := syscall.CertOpenStore(CERT_STORE_PROV_SYSTEM,0,0,CERT_STORE_OPEN_EXISTING_FLAG | store,syscall.StringToUTF16Ptr(certParts[1]))
	if err != nil {
		return nil,err
	}

	//
	// Open the MY system certificate store
	//
	store, err := syscall.CertOpenSystemStore(0, syscall.StringToUTF16Ptr("MY"))
	if err != nil {
		return nil,err
	}
	defer syscall.CertCloseStore(store, 0)

	//
	// Enum all certificates in the store
	//
	var cert *syscall.CertContext
	for {
		cert, err = syscall.CertEnumCertificatesInStore(store, cert)
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				if errno == CRYPT_E_NOT_FOUND {
					break
				}
			}
			return nil,err
		}

		if cert == nil {
			break
		}

		//
		// Check if we have a private key for this certificate
		// by reading the CERT_KEY_PROV_INFO_PROP_ID property
		//
		var len uint32
		rc := CertGetCertificateContextProperty(cert, CERT_KEY_PROV_INFO_PROP_ID,0,&len);
		if rc == 0 {
			tracerx.Printf("http: Ignore windows client cert, no private key info")
			continue
		}

		var wincert tls.Certificate;

		//
		// Copy the centificate data from the store buffer to a new one since the buffer will be overwritten next iteration
		//
		buf := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:]
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)
		
		// Add the certificate to the wincert instance	
		wincert.Certificate = append(wincert.Certificate,buf2)
		
		// Create a new WinKey instance
		winkey := new(WinKey)
		
		// Copy the certificate context since it will be overwritten in the next iteration
		winkey.certCtx = CertDuplicateCertificateContext(cert);
		
		// Tell the runtime to free the certificate context at GC
		runtime.SetFinalizer(winkey,WinKeyClose)
		
		// Setup the private key
		wincert.PrivateKey = winkey
		
		// Append to the list of platform certificates
		wincerts = append(wincerts,wincert)
	}			
	return wincerts,nil
}
