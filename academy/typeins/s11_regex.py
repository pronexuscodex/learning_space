# MATCH -- a tiny regular-expression matcher, in the spirit of the
# famous 30-line matchers of the Unix tradition. Supports:
#   c  a literal character      .  any character
#   ^  start of text            $  end of text
#   *  zero or more of the previous character

def match(pattern, text):
    if pattern.startswith("^"):
        return match_here(pattern[1:], text)
    for start in range(len(text) + 1):   # try every starting position
        if match_here(pattern, text[start:]):
            return True
    return False

def match_here(pattern, text):
    if pattern == "":
        return True
    if len(pattern) >= 2 and pattern[1] == "*":
        return match_star(pattern[0], pattern[2:], text)
    if pattern == "$":
        return text == ""
    if text and (pattern[0] == "." or pattern[0] == text[0]):
        return match_here(pattern[1:], text[1:])
    return False

def match_star(c, rest, text):
    while True:                          # try zero, one, two... copies of c
        if match_here(rest, text):
            return True
        if not text or (c != "." and text[0] != c):
            return False
        text = text[1:]

tests = [("^ab*c$", "ac"), ("^ab*c$", "abbbc"), ("^ab*c$", "abd"),
         ("o.d", "hello world"), ("^h.*d$", "hello world"), ("x", "hello")]
for pattern, text in tests:
    print(f"{pattern:8} {text!r:15} {match(pattern, text)}")
