# AUTOGRAD -- a value that remembers how it was made, so gradients
# can flow backwards through the chain rule.

class Value:
    def __init__(self, data, parents=(), backward=lambda: None):
        self.data, self.grad = data, 0.0
        self.parents, self._backward = parents, backward

    def __add__(self, other):
        out = Value(self.data + other.data, (self, other))
        def backward():
            self.grad += out.grad
            other.grad += out.grad
        out._backward = backward
        return out

    def __mul__(self, other):
        out = Value(self.data * other.data, (self, other))
        def backward():
            self.grad += other.data * out.grad
            other.grad += self.data * out.grad
        out._backward = backward
        return out

    def backward(self):
        order, seen = [], set()
        def visit(v):
            if v not in seen:
                seen.add(v)
                for p in v.parents:
                    visit(p)
                order.append(v)
        visit(self)
        self.grad = 1.0
        for v in reversed(order):
            v._backward()

w, x, b = Value(2.0), Value(3.0), Value(1.0)
y = w * x + b          # y = 7
L = y * y              # L = 49
L.backward()
print("L =", L.data)
print("dL/dw =", w.grad, " dL/dx =", x.grad, " dL/db =", b.grad)
