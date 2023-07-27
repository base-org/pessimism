package registry

const (
	// Error constant strings
	invalidAddrErr  = "invalid address provided for heuristic. expected %s, got %s"
	couldNotCastErr = "could not cast transit data value to %s type"
	noNestedArgsErr = "no nested args found in session params"
	zeroAddressErr  = "provided address cannot be the zero address"

	// Event declaration strings
	OutputProposedEvent   = "OutputProposed(bytes32,uint256,uint256,uint256)"
	WithdrawalProvenEvent = "WithdrawalProven(bytes32,address,address)"
)
