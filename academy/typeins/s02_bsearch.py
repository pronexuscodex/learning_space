# BINARY SEARCH -- count how many guesses it takes.

def search(items, target):
    low, high = 0, len(items) - 1
    guesses = 0
    while low <= high:
        mid = (low + high) // 2
        guesses += 1
        if items[mid] == target:
            return mid, guesses
        if items[mid] < target:
            low = mid + 1
        else:
            high = mid - 1
    return -1, guesses

numbers = list(range(1, 1_000_001))      # 1 .. 1,000,000
for target in (1, 500_000, 999_999, 1_000_001):
    index, guesses = search(numbers, target)
    print(target, "->", "found" if index >= 0 else "missing", "in", guesses, "guesses")
