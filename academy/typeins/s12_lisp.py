# LISP -- a calculator language in a few dozen lines:
# read the text, parse it into nested lists, evaluate them.
import operator

def tokenize(src):
    return src.replace("(", " ( ").replace(")", " ) ").split()

def parse(tokens):
    token = tokens.pop(0)
    if token == "(":
        expr = []
        while tokens[0] != ")":
            expr.append(parse(tokens))
        tokens.pop(0)                    # drop ")"
        return expr
    try:
        return int(token)
    except ValueError:
        return token                     # a symbol such as + or max

ENV = {"+": operator.add, "-": operator.sub, "*": operator.mul,
       "max": max, "min": min}

def evaluate(expr):
    if isinstance(expr, int):
        return expr
    if isinstance(expr, str):
        return ENV[expr]
    op, *args = [evaluate(e) for e in expr]
    return op(*args)

for program in ["(+ 1 2)", "(* 2 (+ 3 4))", "(max 3 (- 10 2) (* 2 2))"]:
    print(program, "=>", evaluate(parse(tokenize(program))))
