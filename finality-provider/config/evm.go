package config

import (
	"fmt"
	"net/url"
)

const (
	defaultEVMRPCAddr = "http://127.0.0.1:8545"
)

type EVMConfig struct {
	L1RPCAddr        string `long:"rpcl1-address" description:"address of the L1 RPC server to connect to"`
	ConsumerRPCAddr  string `long:"consumer-chain-address" description:"address of the consumer chain RPC server to connect to"`
	EOTSVerifierAddr string `long:"EOTSVerifier-address" description:"address of the EOTSVerifier smart contract"`
	FPOracleAddr     string `long:"finality-provider-address" description:"address of the finality provider smart contract"`
}

func DefaultEVMConfig() EVMConfig {
	return EVMConfig{
		L1RPCAddr: defaultEVMRPCAddr,
	}
}

func (cfg *EVMConfig) Validate() error {
	if _, err := url.Parse(cfg.L1RPCAddr); err != nil {
		return fmt.Errorf("rpcl1-address is not correctly formatted: %w", err)
	}
	if _, err := url.Parse(cfg.ConsumerRPCAddr); err != nil {
		return fmt.Errorf("consumer-chain-address is not correctly formatted: %w", err)
	}
	return nil
}
