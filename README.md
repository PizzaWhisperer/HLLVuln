#  Source code of the attacks presented in the paper "HyperLogLog: exponentially bad"

Our attacks demonstrates how an adversary can perform a cardinality manipulation attack on the HyperLogLog sketch, when having access to various low level details of the implementation.

All attacks discussed in our paper are implemented and tested in this repository. 

## Attacks on Classic HLL

We attack ([clarkduvall's](https://github.com/clarkduvall/hyperloglog)) popular implementation of classic HLL.

Please make sure fist that Go is properly installed on your machine (see [the official documentation](https://go.dev/doc/install) if needed).

The, running `cd classic_hll/ && go run attack_hll.go`

will run all attacks for some pre-defined scenarios of interest and print the results.
If you want to test our attacks  in more settings, you can use the `runAttack` function where you can specify, among other variables, the adversarial scenario and the number of items already in the sketch. We refer to the API (L68 of the `attack_hll.go` file) for the exhaustive list of parameters.

Note 1: Depending on your setup you might have a module error. In our case, using the command `GO111MODULE=off go run attack_hll.go` instead did the trick.

Note 2: Be aware that we set the number of iterations to 50 by default, causing the attacks results to take quite some time to be printed. You can lower this number to have results faster.

## Attacks on Redis's implementation

Once in the `redis/` folder, run `make` to compile and install Redis. Our attack is contained in the new command `PFATTACK [key]`.

In a first tab spawn the server with `./src/redis-server`, and in a second tab the client with `./src/redis-cli`. Let the client perform the attack by running `PFATTACK hll` (on the client side). The initial cardinality of the attacked sketch will be printed on the server's console (after inserting `2^(t-1)*m*ln(m)` honest items first). Maliciously picked items are then added to the HLL sketch stored on the server's side. The command returns the number of inserted items. To observe the resulting change in cardinality, use the command `PFCOUNT hll` on the client after the attack.

There are several parameters that can changed in the `src/hyperloglog.c` file:
`T` controls the number of initial items to insert is controlled. It can be changed at the line 214. The possible values are 1: m*ln(m) items are added before the attack, 2: 2*m*ln(m) items are added, and 3: 4*m*ln(m) items are added. 
If you want to conduct an attack against an empty sketch, set `empty` to `true` line 213.
The attack parameter `epsilon` can be changed at line 210.
