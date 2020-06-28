package cert_providers

import "bytes"

type StaticGeckoCertProvider struct {
	key bytes.Buffer
	cert bytes.Buffer
}

func NewStaticGeckoCertProvider(key bytes.Buffer, cert bytes.Buffer) *StaticGeckoCertProvider {
	return &StaticGeckoCertProvider{key: key, cert: cert}
}

func (s StaticGeckoCertProvider) GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error) {
	return s.cert, s.key, nil
}

