# WHO MAY DO WHAT -- role-based access control with deny by default.
# Roles grant permissions; the owner check stops editors from
# changing other people's documents.
ROLES = {
    "viewer": {"read"},
    "editor": {"read", "edit"},
    "admin":  {"read", "edit", "delete"},
}
USERS = {"ada": "admin", "bob": "editor", "cy": "viewer"}
DOCS = {1: "ada", 2: "bob"}          # document id -> owner

def can(user, action, doc):
    perms = ROLES.get(USERS.get(user), set())  # unknown user: no rights
    if action not in perms:
        return False                            # deny by default
    if action == "edit" and USERS[user] != "admin":
        return DOCS.get(doc) == user            # editors edit only their own
    return True

print("user  action  doc  allowed")
for user, action, doc in [("cy", "read", 1), ("cy", "edit", 1),
                          ("bob", "edit", 2), ("bob", "edit", 1),
                          ("bob", "delete", 2), ("ada", "delete", 2),
                          ("eve", "read", 1)]:
    print(f"{user:5} {action:7} {doc:3}  {can(user, action, doc)}")
