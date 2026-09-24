# SIEVE OF ERATOSTHENES -- the classic benchmark of home-computer
# magazines. Cross out multiples; whatever survives is prime.

LIMIT = 10_000
is_prime = [True] * LIMIT
is_prime[0] = is_prime[1] = False
for i in range(2, int(LIMIT ** 0.5) + 1):
    if is_prime[i]:
        for multiple in range(i * i, LIMIT, i):
            is_prime[multiple] = False

primes = [i for i, p in enumerate(is_prime) if p]
print("primes below", LIMIT, ":", len(primes))
print("last five:", primes[-5:])
