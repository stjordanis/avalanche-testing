package cert_providers

import "bytes"

/*
An implementation of GeckoCertProvider that provides the same cert every time
 */
type StaticGeckoCertProvider struct {
	key bytes.Buffer
	cert bytes.Buffer
}

/*
Creates an instance of StaticGeckoCertProvider using the given key and cert

Args:
	key: The private key that the StaticGeckoCertProvider will return on every call to GetCertAndKey
	cert: The cert that will be returned on every call to GetCertAndKey
 */
func NewStaticGeckoCertProvider(key bytes.Buffer, cert bytes.Buffer) *StaticGeckoCertProvider {
	return &StaticGeckoCertProvider{key: key, cert: cert}
}

/*
Return the same cert that was configured at time of construction
 */
func (s StaticGeckoCertProvider) GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error) {
	return s.cert, s.key, nil
}

