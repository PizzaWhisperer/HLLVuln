# HLLVuln

This code demonstrates how an attacker can perform a cardinality manipulation attack on the HyperLogLog sketch when having access to the low level details of the implementation.

## Running the attack
`go build` then `./HLLVuln`

Flag `-scenario` can be set to `"S2"` or `"S3"` to specify the adversarial model.
Flag `-userData` can be set to `false` to perform the attack on an empty datastructure.

## Benchmarking

Flag `-benchmark` can be set to any int `X` to loop `X` times and output the statistics.
