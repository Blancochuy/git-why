"""Conservative, cross-process OpenRouter request pacing. No credentials stored."""

import hashlib
from contextlib import contextmanager
from pathlib import Path
import sqlite3
import time


class QuotaExhausted(Exception):
    pass


class Limiter:
    def __init__(self, path, key, interval=6, daily_limit=50, clock=time.time, sleep=time.sleep):
        self.path = Path(path)
        self.path.parent.mkdir(parents=True, exist_ok=True)
        self.identity = hashlib.sha256(key.encode()).hexdigest()
        self.interval, self.daily_limit = interval, daily_limit
        self.clock, self.sleep = clock, sleep
        with self.connect() as db:
            db.execute("CREATE TABLE IF NOT EXISTS pacing (identity TEXT PRIMARY KEY, next_at REAL NOT NULL)")
            db.execute("CREATE TABLE IF NOT EXISTS requests (identity TEXT NOT NULL, timestamp REAL NOT NULL)")
            db.execute("CREATE INDEX IF NOT EXISTS request_window ON requests(identity, timestamp)")

    @contextmanager
    def connect(self):
        db = sqlite3.connect(str(self.path), timeout=10)
        try:
            with db:
                yield db
        finally:
            db.close()

    def acquire(self, deadline, remaining_allowance=None, since=None):
        waited = 0.0
        while True:
            now = self.clock()
            with self.connect() as db:
                db.execute("BEGIN IMMEDIATE")
                count = db.execute("SELECT count(*) FROM requests WHERE identity=? AND timestamp>?",
                                   (self.identity, now - 86400)).fetchone()[0]
                if count >= self.daily_limit:
                    raise QuotaExhausted("Conservative rolling 24-hour local request cap reached")
                if remaining_allowance is not None and since is not None:
                    recent = db.execute("SELECT count(*) FROM requests WHERE identity=? AND timestamp>=?",
                                        (self.identity, since)).fetchone()[0]
                    if recent >= remaining_allowance:
                        raise QuotaExhausted("Server-reported free quota allowance exhausted")
                row = db.execute("SELECT next_at FROM pacing WHERE identity=?", (self.identity,)).fetchone()
                wait = max(0, (row[0] if row else now) - now)
                if wait == 0:
                    db.execute("INSERT OR REPLACE INTO pacing VALUES (?, ?)", (self.identity, now + self.interval))
                    db.execute("INSERT INTO requests VALUES (?, ?)", (self.identity, now))
                    return waited
            if time.monotonic() + wait >= deadline:
                raise QuotaExhausted("Rate-limit wait would exceed attempt deadline")
            self.sleep(wait)
            waited += wait

    def cooldown(self, seconds):
        with self.connect() as db:
            db.execute("BEGIN IMMEDIATE")
            row = db.execute("SELECT next_at FROM pacing WHERE identity=?", (self.identity,)).fetchone()
            next_at = max(row[0] if row else 0, self.clock() + seconds)
            db.execute("INSERT OR REPLACE INTO pacing VALUES (?, ?)", (self.identity, next_at))
