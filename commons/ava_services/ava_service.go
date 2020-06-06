package ava_services

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

type AvaService interface {
	services.Service

	GetStakingSocket() services.ServiceSocket

}
