# INJECTION -- the same login, written badly and written well.
import sqlite3

db = sqlite3.connect(":memory:")
db.execute("CREATE TABLE users (name TEXT, password TEXT)")
db.executemany("INSERT INTO users VALUES (?, ?)",
               [("alice", "s3cret"), ("bob", "hunter2"), ("carol", "letmein")])

typed_name = "' OR '1'='1"              # what an attacker types
typed_password = "' OR '1'='1"

unsafe = ("SELECT name FROM users WHERE name = '" + typed_name +
          "' AND password = '" + typed_password + "'")
print("unsafe query returns:", db.execute(unsafe).fetchall())

safe = "SELECT name FROM users WHERE name = ? AND password = ?"
print("safe query returns:  ", db.execute(safe, (typed_name, typed_password)).fetchall())
