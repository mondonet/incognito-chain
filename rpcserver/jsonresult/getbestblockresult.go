package jsonresult

import "github.com/incognitochain/incognito-chain/blockchain"

type GetBestBlockResult struct {
	BestBlocks map[int]GetBestBlockItem `json:"BestBlocks"`
}

type GetBestBlockItem struct {
	Height         uint64 `json:"Height"`
	Hash           string `json:"Hash"`
	TotalTxs       uint64 `json:"TotalTxs"`
	BlockProducer  string `json:"BlockProducer"`
	ValidationData string `json:"ValidationData"`
	Epoch          uint64 `json:"Epoch"`
	Time           int64  `json:"Time"`
}

func NewGetBestBlockItemFromShard(bestState *blockchain.ShardView) *GetBestBlockItem {
	result := &GetBestBlockItem{
		Height:         bestState.BestBlock.Header.Height,
		Hash:           bestState.BestBlockHash.String(),
		TotalTxs:       bestState.TotalTxns,
		BlockProducer:  bestState.BestBlock.Header.Producer,
		ValidationData: bestState.BestBlock.GetValidationField(),
		Time:           bestState.BestBlock.Header.Timestamp,
	}
	return result
}

func NewGetBestBlockItemFromBeacon(bestState *blockchain.BeaconView) *GetBestBlockItem {
	result := &GetBestBlockItem{
		Height:         bestState.BestBlock.Header.Height,
		Hash:           bestState.BestBlock.Hash().String(),
		BlockProducer:  bestState.BestBlock.Header.Producer,
		ValidationData: bestState.BestBlock.GetValidationField(),
		Epoch:          bestState.Epoch,
		Time:           bestState.BestBlock.Header.Timestamp,
	}
	return result
}

type GetBestBlockHashResult struct {
	BestBlockHashes map[int]string `json:"BestBlockHashes"`
}
