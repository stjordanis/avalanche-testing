package ava_services

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math/big"
	mathrand "math/rand"
	"os"
	"strings"
	"time"
)

const (
	httpPort             = 9650
	stakingPort          = 9651
	stakingTlsCertFileId = "staking-tls-cert"
	stakingTlsKeyFileId  = "staking-tls-key"
	maxCerts             = 4000
	certificatePreamble  = "CERTIFICATE"
	privateKeyPreamble   = "RSA PRIVATE KEY"

	testVolumeMountpoint = "/shared"
)

// ========= Loglevel Enum ========================
type geckoLogLevel string
const (
	LOG_LEVEL_VERBOSE geckoLogLevel = "verbo"
	LOG_LEVEL_DEBUG   geckoLogLevel = "debug"
	LOG_LEVEL_INFO    geckoLogLevel = "info"
)

// ========= Initializer Core ========================
type GeckoServiceInitializerCore struct {
	snowSampleSize    int
	snowQuorumSize    int
	stakingTlsEnabled bool
	bootstrapperNodeIds []string
	logLevel          geckoLogLevel
}

/*
Creates a new Gecko service initializer core with the following parameters:

Args:
	snowSampleSize: Sample size for Snow consensus protocol
	snowQuroumSize: Quorum size for Snow consensus protocol
	stakingTlsEnabled: Whether this node will use staking & TLS
	bootstrapperNodeIds: The node IDs of the bootstrapper nodes that this node will connect to. While this *seems* unintuitive
		why this would be required, it's because Gecko doesn't actually use certs. So, to prevent against man-in-the-middle attacks,
		the user is required to manually specify the node IDs of the nodese it's connecting to.
	logLevel: The loglevel that the Gecko node should output at.

Returns:
	An intializer core for creating Gecko nodes with the specified parameers.
 */
func NewGeckoServiceInitializerCore(
			snowSampleSize int,
			snowQuorumSize int,
			stakingTlsEnabled bool,
			bootstrapperNodeIds []string,
			logLevel geckoLogLevel) *GeckoServiceInitializerCore {
	// Defensive copy
	bootstrapperIdsCopy := make([]string, 0, len(bootstrapperNodeIds))
	copy(bootstrapperIdsCopy, bootstrapperNodeIds)
	return &GeckoServiceInitializerCore{
		snowSampleSize:    snowSampleSize,
		snowQuorumSize:    snowQuorumSize,
		stakingTlsEnabled: stakingTlsEnabled,
		bootstrapperNodeIds: bootstrapperIdsCopy,
		logLevel:          logLevel,
	}
}

func (core GeckoServiceInitializerCore) GetUsedPorts() map[int]bool {
	return map[int]bool{
		httpPort:    true,
		stakingPort: true,
	}
}

func (core GeckoServiceInitializerCore) GetFilesToMount() map[string]bool {
	if core.stakingTlsEnabled {
		return map[string]bool{
			stakingTlsCertFileId: true,
			stakingTlsKeyFileId:  true,
		}
	}
	return make(map[string]bool)
}

func (core GeckoServiceInitializerCore) InitializeMountedFiles(osFiles map[string]*os.File, dependencies []services.Service) (err error) {
	certFilePointer := osFiles[stakingTlsCertFileId]
	keyFilePointer := osFiles[stakingTlsKeyFileId]
	if len(dependencies) == 0 {
		certFilePointer.WriteString(STAKER_1_CERT)
		keyFilePointer.WriteString(STAKER_1_PRIVATE_KEY)
	} else {
		rootCA := getRootCA()
		serviceCert := getServiceCert()
		certPEM, keyPEM, err := getServiceCertAndKeyFiles(serviceCert, rootCA)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to write files.")
		}
		certFilePointer.Write(certPEM.Bytes())
		keyFilePointer.Write(keyPEM.Bytes())
	}
	return nil
}

