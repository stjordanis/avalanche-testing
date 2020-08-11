package cert_providers

import "bytes"

/*
Interface representing a cert provider for a Gecko service (used in the duplicate node ID test, which requires that
	multiple Gecko services start with the same cert)
 */
type GeckoCertProvider interface {
	/*
	Generates a cert and accompanying private key

	Returns:
		certPemBytes: The bytes of the generated cert
		keyPemBytes: The bytes of the private key generated with the cert
	 */
	GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error)
}
