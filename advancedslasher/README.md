# Advanced Slasher Implementation

This is the main project folder for the advanced slasher implementation for eth2 written in Go by [Prysmatic Labs](https://prysmaticlabs.com). A slasher listens for all broadcasted messages using a running beacon node in order to detect slashable attestations and block proposals. 
It uses the [min-max-surround](https://github.com/protolambda/eth2-surround#min-max-surround) method by Protolambda.

The slasher requires a connection to a synced beacon node in order to listen for attestations and block proposals. To run the slasher, type:
```
bazel run //advancedslasher -- \
    --datadir PATH/FOR/DB \
    --beacon-rpc-provider localhost:4000
```
