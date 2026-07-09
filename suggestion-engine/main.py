import json
import math
from http.server import HTTPServer, BaseHTTPRequestHandler

# Geofence restricted zone coordinates (Tan Son Nhat Airport area mock center)
NFZ_CENTER = (10.7725, 106.69)
NFZ_RADIUS_METERS = 800.0

def haversine(lat1, lon1, lat2, lon2):
    """Computes distance in meters between WGS84 points."""
    R = 6371000.0
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    delta_phi = math.radians(lat2 - lat1)
    delta_lambda = math.radians(lon2 - lon1)
    
    a = (math.sin(delta_phi/2.0)**2 + 
         math.cos(phi1) * math.cos(phi2) * math.sin(delta_lambda/2.0)**2)
    c = 2.0 * math.atan2(math.sqrt(a), math.sqrt(1.0 - a))
    return R * c

def is_point_in_polygon(lat, lng, polygon):
    """Ray-casting point containment checker."""
    n = len(polygon)
    inside = False
    p1y, p1x = polygon[0] # lat is Y, lng is X
    for i in range(n + 1):
        p2y, p2x = polygon[i % n]
        if lat > min(p1y, p2y):
            if lat <= max(p1y, p2y):
                if lng <= max(p1x, p2x):
                    if p1y != p2y:
                        xints = (lat - p1y) * (p2x - p1x) / (p2y - p1y) + p1x
                    if p1x == p2x or lng <= xints:
                        inside = not inside
        p1y, p1x = p2y, p2x
    return inside

class SuggestionHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        
        try:
            req = json.loads(post_data.decode('utf-8'))
        except Exception:
            self.send_response(400)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"error": "Invalid request body"}')
            return

        if self.path == '/api/suggest':
            # Drone Allocation suggestions
            response_payload = {
                "recommended_drone": "Drone-01 (M300 RTK)",
                "recommended_dock": "Dock Alpha",
                "distance_meters": 14.8,
                "success": True
            }
            
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response_payload).encode('utf-8'))

        elif self.path == '/api/plan':
            vertices_raw = req.get('vertices', [])
            if len(vertices_raw) < 3:
                self.send_response(400)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(b'{"error": "Polygon must contain at least 3 vertices"}')
                return

            vertices = [(float(v['lat']), float(v['lng'])) for v in vertices_raw]

            # 1. No-Fly Zone (Restricted Zone) Intersection Audit
            restricted_intersect = False
            for v in vertices:
                dist = haversine(v[0], v[1], NFZ_CENTER[0], NFZ_CENTER[1])
                if dist <= NFZ_RADIUS_METERS:
                    restricted_intersect = True
                    break

            if restricted_intersect:
                res = {
                    "safe": False,
                    "message": "Geofence intersects restricted airport airspace (No-Fly Zone)",
                    "path": []
                }
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(res).encode('utf-8'))
                return

            # 2. Lawnmower Sweep Path Grid Generation
            lats = [v[0] for v in vertices]
            lngs = [v[1] for v in vertices]
            min_lat, max_lat = min(lats), max(lats)
            min_lng, max_lng = min(lngs), max(lngs)

            path = []
            direction = 1
            lat_steps = int((max_lat - min_lat) / 0.0003) + 1
            lng_steps = int((max_lng - min_lng) / 0.0003) + 1

            for i in range(lat_steps):
                lat_val = min_lat + i * 0.0003
                line_points = []
                for j in range(lng_steps):
                    lng_val = min_lng + j * 0.0003
                    if is_point_in_polygon(lat_val, lng_val, vertices):
                        # Ensure 7 decimals precision rules (AD-6)
                        line_points.append({
                            "lat": round(lat_val, 7), 
                            "lng": round(lng_val, 7)
                        })
                
                # Alternate directions to simulate sweep turns
                if direction < 0:
                    line_points.reverse()
                path.extend(line_points)
                if len(line_points) > 0:
                    direction *= -1

            res = {
                "safe": True,
                "message": "Patrol flight path generated successfully",
                "path": path
            }

            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(res).encode('utf-8'))
        else:
            self.send_response(404)
            self.end_headers()

def run(port=50051):
    server_address = ('', port)
    httpd = HTTPServer(server_address, SuggestionHandler)
    print(f"Suggestion & Planner Engine listening on port {port}...")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass
    httpd.server_close()

if __name__ == '__main__':
    run()
