package config

import (
	"fmt"
	"net/url"
)

const (
	defaultEVMRPCAddr = "http://127.0.0.1:8545"
)

type EVMConfig struct {
	RPCL1Addr          string `long:"rpcl1-address" description:"address of the L1 RPC server to connect to"`
	RPCL2Addr          string `long:"rpcl2-address" description:"address of the L2 RPC server to connect to"`
	L2OutputOracleAddr string `long:"l2outputoracle-address" description:"address of the L2OutputOracle smart contract"`
	BSAddr             string `long:"bitcoinstacking-address" description:"address of the BitcoinStaking smart contract"`
}

func DefaultEVMConfig() EVMConfig {
	return EVMConfig{
		RPCL2Addr: defaultEVMRPCAddr,
	}
}

func (cfg *EVMConfig) Validate() error {
	if _, err := url.Parse(cfg.RPCL1Addr); err != nil {
		return fmt.Errorf("rpcl1-addr is not correctly formatted: %w", err)
	}
	if _, err := url.Parse(cfg.RPCL2Addr); err != nil {
		return fmt.Errorf("rpcl2-addr is not correctly formatted: %w", err)
	}
	return nil
}
