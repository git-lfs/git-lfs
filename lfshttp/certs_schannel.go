//go:build windows
// +build windows

package lfshttp

import (
	"bytes"
	"crypto"
	"io"
	"crypto/tls"
	"crypto/rsa"
	"crypto/x509"
	"unsafe"
	"errors"
	"github.com/rubyist/tracerx"
	"fmt"
	"strings"
	"encoding/hex"
	"runtime"
	
	"golang.org/x/sys/windows"
)

const (

	nCryptSilentFlag   = 0x00000040 // ncrypt.h NCRYPT_SILENT_FLAG
	bCryptPadPss       = 0x00000008 // bcrypt.h BCRYPT_PAD_PSS
	supportedAlgorithm = tls.PSSWithSHA256

)

type DataBlob struct {
	Size uint32
	Data *byte
}
type CryptHashBlob DataBlob

var (
	nCrypt         = windows.MustLoadDLL("ncrypt.dll")
	nCryptSignHash = nCrypt.MustFindProc("NCryptSignHash")
)

//
// WinKey holds a certContext pointing to a system certificate and private key
//
type WinKey struct {
	certCtx			 *windows.CertContext
	x509Cert     *x509.Certificate
}

// WinKeyClose is the finalizer for the WinKey certificate context
//
func WinKeyClose(t * WinKey) {
	windows.CertFreeCertificateContext(t.certCtx)
}

// WinKey::Public return the public key for this WinKey. We fake it by returning an empty key.
// The caller of this function is mostly interested int eh type of key, RSA in our case
//
func (t WinKey) Public() crypto.PublicKey {
  tracerx.PrintfKey("SCHANNEL","schannel: crypto.PublicKey")
  
  if(t.x509Cert.PublicKey!=0) {
		return t.x509Cert.PublicKey
	} 
	
	return nil
}


func (k *WinKey) Sign(_ io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	tracerx.PrintfKey("SCHANNEL","crypto.Signer.Sign with key type %T, opts type %T, hash %s\n", k.Public(), opts, opts.HashFunc().String())

	// Get private key
	var (
		privateKey                  windows.Handle
		pdwKeySpec                  uint32
		pfCallerFreeProvOrNCryptKey bool
	)
	err = windows.CryptAcquireCertificatePrivateKey(
		k.certCtx,
		windows.CRYPT_ACQUIRE_CACHE_FLAG|windows.CRYPT_ACQUIRE_SILENT_FLAG|windows.CRYPT_ACQUIRE_ONLY_NCRYPT_KEY_FLAG,
		nil,
		&privateKey,
		&pdwKeySpec,
		&pfCallerFreeProvOrNCryptKey,
	)
	if err != nil {
		return nil, err
	}

	// We always use RSA-PSS padding
	flags := nCryptSilentFlag | bCryptPadPss
	pPaddingInfo, err := getRsaPssPadding(opts)
	if err != nil {
		return nil, err
	}

	// Sign the digest
	// The first call to NCryptSignHash retrieves the size of the signature
	var size uint32
	success, _, _ := nCryptSignHash.Call(
		uintptr(privateKey),
		uintptr(pPaddingInfo),
		uintptr(unsafe.Pointer(&digest[0])),
		uintptr(len(digest)),
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(&size)),
		uintptr(flags),
	)
	if success != 0 {
		return nil, fmt.Errorf("NCryptSignHash: failed to get signature length: %#x", success)
	}

	// The second call to NCryptSignHash retrieves the signature
	signature = make([]byte, size)
	success, _, _ = nCryptSignHash.Call(
		uintptr(privateKey),
		uintptr(pPaddingInfo),
		uintptr(unsafe.Pointer(&digest[0])),
		uintptr(len(digest)),
		uintptr(unsafe.Pointer(&signature[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&size)),
		uintptr(flags),
	)
	if success != 0 {
		return nil, fmt.Errorf("NCryptSignHash: failed to generate signature: %#x", success)
	}
	return signature, nil
}

