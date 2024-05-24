package clientcontroller

import (
	"context"
	"fmt"
	"math/big"

	fpcfg "github.com/babylonchain/finality-provider/finality-provider/config"
	"github.com/babylonchain/finality-provider/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// TODO: rename the file name, class name and etc
// This is not a simple EVM chain. It's a OP Stack L2 chain, which has many
// implications. So we should rename to sth like e.g. OPStackL2Consumer
// This helps distinguish from pure EVM sidechains e.g. Binance Chain
var _ ConsumerController = &EVMConsumerController{}

type EVMConsumerController struct {
	l1Client *ethclient.Client
	l2Client *ethclient.Client
	cfg      *fpcfg.EVMConfig
	logger   *zap.Logger
}

func NewEVMConsumerController(
	evmCfg *fpcfg.EVMConfig,
	logger *zap.Logger,
) (*EVMConsumerController, error) {
	if err := evmCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config for EVM RPC client: %w", err)
	}
	l1Client, err := ethclient.Dial(evmCfg.RPCL1Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the L1 RPC server %s: %w", evmCfg.RPCL1Addr, err)
	}
	l2Client, err := ethclient.Dial(evmCfg.RPCL2Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the L2 RPC server %s: %w", evmCfg.RPCL2Addr, err)
	}
	return &EVMConsumerController{
		l1Client,
		l2Client,
		evmCfg,
		logger,
	}, nil
}

// SubmitFinalitySig submits the finality signature
func (ec *EVMConsumerController) SubmitFinalitySig(fpPk *btcec.PublicKey, blockHeight uint64, blockHash []byte, sig *btcec.ModNScalar) (*types.TxResponse, error) {

	return &types.TxResponse{TxHash: "", Events: nil}, nil
}

// SubmitBatchFinalitySigs submits a batch of finality signatures to Babylon
func (ec *EVMConsumerController) SubmitBatchFinalitySigs(fpPk *btcec.PublicKey, blocks []*types.BlockInfo, sigs []*btcec.ModNScalar) (*types.TxResponse, error) {
	if len(blocks) != len(sigs) {
		return nil, fmt.Errorf("the number of blocks %v should match the number of finality signatures %v", len(blocks), len(sigs))
	}

	return &types.TxResponse{TxHash: "", Events: nil}, nil
}

// QueryFinalityProviderVotingPower queries the voting power of the finality provider at a given height
func (ec *EVMConsumerController) QueryFinalityProviderVotingPower(fpPk *btcec.PublicKey, blockHeight uint64) (uint64, error) {
	/* TODO: implement

	latest_committed_l2_height = read `latestBlockNumber()` from the L1 L2OutputOracle contract and return the result

	if blockHeight > latest_committed_l2_height:

		query the VP from the L1 oracle contract using "latest" as the block tag

	else:

		1. query the L1 event `emit OutputProposed(_outputRoot, nextOutputIndex(), _l2BlockNumber, block.timestamp, block.number);`
		  to find the first event where the `_l2BlockNumber` >= blockHeight
		2. get the block.number from the event
		3. query the VP from the L1 oracle contract using `block.number` as the block tag

	*/

	return 0, nil
}

func (ec *EVMConsumerController) QueryLatestFinalizedBlock() (*types.BlockInfo, error) {

	lastNumber, err := ec.queryLatestFinalizedNumber()
	if err != nil {
		return nil, fmt.Errorf("can't get latest finalized block number:%s", err)
	}

	block, err := ec.QueryBlock(lastNumber)
	if err != nil {
		return nil, fmt.Errorf("can't get latest finalized block:%s", err)
	}

	return block, nil
}

func (ec *EVMConsumerController) QueryBlocks(startHeight, endHeight, limit uint64) ([]*types.BlockInfo, error) {

	if endHeight < startHeight {
		return nil, fmt.Errorf("the startHeight %v should not be higher than the endHeight %v", startHeight, endHeight)
	}
	count := endHeight - startHeight
	if count > limit {
		count = limit
	}

	var blocks []*types.BlockInfo

	for i := 0; i < int(count); i++ {

		block, err := ec.QueryBlock(startHeight)
		if err != nil {
			return nil, fmt.Errorf("failed to get start block:%s", err)
		}
		blocks = append(blocks, block)
		startHeight++

	}

	return blocks, nil
}

func (ec *EVMConsumerController) QueryBlock(height uint64) (*types.BlockInfo, error) {

	number := new(big.Int).SetUint64(height)

	header, err := ec.l2Client.HeaderByNumber(context.Background(), number)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block:%s", err)
	}

	blockinfo := &types.BlockInfo{
		Height: header.Number.Uint64(),
		Hash:   header.Hash().Bytes(),
	}

	return blockinfo, nil
}

func (ec *EVMConsumerController) QueryIsBlockFinalized(height uint64) (bool, error) {

	lastNumber, err := ec.queryLatestFinalizedNumber()
	if err != nil {
		return false, fmt.Errorf("can't get latest finalized block:%s", err)
	}

	var finalized bool = false

	if height <= lastNumber {
		finalized = true
	}

	return finalized, nil
}

func (ec *EVMConsumerController) QueryActivatedHeight() (uint64, error) {
	/* TODO: implement

		oracle_event = query the event in the L1 oracle contract where the FP's voting power is firstly set

		l1_activated_height = get the L1 block number from the `oracle_event`

	  output_event = query the L1 event `emit OutputProposed(_outputRoot, nextOutputIndex(), _l2BlockNumber, block.timestamp, block.number);`
				to find the first event where the `block.number` >= l1_activated_height

		if output_event == nil:

				read `nextBlockNumber()` from the L1 L2OutputOracle contract and return the result

		else:

				return output_event._l2BlockNumber

	*/

	return 0, nil
}

func (ec *EVMConsumerController) QueryLatestBlockHeight() (uint64, error) {

	header, err := ec.l2Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block:%s", err)
	}

	return header.Number.Uint64(), nil
}

func (ec *EVMConsumerController) Close() error {

	ec.l1Client.Close()
	ec.l2Client.Close()

	return nil
}

func (ec *EVMConsumerController) queryLatestFinalizedNumber() (uint64, error) {

	output, err := bindings.NewL2OutputOracle(common.HexToAddress(ec.cfg.L2OutputOracleAddr), ec.l1Client)
	if err != nil {
		return 0, fmt.Errorf("failed to instantiate L2OutputOracle contract:%s ", err)
	}

	lastNumber, err := output.LatestBlockNumber(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest finalize block number:%s ", err)
	}

	return lastNumber.Uint64(), err
}
