package kv

var (
	// Genesis information bucket key.
	genesisInfoBucket = []byte("genesis-info-bucket")
	// Genesis validators root key.
	genesisValidatorsRootKey = []byte("genesis-val-root")

	// Validator slashing protection from double proposals.
	historicProposalsBucket = []byte("proposal-history-bucket")
	// Validator slashing protection from double proposals.
	newhistoricProposalsBucket = []byte("proposal-history-bucket-interchange")
	// Validator slashing protection from slashable attestations.
	historicAttestationsBucket = []byte("attestation-history-bucket")
	// New Validator slashing protection from slashable attestations.
	newHistoricAttestationsBucket = []byte("attestation-history-bucket-interchange")
	// Key prefix to minimal attestation source epoch in attestation bucket.
	minimalAttestationSourceEpochKeyPrefix = "minimal-attestation-source-epoch"
	// Key prefix to minimal attestation target epoch in attestation bucket.
	minimalAttestationTargetEpochKeyPrefix = "minimal-attestation-target-epoch"
)

// GetMinTargetKey given a public key returns the min source db key.
func GetMinSourceKey(pubKey [48]byte) []byte {
	return append([]byte(minimalAttestationSourceEpochKeyPrefix), pubKey[:]...)
}

// GetMinTargetKey given a public key returns the min target db key.
func GetMinTargetKey(pubKey [48]byte) []byte {
	return append([]byte(minimalAttestationTargetEpochKeyPrefix), pubKey[:]...)
}
