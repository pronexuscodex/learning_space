# INTEGRITY CHECK -- the idea behind Tripwire and forensic hashing.
# Take a fingerprint (SHA-256) of every file while the system is known
# to be good; later, any change, addition or deletion stands out.
import hashlib

def fingerprint(files):
    return {name: hashlib.sha256(data).hexdigest()[:12] for name, data in files.items()}

before = {
    "/etc/passwd":   b"root:x:0:0:root:/root:/bin/bash\n",
    "/etc/ssh/sshd_config": b"PasswordAuthentication no\n",
    "/usr/bin/ls":   b"\x7fELF...the real ls...",
}
baseline = fingerprint(before)

after = dict(before)
after["/etc/ssh/sshd_config"] = b"PasswordAuthentication yes\n"    # weakened
after["/usr/bin/.cache"] = b"#!/bin/sh\ncurl ... | sh\n"         # dropped
del after["/usr/bin/ls"]                                           # removed
now = fingerprint(after)

for name in sorted(baseline.keys() | now.keys()):
    if name not in now:
        print(f"DELETED   {name}")
    elif name not in baseline:
        print(f"NEW       {name}  {now[name]}")
    elif now[name] != baseline[name]:
        print(f"MODIFIED  {name}  {baseline[name]} -> {now[name]}")
print("baseline of /etc/passwd:", baseline["/etc/passwd"], "(unchanged)")
