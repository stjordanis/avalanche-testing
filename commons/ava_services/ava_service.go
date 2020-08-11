package ava_services

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

/*
An implementation of Kurtosis' generic services.Service interface that represents the minimum interface an Ava node has
 */
type AvaService interface {
	services.Service

	/*
	Gets the staking socket of an Ava node (which all nodes must have if they're part of an Ava network)
	 */
	GetStakingSocket() ServiceSocket

}
