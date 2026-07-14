import json
import threading
import unittest
import urllib.request
from http.server import HTTPServer

from main import SuggestionHandler, haversine, is_point_in_polygon, NFZ_CENTER


class TestGeometryHelpers(unittest.TestCase):
    def test_haversine_zero_distance(self):
        self.assertAlmostEqual(haversine(10.76, 106.66, 10.76, 106.66), 0.0, places=6)

    def test_haversine_known_distance(self):
        # ~111.19 km spans one degree of latitude at the equator.
        dist = haversine(0.0, 0.0, 1.0, 0.0)
        self.assertAlmostEqual(dist, 111195.0, delta=200.0)

    def test_point_in_polygon_inside(self):
        square = [(10.0, 106.0), (10.0, 106.01), (10.01, 106.01), (10.01, 106.0)]
        self.assertTrue(is_point_in_polygon(10.005, 106.005, square))

    def test_point_in_polygon_outside(self):
        square = [(10.0, 106.0), (10.0, 106.01), (10.01, 106.01), (10.01, 106.0)]
        self.assertFalse(is_point_in_polygon(20.0, 120.0, square))


class TestSuggestionServer(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = HTTPServer(("127.0.0.1", 0), SuggestionHandler)
        cls.port = cls.server.server_address[1]
        cls.thread = threading.Thread(target=cls.server.serve_forever, daemon=True)
        cls.thread.start()

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        cls.thread.join()

    def _post(self, path, payload):
        req = urllib.request.Request(
            f"http://127.0.0.1:{self.port}{path}",
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(req) as resp:
            return resp.status, json.loads(resp.read().decode("utf-8"))

    def test_suggest_endpoint(self):
        status, body = self._post("/api/suggest", {"lat": 10.76, "lng": 106.66})
        self.assertEqual(status, 200)
        self.assertTrue(body["success"])
        self.assertIn("recommended_drone", body)

    def test_plan_endpoint_returns_safe_path(self):
        vertices = [
            {"lat": 10.762, "lng": 106.660},
            {"lat": 10.763, "lng": 106.660},
            {"lat": 10.763, "lng": 106.661},
            {"lat": 10.762, "lng": 106.661},
        ]
        status, body = self._post("/api/plan", {"vertices": vertices})
        self.assertEqual(status, 200)
        self.assertTrue(body["safe"])
        self.assertGreater(len(body["path"]), 0)

    def test_plan_endpoint_rejects_too_few_vertices(self):
        req = urllib.request.Request(
            f"http://127.0.0.1:{self.port}/api/plan",
            data=json.dumps({"vertices": [{"lat": 1, "lng": 1}]}).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            urllib.request.urlopen(req)
            self.fail("expected HTTPError for polygon with < 3 vertices")
        except urllib.error.HTTPError as e:
            self.assertEqual(e.code, 400)

    def test_plan_endpoint_blocks_restricted_airspace(self):
        lat, lng = NFZ_CENTER
        vertices = [
            {"lat": lat, "lng": lng},
            {"lat": lat + 0.0001, "lng": lng},
            {"lat": lat + 0.0001, "lng": lng + 0.0001},
            {"lat": lat, "lng": lng + 0.0001},
        ]
        status, body = self._post("/api/plan", {"vertices": vertices})
        self.assertEqual(status, 200)
        self.assertFalse(body["safe"])


if __name__ == "__main__":
    unittest.main()
