package clientcontroller

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	finalitytypes "github.com/babylonchain/babylon/x/finality/types"
	fpcfg "github.com/babylonchain/finality-provider/finality-provider/config"
	"github.com/babylonchain/finality-provider/types"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/btcsuite/btcd/btcec/v2"
	"go.uber.org/zap"
)

// TODO: rename the file name, class name and etc
// This is not a simple EVM chain. It's a OP Stack L2 chain, which has many
// implications. So we should rename to sth like e.g. OPStackL2Consumer
// This helps distinguish from pure EVM sidechains e.g. Binance Chain
var _ ConsumerController = &EVMConsumerController{}

type EVMConsumerController struct {
	evmClient *rpc.Client
	cfg       *fpcfg.EVMConfig
	logger    *zap.Logger
}

func NewEVMConsumerController(
	evmCfg *fpcfg.EVMConfig,
	logger *zap.Logger,
) (*EVMConsumerController, error) {
	if err := evmCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config for EVM RPC client: %w", err)
	}
	ec, err := rpc.Dial(evmCfg.RPCL2Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the EVM RPC server %s: %w", evmCfg.RPCL2Addr, err)
	}
	return &EVMConsumerController{
		ec,
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

func (ec *EVMConsumerController) QueryLatestFinalizedBlocks(count uint64) ([]*types.BlockInfo, error) {

	lastnumber, err := ec.GetLatestFinalizedNumber()
	if err != nil {
		return nil, fmt.Errorf("can't get latest finalized block:%s", err)

	}

	type Block struct {
		Number string
		Hash   string
	}

	var Batch []rpc.BatchElem

	InitBatch(Batch, lastnumber, count, "descent")

	err = ec.evmClient.BatchCall(Batch)

	if err != nil {
		return nil, fmt.Errorf("can't get latest block:%s", err)

	}
	var blocks []*types.BlockInfo

	for _, batch := range Batch {
		nb := batch.Result.(*Block)
		num, err := strconv.ParseUint(strings.TrimPrefix(nb.Number, "0x"), 16, 64)
		if err != nil {
			fmt.Println("Error:", err)
		}
		ibatch := &types.BlockInfo{
			Height: num,
			Hash:   []byte(nb.Hash),
		}
		blocks = append(blocks, ibatch)
	}
	return blocks, nil
}

func (ec *EVMConsumerController) QueryBlocks(startHeight, endHeight, limit uint64) ([]*types.BlockInfo, error) {

	if endHeight < startHeight {
		return nil, fmt.Errorf("the startHeight %v should not be higher than the endHeight %v", startHeight, endHeight)
	}
	count := endHeight - startHeight
	if count > limit {
		count = limit
	}

	type Block struct {
		Number string
		Hash   string
	}

	startnumber := new(big.Int).SetUint64(startHeight)

	var Batch []rpc.BatchElem

	InitBatch(Batch, startnumber, count, "ascent")

	err := ec.evmClient.BatchCall(Batch)

	if err != nil {
		return nil, fmt.Errorf("can't get blocks")

	}

	var blocks []*types.BlockInfo

	for _, batch := range Batch {
		nb := batch.Result.(*Block)
		num, err := strconv.ParseUint(strings.TrimPrefix(nb.Number, "0x"), 16, 64)
		if err != nil {
			return nil, fmt.Errorf("error:%s", err)
		}

		ibatch := &types.BlockInfo{
			Height: num,
			Hash:   []byte(nb.Hash),
		}
		blocks = append(blocks, ibatch)
	}
	return blocks, nil
}

func (ec *EVMConsumerController) queryLatestBlocks(startKey []byte, count uint64, status finalitytypes.QueriedBlockStatus, reverse bool) ([]*types.BlockInfo, error) {
	var blocks []*types.BlockInfo
	// Can be deleted for never using
	return blocks, nil
}

func (ec *EVMConsumerController) QueryBlock(height uint64) (*types.BlockInfo, error) {

	number := new(big.Int).SetUint64(height)

	hexStr := Transform(number)

	type Block struct {
		Number string
		Hash   string
	}

	var block Block
	err := ec.evmClient.Call(&block, "eth_getBlockByNumber", hexStr, true)
	if err != nil {
		return nil, fmt.Errorf("can't get block by number:%s", err)
	}

	num, err := strconv.ParseUint(strings.TrimPrefix(block.Number, "0x"), 16, 64)
	if err != nil {
		return nil, fmt.Errorf("error:%s", err)
	}

	blockinfo := &types.BlockInfo{
		Height: num,
		Hash:   []byte(block.Hash),
	}

	return blockinfo, nil
}

func (ec *EVMConsumerController) QueryIsBlockFinalized(height uint64) (bool, error) {

	lastnumber, err := ec.GetLatestFinalizedNumber()
	if err != nil {
		return false, fmt.Errorf("can't get latest finalized block:%s", err)
	}
	number := new(big.Int).SetUint64(height)
	var finalized bool = false
	if number.Cmp(lastnumber) <= 0 {
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

	type Block struct {
		Number string
	}

	var block Block
	err := ec.evmClient.Call(&block, "eth_getBlockByNumber", "latest", true)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block:%s", err)
	}

	num, err := strconv.ParseUint(strings.TrimPrefix(block.Number, "0x"), 16, 64)
	if err != nil {
		fmt.Println("Error:%s", err)
	}

	return num, nil
}

func (ec *EVMConsumerController) Close() error {
	ec.evmClient.Close()
	return nil
}

func Transform(number *big.Int) string {

	hexStr := fmt.Sprintf("%x", number)
	if len(hexStr) >= 2 && hexStr[:2] != "0x" {
		hexStr = "0x" + hexStr
	}
	return hexStr
}

func (ec *EVMConsumerController) GetLatestFinalizedNumber() (*big.Int, error) {

	conn, err := ethclient.Dial(ec.cfg.RPCL1Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client:%s", err)
	}
	output, err := bindings.NewL2OutputOracle(common.HexToAddress(ec.cfg.L2OutputOracleContractAddress), conn)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate L2OutputOracle contract:%s ", err)
	}

	lastnumber, err := output.LatestBlockNumber(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest finalize block number:%s ", err)
	}
	return lastnumber, err
}

func InitBatch(Batch []rpc.BatchElem, number *big.Int, count uint64, order string) {

	type Block struct {
		Number string
		Hash   string
	}

	for i := 0; i < int(count); i++ {

		hexStr := Transform(number)
		ibatch := rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{hexStr, true},
			Result: new(Block),
		}
		Batch = append(Batch, ibatch)
		if order == "ascent" {
			number.Add(number, big.NewInt(1))
		} else if order == "descent" {
			number.Sub(number, big.NewInt(1))
		}
	}
}
