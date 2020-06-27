package cert_providers

import "bytes"

type GeckoCertProvider interface {
	GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error)
}