func (core GeckoServiceInitializerCore) GetStartCommand(mountedFileFilepaths map[string]string, publicIpAddr string, dependencies []services.Service) ([]string, error) {
	publicIpFlag := fmt.Sprintf("--public-ip=%s", publicIpAddr)
	commandList := []string{
		"/gecko/build/ava",
		publicIpFlag,
		"--network-id=local",
		fmt.Sprintf("--http-port=%d", httpPort),
		fmt.Sprintf("--staking-port=%d", stakingPort),
		fmt.Sprintf("--log-level=%s", core.logLevel),
		fmt.Sprintf("--snow-sample-size=%d", core.snowSampleSize),
		fmt.Sprintf("--snow-quorum-size=%d", core.snowQuorumSize),
		fmt.Sprintf("--staking-tls-enabled=%v", core.stakingTlsEnabled),
	}

	if core.stakingTlsEnabled {
		certFilepath, found := mountedFileFilepaths[stakingTlsCertFileId]
		if !found {
			return nil, stacktrace.NewError("Could not find file key '%v' in the mounted filepaths map; this is likely a code bug", stakingTlsCertFileId)
		}
		keyFilepath, found := mountedFileFilepaths[stakingTlsKeyFileId]
		if !found {
			return nil, stacktrace.NewError("Could not find file key '%v' in the mounted filepaths map; this is likely a code bug", stakingTlsKeyFileId)
		}
		commandList = append(commandList, fmt.Sprintf("--staking-tls-cert-file=%s", certFilepath))
		commandList = append(commandList, fmt.Sprintf("--staking-tls-key-file=%s", keyFilepath))
	}


	// If bootstrap nodes are down then Gecko will wait until they are, so we don't actually need to busy-loop making
	// requests to the nodes
	if dependencies != nil && len(dependencies) > 0 {
		avaDependencies := make([]AvaService, 0, len(dependencies))
		for _, service := range dependencies {
			avaDependencies = append(avaDependencies, service.(AvaService))
		}

		socketStrs := make([]string, 0, len(avaDependencies))
		for _, service := range avaDependencies {
			socket := service.GetStakingSocket()
			socketStrs = append(socketStrs, fmt.Sprintf("%s:%d", socket.GetIpAddr(), socket.GetPort().Int()))
			if core.stakingTlsEnabled {
				bootstrapperIdsList := strings.Join(core.bootstrapperNodeIds, ",")

				// NOTE: This seems weird, BUT there's a reason for it: Gecko doesn't use certs, and instead relies on
				//  the user explicitly passing in the node ID of the bootstrapper it wants. This prevents man-in-the-middle
				//  attacks, just like using a cert would. Us hardcoding this bootstrapper ID here is the equivalent
				//  of a user knowing the node ID in advance, which provides the same level of protection.
				commandList = append(commandList, "--bootstrap-ids=" + bootstrapperIdsList)
				// We currently have one cert -> ID mapping so break the for loop here.
				break
			}
		}
		joinedSockets := strings.Join(socketStrs, ",")
		commandList = append(commandList, "--bootstrap-ips=" + joinedSockets)
	}
	logrus.Debugf("Command list: %+v", commandList)
	return commandList, nil
}

func (core GeckoServiceInitializerCore) GetServiceFromIp(ipAddr string) services.Service {
	return GeckoService{ipAddr: ipAddr}
}

func (core GeckoServiceInitializerCore) GetTestVolumeMountpoint() string {
	return testVolumeMountpoint
}

func getServiceCertAndKeyFiles(cert *x509.Certificate, ca *x509.Certificate) (certFile *bytes.Buffer, keyFile *bytes.Buffer, err error) {
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to generate random private key.")
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &(certPrivKey.PublicKey), certPrivKey)
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
		Type: privateKeyPreamble,
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	return certPEM, certPrivKeyPEM, nil
}

func getRootCA() *x509.Certificate {
	ca := &x509.Certificate{
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
	return ca
}

func getServiceCert() *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(int64(mathrand.Intn(maxCerts))),
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
