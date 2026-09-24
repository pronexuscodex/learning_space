# COLLATZ -- a type-in in the spirit of 1980s magazine listings.
# Start with n. If n is even, halve it; if odd, make it 3n + 1.
# Count the steps until n reaches 1.

n = 27
steps = 0
peak = n
while n != 1:
    if n % 2 == 0:
        n = n // 2
    else:
        n = 3 * n + 1
    steps = steps + 1
    if n > peak:
        peak = n
print("steps:", steps)
print("highest value:", peak)
