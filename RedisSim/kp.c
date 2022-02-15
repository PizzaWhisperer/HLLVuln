#include<stdio.h>
#include<math.h>
#include<stdint.h>

int main()
{
	uint64_t n=  14; 							// number of bits to determine bucket - this setting is for Redis
	uint64_t m = ((uint64_t) 1) << n; 	// number of buckets
	uint64_t ell = 50; 							// number of bits to determine bucket updates - this setting also for Redis
	int ell_test; 									// difference to make to D
	double alpha = 0.721; 					// magic constant
	printf("m: %6llu\n",m);
	double CE = 100000000.0; 				// starting cardinality estimate; vary this to simulate different conditions but don't go too small or the maths falls apart
	printf("CE: %6f\n",CE);
	double N = alpha*m*m; 				// numerator in CE equation, a fixed function of alpha and m
	printf("N: %6f\n",N);
	double D = N/CE; 							// denominator - CE = N/D
	printf("D: %6f\n",D);
	double delta = 1.0;
	for (ell_test = 0; ell_test <= ell; ell_test++) {
		delta = 0.5*delta; 						//update delta by halving it each time around the loop 
		printf("ell_test: %6d\n",ell_test);
		printf("delta: %.10e\n",delta);
		double newD = D - delta; 			// newD updates D with the change // here we can get underflows; a real implementation of HLL would not die here because it computes D afresh each time, though it's clear the Redis implementation will have accuracy issues for \ell = 50 unless it uses quad accuracy?
		if (newD < 0) 		{					// this can't occur in a real implementation but can in our simulation; this simply means ell_test is too small to be consistent with the starting value of CE
			printf(" Value ell_test = %6d not consistent with starting CE\n",ell_test);  // this is not an error, just a limitation of the simulation
			continue; 
			}
		printf("oldD: %.10e\n",D);
		printf("newD: %.10e\n",newD);
		double newCE = N/newD; 		// newCE updates the value of CE = N/D to the new value
		printf("new CE: %.10e\n",newCE);
		printf("CE difference: %.10e\n",newCE-CE);
		if (newCE-CE > 1)
			printf("Success for ell_test = %6d\n",ell_test); // we want a difference larger than 1 between old and new CE so as to survive any rounding
		else 
			{
			printf("Failure for ell_test = %6d\n",ell_test); // 
			break; //lets not go any further on failure
			}
		printf("\n");
	}

}