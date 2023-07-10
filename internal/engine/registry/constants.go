package registry

const (
	// Error constant strings
	invalidAddrErr  = "invalid address provided for invariant. expected %s, got %s"
	couldNotCastErr = "could not cast transit data value to %s type"

	// Event declaration strings
	OutputProposedEvent   = "OutputProposed(bytes32,uint256,uint256,uint256)"
	WithdrawalProvenEvent = "WithdrawalProven(bytes32,address,address)"
)
