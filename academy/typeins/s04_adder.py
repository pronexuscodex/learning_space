# RIPPLE-CARRY ADDER -- addition built only from logic gates.

def AND(a, b): return a & b
def XOR(a, b): return a ^ b
def OR(a, b):  return a | b

def full_adder(a, b, carry_in):
    s = XOR(XOR(a, b), carry_in)
    carry_out = OR(AND(a, b), AND(carry_in, XOR(a, b)))
    return s, carry_out

def add8(x, y):
    carry, result = 0, 0
    for bit in range(8):                     # least significant bit first
        a = (x >> bit) & 1
        b = (y >> bit) & 1
        s, carry = full_adder(a, b, carry)
        result |= s << bit
    return result, carry

for x, y in ((12, 30), (200, 100), (255, 1)):
    total, carry = add8(x, y)
    print(f"{x} + {y} = {total} (carry out {carry})  bits {total:08b}")
