# BRUTE-FORCE DETECTOR -- the heart of tools like fail2ban.
# Alert when one address fails to log in LIMIT times within WINDOW
# seconds. (Times are seconds; the addresses are documentation ones.)
from collections import defaultdict, deque

LIMIT, WINDOW = 5, 60
LOG = """\
0 203.0.113.9 FAIL root
2 203.0.113.9 FAIL admin
3 198.51.100.4 OK ada
5 203.0.113.9 FAIL root
9 203.0.113.9 FAIL test
14 198.51.100.4 FAIL ada
20 203.0.113.9 FAIL root
90 198.51.100.7 FAIL bob
95 198.51.100.7 FAIL bob
200 198.51.100.7 FAIL bob"""

recent = defaultdict(deque)   # address -> times of its recent failures
for line in LOG.splitlines():
    t, ip, result, user = line.split()
    t = int(t)
    if result != "FAIL":
        continue
    q = recent[ip]
    q.append(t)
    while q and q[0] <= t - WINDOW:   # forget failures outside the window
        q.popleft()
    if len(q) == LIMIT:
        print(f"t={t:>3}s ALERT {ip}: {LIMIT} failures in {t - q[0]}s (last user: {user})")

print("addresses seen failing:", ", ".join(sorted(recent)))
