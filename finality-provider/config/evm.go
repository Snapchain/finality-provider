package config

import (
	"fmt"
	"net/url"
)

const (
	defaultEVMRPCAddr = "http://127.0.0.1:8545"
)

type EVMConfig struct {
	RPCL1Addr                      string `long:"rpc-address" description:"address of the L2rpc server to connect to"`
	RPCL2Addr                      string `long:"rpc-address" description:"address of the L1rpc server to connect to"`
	L2OutputOracleContractAddress  string `long:"sol-address" description:"address of the L2output smart contract"`
	BitcoinStackingContractAddress string `long:"sol-address" description:"address of the Bitcoinstaking smart contract"`
}

func DefaultEVMConfig() EVMConfig {
	return EVMConfig{
		RPCL2Addr: defaultEVMRPCAddr,
	}
}

func (cfg *EVMConfig) Validate() error {
	if _, err := url.Parse(cfg.RPCL2Addr); err != nil {
		return fmt.Errorf("rpcl2-addr is not correctly formatted: %w", err)
	}
	return nil
}
