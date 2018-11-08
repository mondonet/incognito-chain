package blockchain

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
)

type GovernanceParams struct {
	SalaryPerTx uint64 // salary for each tx in block(mili constant)
	BasicSalary uint64 // basic salary per block(mili constant)
}

type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int

	// Hash of the previous block header in the block chain.
	PrevBlockHash common.Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot common.Hash

	// Merkle tree reference to hash of all commitments to the current block.
	MerkleRootCommitments common.Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint64 on the wire and therefore is limited to 2106.
	Timestamp int64

	// Parallel PoS
	BlockCommitteeSigs []string //Include producer and validators signature
	Committee          []string //Voted committee for the next block
	ChainID            byte
	ChainsHeight       []int //height of 20 chain when this block is created
	CandidateHash      common.Hash

	SalaryFund uint64 // use to pay salary for miners(block producer or current leader) in chain

	GovernanceParams GovernanceParams // params which get from governance for network

	LoanParams transaction.LoanParams // params for collateralized loans of Constant

	//Block Height
	Height int32
}
