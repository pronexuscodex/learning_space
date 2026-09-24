# POWER ITERATION -- find a matrix's strongest direction by
# multiplying by it over and over.

A = [[2.0, 1.0],
     [1.0, 2.0]]

def multiply(m, v):
    return [m[0][0] * v[0] + m[0][1] * v[1],
            m[1][0] * v[0] + m[1][1] * v[1]]

v = [1.0, 0.0]                           # any starting vector
for step in range(1, 21):
    w = multiply(A, v)
    length = (w[0] ** 2 + w[1] ** 2) ** 0.5
    v = [w[0] / length, w[1] / length]
    if step in (1, 2, 5, 20):
        estimate = sum(a * b for a, b in zip(multiply(A, v), v))
        print(f"step {step:2}: direction ({v[0]:.4f}, {v[1]:.4f})  eigenvalue ~ {estimate:.4f}")
