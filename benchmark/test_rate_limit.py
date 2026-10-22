from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
import tempfile
import time
import unittest

from rate_limit import Limiter, QuotaExhausted


class RateLimitTests(unittest.TestCase):
    def test_pacing_is_shared_across_instances_without_storing_key(self):
        with tempfile.TemporaryDirectory() as temp:
            state = Path(temp) / "state.sqlite"
            now = [1000.0]
            sleep = lambda seconds: now.__setitem__(0, now[0] + seconds)
            first = Limiter(state, "do-not-save-this-secret", 6, clock=lambda: now[0], sleep=sleep)
            second = Limiter(state, "do-not-save-this-secret", 6, clock=lambda: now[0], sleep=sleep)
            first.acquire(time.monotonic() + 100)
            self.assertEqual(second.acquire(time.monotonic() + 100), 6)
            first.cooldown(20)
            self.assertEqual(second.acquire(time.monotonic() + 100), 20)
            self.assertNotIn(b"do-not-save-this-secret", state.read_bytes())

    def test_daily_quota_reservation_is_atomic(self):
        with tempfile.TemporaryDirectory() as temp:
            state = Path(temp) / "state.sqlite"
            limiters = [Limiter(state, "shared-key", interval=0, daily_limit=1) for _ in range(2)]
            def reserve(limiter):
                try:
                    limiter.acquire(time.monotonic() + 10)
                    return True
                except QuotaExhausted:
                    return False
            with ThreadPoolExecutor(max_workers=2) as pool:
                self.assertEqual(sorted(pool.map(reserve, limiters)), [False, True])

    def test_server_allowance_and_deadline(self):
        with tempfile.TemporaryDirectory() as temp:
            state = Path(temp) / "state.sqlite"
            now = [1000.0]
            limiter = Limiter(state, "key", interval=6, clock=lambda: now[0])
            limiter.acquire(time.monotonic() + 100, remaining_allowance=1, since=999)
            with self.assertRaises(QuotaExhausted):
                limiter.acquire(time.monotonic() + 100, remaining_allowance=1, since=999)
            with self.assertRaises(QuotaExhausted):
                limiter.acquire(time.monotonic() + 1)


if __name__ == "__main__":
    unittest.main()
