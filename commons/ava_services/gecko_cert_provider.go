package ava_services

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/palantir/stacktrace"
	"math/big"
	mathrand "math/rand"
	"time"
)

const (
	certificatePreamble  = "CERTIFICATE"
	privateKeyPreamble   = "RSA PRIVATE KEY"
)

var rootCert = x509.Certificate{
	SerialNumber: big.NewInt(2020),
	Subject: pkix.Name{
		Organization:  []string{"Kurtosis Technologies"},
		Country:       []string{"USA"},
		Province:      []string{""},
		Locality:      []string{""},
		StreetAddress: []string{""},
		PostalCode:    []string{""},
	},
	NotBefore:             time.Now(),
	NotAfter:              time.Now().AddDate(10, 0, 0),
	IsCA:                  true,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	BasicConstraintsValid: true,
}

/*
A provider for Gecko service certs, with all certs signed by the same root CA
 */
type GeckoCertProvider struct {
	nextSerialNumber int64
	varyCerts bool
}

/*
Creates a new cert provider that can optionally return either the same cert every time, or different ones

Args:
	varyCerts: Whether to produce a different cert on each call to GetCertAndKey
 */
func NewGeckoCertProvider(varyCerts bool) *GeckoCertProvider {
	return &GeckoCertProvider{
		nextSerialNumber: mathrand.Int63(),
		varyCerts: varyCerts,
	}
}

func (r *GeckoCertProvider) GetCertAndKey() (certPemBytes *bytes.Buffer, keyPemBytes *bytes.Buffer, err error) {
	serialNum := r.nextSerialNumber
	if (r.varyCerts) {
		r.nextSerialNumber = mathrand.Int63()
	}
	serviceCert := getServiceCert(serialNum)

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to generate random private key.")
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, serviceCert, &rootCert, &(certPrivKey.PublicKey), certPrivKey)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to sign service cert with cert authority.")
	}
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  certificatePreamble,
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  privateKeyPreamble,
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	return certPEM, certPrivKeyPEM, nil
}

// ================= Helper functions ===================
func getServiceCert(serialNumber int64) *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(serialNumber),
		Subject: pkix.Name{
			Organization:  []string{"Kurtosis Technologies"},
			Country:       []string{"USA"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	return cert
}
