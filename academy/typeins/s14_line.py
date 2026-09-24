# LEAST SQUARES -- the best straight line through some points,
# using the textbook formulas (no libraries).

xs = [1, 2, 3, 4, 5]
ys = [3.1, 4.9, 7.2, 8.8, 11.0]          # roughly y = 2x + 1, with noise

n = len(xs)
mean_x = sum(xs) / n
mean_y = sum(ys) / n
slope = (sum((x - mean_x) * (y - mean_y) for x, y in zip(xs, ys)) /
         sum((x - mean_x) ** 2 for x in xs))
intercept = mean_y - slope * mean_x
print(f"best line: y = {slope:.2f}x + {intercept:.2f}")

errors = [y - (slope * x + intercept) for x, y in zip(xs, ys)]
print("errors:", [round(e, 2) for e in errors])
print("sum of errors:", round(sum(errors), 10))
