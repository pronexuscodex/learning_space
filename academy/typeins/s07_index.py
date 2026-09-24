# INDEX -- watch the database change its plan when an index appears.
import sqlite3

db = sqlite3.connect(":memory:")
db.execute("CREATE TABLE people (id INTEGER PRIMARY KEY, email TEXT, city TEXT)")
db.executemany("INSERT INTO people (email, city) VALUES (?, ?)",
               [(f"user{i}@example.com", f"city{i % 50}") for i in range(10_000)])

query = "SELECT id FROM people WHERE email = 'user9999@example.com'"

def plan():
    return [row[-1] for row in db.execute("EXPLAIN QUERY PLAN " + query)]

print("before index:", plan())
db.execute("CREATE INDEX idx_email ON people (email)")
print("after index: ", plan())
print("result:", db.execute(query).fetchall())
