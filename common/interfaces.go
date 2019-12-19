package common

type BlockInterface interface {
	GetHeight() uint64
	Hash() *Hash
	GetProducer() string
	GetValidationField() string
	GetRound() int
	GetRoundKey() string
	GetInstructions() [][]string
	GetConsensusType() string
	GetEpoch() uint64
	GetPreviousViewHash() *Hash
	GetPreviousBlockHash() *Hash
	GetTimeslot() uint64
	GetBlockTimestamp() int64
}