func getRsaPssPadding(opts crypto.SignerOpts) (unsafe.Pointer, error) {
	pssOpts, ok := opts.(*rsa.PSSOptions)
	if !ok || pssOpts.Hash != crypto.SHA256 {
		return nil, fmt.Errorf("unsupported hash function %s", opts.HashFunc().String())
	}
	if pssOpts.SaltLength != rsa.PSSSaltLengthEqualsHash {
		return nil, fmt.Errorf("unsupported salt length %d", pssOpts.SaltLength)
	}
	sha256, _ := windows.UTF16PtrFromString("SHA256")
	// Create BCRYPT_PSS_PADDING_INFO structure:
	// typedef struct _BCRYPT_PSS_PADDING_INFO {
	// 	LPCWSTR pszAlgId;
	// 	ULONG   cbSalt;
	// } BCRYPT_PSS_PADDING_INFO;
	return unsafe.Pointer(
		&struct {
			pszAlgId *uint16
			cbSalt   uint32
		}{
			pszAlgId: sha256,
			cbSalt:   uint32(pssOpts.HashFunc().Size()),
		},
	), nil
}

//
// getClientCertForHostFromSchannel return a platform certificates used to client authentication based on sslcert "string"
//
func getClientCertForHostFromSchannel(c *Client, host string) (*tls.Certificate, error) {
	//var wincerts [] tls.Certificate

	configSslcert , _ := c.uc.Get("http", fmt.Sprintf("https://%v/", host), "sslcert")

	certParts := strings.SplitN(configSslcert,"\\",3)
	
	tracerx.PrintfKey("SCHANNEL","schannel: Should get certiifcate %s",certParts[2])
	
	if len(certParts) != 3 {
	  return nil , errors.New("Invalid sslcert format for schannel")
	}

	store, err := windows.CertOpenSystemStore(0,windows.StringToUTF16Ptr(certParts[1]))
	if err != nil {
		tracerx.PrintfKey("SCHANNEL","schannel: Failed to open cert store %s",certParts[1])
		return nil,err
	}
	defer windows.CertCloseStore(store, 0)

	var find_type uint32

	find_type = windows.CERT_FIND_HASH
	var hash CryptHashBlob
	
	data, err := hex.DecodeString(certParts[2]) 
	if err != nil {
			tracerx.PrintfKey("SCHANNEL","schannel: Error decoding %s",certParts[2])
    	return nil,err
	}
	
	
	len := hex.DecodedLen(len(certParts[2]))
		
	hash.Size = uint32(len)
	hash.Data = &data[0]
	
	tracerx.PrintfKey("SCHANNEL","schannel: Search for certificate")
	
	var rv *windows.CertContext
	var cert *windows.CertContext
	
	cert, err = windows.CertFindCertificateInStore(store, windows.X509_ASN_ENCODING | windows.PKCS_7_ASN_ENCODING,0, find_type,unsafe.Pointer(&hash), rv);	
	
	var wincert tls.Certificate;
			
	// Copy the certificate data so that we have our own copy outside the windows context
	encodedCert := unsafe.Slice(cert.EncodedCert, cert.Length)
	buf := bytes.Clone(encodedCert)
	foundCert, err := x509.ParseCertificate(buf)
	if err != nil {
		return nil, err
	}
	
	// Add the certificate to the wincert instance	
	wincert.Certificate = [][]byte{foundCert.Raw}
	
	// Create a new WinKey instance
	winkey := new(WinKey)
	
	// Copy the certificate context since it will be overwritten in the next iteration
	winkey.certCtx = windows.CertDuplicateCertificateContext(cert);
	
	// Save the certificate data
	winkey.x509Cert = foundCert

	// Tell the runtime to free the certificate context at GC
	runtime.SetFinalizer(winkey,WinKeyClose)
	
	// Setup the private key
	wincert.PrivateKey = winkey
	
	wincert.SupportedSignatureAlgorithms = []tls.SignatureScheme{supportedAlgorithm} 
		
	tracerx.PrintfKey("SCHANNEL","schannel: Found certificate")
	
	return &wincert,err
}
