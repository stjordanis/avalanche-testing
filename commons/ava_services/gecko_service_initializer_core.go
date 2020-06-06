package ava_services

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	httpPort = 9650
	stakingPort = 9651
	stakingTlsCertPath = "node.crt"
	stakingTlsKeyPath = "node.key"
)

// ================= Service ==================================

type GeckoService struct {
	ipAddr string
}

func (g GeckoService) GetStakingSocket() services.ServiceSocket {
	stakingPort, err := nat.NewPort("tcp", strconv.Itoa(stakingPort))
	if err != nil {
		// Realllllly don't think we should deal with propagating this one.... it means the user mistyped an integer
		panic(err)
	}
	return *services.NewServiceSocket(g.ipAddr, stakingPort)
}

func (g GeckoService) GetJsonRpcSocket() services.ServiceSocket {
	httpPort, err := nat.NewPort("tcp", strconv.Itoa(httpPort))
	if err != nil {
		panic(err)
	}
	return *services.NewServiceSocket(g.ipAddr, httpPort)
}


// ================ Initializer Core =============================
type geckoLogLevel string
const (
	LOG_LEVEL_VERBOSE geckoLogLevel = "verbo"
	LOG_LEVEL_DEBUG   geckoLogLevel = "debug"
	LOG_LEVEL_INFO    geckoLogLevel = "info"
)

type GeckoServiceInitializerCore struct {
	snowSampleSize    int
	snowQuorumSize    int
	stakingTlsEnabled bool
	logLevel          geckoLogLevel
}

func NewGeckoServiceInitializerCore(
	snowSampleSize int,
	snowQuorumSize int,
	stakingTlsEnabled bool,
	logLevel geckoLogLevel) *GeckoServiceInitializerCore {
	return &GeckoServiceInitializerCore{
		snowSampleSize:    snowSampleSize,
		snowQuorumSize:    snowQuorumSize,
		stakingTlsEnabled: stakingTlsEnabled,
		logLevel:          logLevel,
	}
}

func (g GeckoServiceInitializerCore) GetUsedPorts() map[int]bool {
	return map[int]bool{
		httpPort:    true,
		stakingPort: true,
	}
}


func (g GeckoServiceInitializerCore) GetFilepathsToMount() map[string]bool {
	if g.stakingTlsEnabled {
		return map[string]bool{
			stakingTlsCertPath: true,
			stakingTlsKeyPath: true,
		}
	}
	return make(map[string]bool)
}

func (g GeckoServiceInitializerCore) InitializeMountedFiles(osFiles map[string]*os.File, dependencies []services.Service) (err error) {
	certFilePointer := osFiles[stakingTlsCertPath]
	keyFilePointer := osFiles[stakingTlsKeyPath]
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

func (g  GeckoServiceInitializerCore) GetStartCommand(publicIpAddr string, serviceDataDir string, dependencies []services.Service) ([]string, error) {
	publicIpFlag := fmt.Sprintf("--public-ip=%s", publicIpAddr)
	commandList := []string{
		"/gecko/build/ava",
		publicIpFlag,
		"--network-id=local",
		fmt.Sprintf("--http-port=%d", httpPort),
		fmt.Sprintf("--staking-port=%d", stakingPort),
		fmt.Sprintf("--log-level=%s", g.logLevel),
		fmt.Sprintf("--snow-sample-size=%d", g.snowSampleSize),
		fmt.Sprintf("--snow-quorum-size=%d", g.snowQuorumSize),
		fmt.Sprintf("--staking-tls-enabled=%v", g.stakingTlsEnabled),
	}
	if g.stakingTlsEnabled {
		commandList = append(commandList, fmt.Sprintf("--staking-tls-cert-file=%s", serviceDataDir + stakingTlsCertPath))
		commandList = append(commandList, fmt.Sprintf("--staking-tls-key-file=%s", serviceDataDir + stakingTlsKeyPath))
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
			break
		}
		joinedSockets := strings.Join(socketStrs, ",")
		commandList = append(commandList, "--bootstrap-ips=" + joinedSockets)
		if g.stakingTlsEnabled {
			commandList = append(commandList, "--bootstrap-ids="+"7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg")
		}
	}
	logrus.Debugf("Command list: %+v", commandList)
	return commandList, nil
}

func (g GeckoServiceInitializerCore) GetServiceFromIp(ipAddr string) services.Service {
	return GeckoService{ipAddr: ipAddr}
}


func getServiceCertAndKeyFiles(cert *x509.Certificate, ca *x509.Certificate) (certFile *bytes.Buffer, keyFile *bytes.Buffer, err error) {
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "")
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, certPrivKey)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "")
	}
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
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
		SerialNumber: big.NewInt(1993),
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
