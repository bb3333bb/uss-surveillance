import React, { useState, useEffect, useRef } from 'react';
import { 
  ThemeProvider, CssBaseline, Button, Box, Typography, Paper, Avatar, 
  Tabs, Tab, Grid, AppBar, Toolbar, Chip, IconButton 
} from '@mui/material';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import FlightTakeoffIcon from '@mui/icons-material/FlightTakeoff';
import LogoutIcon from '@mui/icons-material/Logout';
import MenuOpenIcon from '@mui/icons-material/MenuOpen';
import VideocamIcon from '@mui/icons-material/Videocam';
import PauseIcon from '@mui/icons-material/Pause';
import HomeIcon from '@mui/icons-material/Home';
import LayersIcon from '@mui/icons-material/Layers';
import GestureIcon from '@mui/icons-material/Gesture';
import DeleteIcon from '@mui/icons-material/Delete';
import WarningIcon from '@mui/icons-material/Warning';

import L from 'leaflet';
import axios from 'axios';
import { useAuth } from './context/AuthContext.jsx';
import theme from './theme.js';

function App() {
  const { isAuthenticated, user, role, login, logout, loading } = useAuth();
  const [tabVal, setTabVal] = useState(0);
  const [leftOpen, setLeftOpen] = useState(true);
  const [rightOpen, setRightOpen] = useState(true);

  // Map & drawing state
  const mapRef = useRef(null);
  const leafletMapRef = useRef(null);
  const flightPathLayerRef = useRef(null);

  const [isDrawing, setIsDrawing] = useState(false);
  const [polygonPoints, setPolygonPoints] = useState([]);
  const [weatherInfo, setWeatherInfo] = useState(null);
  const [suggestion, setSuggestion] = useState(null);
  const [flightPath, setFlightPath] = useState(null);
  const [plannerSafe, setPlannerSafe] = useState(null);
  const [plannerMessage, setPlannerMessage] = useState(null);

  // Mutable drawing layers tracking (using refs to bypass React re-renders during drag/draw)
  const drawStateRef = useRef({
    isDrawing: false,
    vertices: [],
    markers: [],
    tempLines: null,
    cursorLine: null,
    polygonShape: null
  });

  // Keep drawing boolean state in sync
  useEffect(() => {
    drawStateRef.current.isDrawing = isDrawing;
  }, [isDrawing]);

  // Query Planner Engine for path sweep grid
  const fetchFlightPlan = async (vertices) => {
    try {
      const response = await axios.post('http://localhost:8080/api/operator/plan', {
        vertices: vertices
      });
      setPlannerSafe(response.data.safe);
      setPlannerMessage(response.data.message);
      if (response.data.safe) {
        setFlightPath(response.data.path);
      } else {
        setFlightPath([]);
      }
    } catch (err) {
      console.error("Failed to query flight plan:", err);
    }
  };

  // Query Suggestion Engine
  const fetchSuggestion = async (centroid) => {
    try {
      const response = await axios.post('http://localhost:8080/api/operator/suggestion', {
        lat: centroid.lat,
        lng: centroid.lng
      });
      setSuggestion(response.data);
    } catch (err) {
      console.error("Failed to query suggestion engine:", err);
    }
  };

  // Fetch weather parameters from Go Gateway
  const fetchWeatherAlerts = async (lat, lng, vertices) => {
    try {
      const response = await axios.post('http://localhost:8080/api/operator/weather', {
        lat: parseFloat(lat.toFixed(7)),
        lng: parseFloat(lng.toFixed(7))
      });
      setWeatherInfo(response.data);
      
      // If weather conditions are safe, fetch allocations & flight grid plans
      if (response.data.safe) {
        fetchSuggestion({ lat, lng });
        fetchFlightPlan(vertices);
      }
    } catch (err) {
      console.error("Failed to query weather alerts:", err);
    }
  };

  // Map Initialization
  useEffect(() => {
    if (!isAuthenticated || !mapRef.current) return;

    const map = L.map(mapRef.current, {
      center: [10.762622, 106.660172],
      zoom: 16,
      zoomControl: false
    });
    L.control.zoom({ position: 'bottomright' }).addTo(map);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution: '&copy; <a href="https://openstreetmap.org/copyright">OpenStreetMap</a> contributors'
    }).addTo(map);

    leafletMapRef.current = map;

    // Setup drawing listeners
    const handleMapClick = (e) => {
      const state = drawStateRef.current;
      if (!state.isDrawing) return;

      const latlng = e.latlng;
      const vertex = [latlng.lat, latlng.lng];
      state.vertices.push(vertex);

      // Create circle marker for vertex
      const isFirst = state.vertices.length === 1;
      const marker = L.circleMarker(latlng, {
        radius: isFirst ? 8 : 6,
        color: isFirst ? '#00e5ff' : '#00b0ff',
        fillColor: '#0a192f',
        fillOpacity: 1,
        weight: 2
      }).addTo(map);

      state.markers.push(marker);

      if (isFirst) {
        marker.on('click', (ev) => {
          L.DomEvent.stopPropagation(ev);
          closePolygonShape();
        });
      }

      if (!state.tempLines) {
        state.tempLines = L.polyline(state.vertices, {
          color: '#00e5ff',
          dashArray: '5, 5',
          weight: 2
        }).addTo(map);
      } else {
        state.tempLines.setLatLngs(state.vertices);
      }

      if (!state.cursorLine) {
        state.cursorLine = L.polyline([latlng, latlng], {
          color: '#00e5ff',
          dashArray: '5, 5',
          weight: 1.5
        }).addTo(map);
      }
    };

    const handleMouseMove = (e) => {
      const state = drawStateRef.current;
      if (!state.isDrawing || state.vertices.length === 0 || !state.cursorLine) return;
      
      const lastVertex = state.vertices[state.vertices.length - 1];
      state.cursorLine.setLatLngs([lastVertex, [e.latlng.lat, e.latlng.lng]]);
    };

    map.on('click', handleMapClick);
    map.on('mousemove', handleMouseMove);

    return () => {
      map.off('click', handleMapClick);
      map.off('mousemove', handleMouseMove);
      map.remove();
      leafletMapRef.current = null;
    };
  }, [isAuthenticated]);

  // Leaflet Flight Path Layer Rendering
  useEffect(() => {
    const map = leafletMapRef.current;
    if (!map) return;

    if (flightPathLayerRef.current) {
      map.removeLayer(flightPathLayerRef.current);
      flightPathLayerRef.current = null;
    }

    if (flightPath && flightPath.length > 0) {
      const latlngs = flightPath.map(pt => [pt.lat, pt.lng]);
      const group = L.featureGroup().addTo(map);

      // Lawnmower Grid lines
      L.polyline(latlngs, {
        color: '#00e5ff',
        weight: 3,
        opacity: 0.85
      }).addTo(group);

      // Takeoff Marker
      L.circleMarker(latlngs[0], {
        radius: 6,
        color: '#2e7d32',
        fillColor: '#0a192f',
        fillOpacity: 1,
        weight: 2
      }).bindTooltip("Takeoff").addTo(group);

      // Landing Marker
      L.circleMarker(latlngs[latlngs.length - 1], {
        radius: 6,
        color: '#c62828',
        fillColor: '#0a192f',
        fillOpacity: 1,
        weight: 2
      }).bindTooltip("Landing").addTo(group);

      flightPathLayerRef.current = group;
    }
  }, [flightPath]);

  // Close Active Drawing Boundary
  const closePolygonShape = () => {
    const map = leafletMapRef.current;
    const state = drawStateRef.current;
    if (!map || state.vertices.length < 3) return;

    state.polygonShape = L.polygon(state.vertices, {
      color: '#00b0ff',
      fillColor: '#00b0ff',
      fillOpacity: 0.25,
      weight: 2
    }).addTo(map);

    const closedCoords = state.vertices.map(v => ({
      lat: parseFloat(v[0].toFixed(7)),
      lng: parseFloat(v[1].toFixed(7))
    }));

    setPolygonPoints(closedCoords);
    console.log("Boundary geofence coordinates finalized (WGS84):", JSON.stringify(closedCoords));

    let latSum = 0;
    let lngSum = 0;
    state.vertices.forEach(v => {
      latSum += v[0];
      lngSum += v[1];
    });
    const centroidLat = latSum / state.vertices.length;
    const centroidLng = lngSum / state.vertices.length;

    // Trigger API validation pipeline
    fetchWeatherAlerts(centroidLat, centroidLng, closedCoords);

    cleanupTemporaryDrawLayers();
    setIsDrawing(false);
  };

  const cleanupTemporaryDrawLayers = () => {
    const map = leafletMapRef.current;
    const state = drawStateRef.current;
    if (!map) return;

    state.markers.forEach(m => map.removeLayer(m));
    state.markers = [];
    if (state.tempLines) {
      map.removeLayer(state.tempLines);
      state.tempLines = null;
    }
    if (state.cursorLine) {
      map.removeLayer(state.cursorLine);
      state.cursorLine = null;
    }
  };

  // Toggle boundary drawing mode
  const startDrawingBoundary = () => {
    if (isDrawing) {
      cleanupTemporaryDrawLayers();
      drawStateRef.current.vertices = [];
      setIsDrawing(false);
    } else {
      clearGeofenceBoundary();
      setIsDrawing(true);
    }
  };

  // Reset geofence states
  const clearGeofenceBoundary = () => {
    const map = leafletMapRef.current;
    const state = drawStateRef.current;
    if (!map) return;

    cleanupTemporaryDrawLayers();
    if (state.polygonShape) {
      map.removeLayer(state.polygonShape);
      state.polygonShape = null;
    }
    if (flightPathLayerRef.current) {
      map.removeLayer(flightPathLayerRef.current);
      flightPathLayerRef.current = null;
    }
    state.vertices = [];
    setPolygonPoints([]);
    setWeatherInfo(null);
    setSuggestion(null);
    setFlightPath(null);
    setPlannerSafe(null);
    setPlannerMessage(null);
  };

  if (loading) {
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
          <Typography color="textSecondary">Loading portal...</Typography>
        </Box>
      </ThemeProvider>
    );
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {!isAuthenticated ? (
        <Box 
          display="flex" 
          justifyContent="center" 
          alignItems="center" 
          minHeight="100vh"
          sx={{ backgroundColor: '#0a192f' }}
        >
          <Paper 
            elevation={4} 
            sx={{ 
              p: 4, 
              display: 'flex', 
              flexDirection: 'column', 
              alignItems: 'center',
              backgroundColor: 'background.paper',
              borderRadius: 2,
              border: '1px solid #233554',
              maxWidth: 400
            }}
          >
            <Avatar sx={{ m: 1, bgcolor: 'primary.main' }}>
              <LockOutlinedIcon />
            </Avatar>
            <Typography component="h1" variant="h5" sx={{ mb: 1, color: 'text.primary', fontWeight: 600 }}>
              USS Surveillance
            </Typography>
            <Typography variant="body2" color="textSecondary" align="center" sx={{ mb: 3 }}>
              On-Premises Drone Fleet Control Portal
            </Typography>
            <Button
              fullWidth
              variant="contained"
              color="primary"
              onClick={login}
              sx={{ py: 1.5, fontWeight: 600 }}
            >
              Sign In with SSO
            </Button>
          </Paper>
        </Box>
      ) : (
        <Box display="flex" flexDirection="column" height="100vh">
          {/* Header */}
          <AppBar position="static" sx={{ bgcolor: 'background.paper', borderBottom: '1px solid #233554', backgroundImage: 'none' }}>
            <Toolbar sx={{ justifyContent: 'space-between' }}>
              <Box display="flex" alignItems="center">
                <Avatar sx={{ bgcolor: 'primary.main', mr: 1.5, width: 32, height: 32 }}>
                  <FlightTakeoffIcon sx={{ fontSize: 18 }} />
                </Avatar>
                <Typography variant="h6" component="div" sx={{ fontWeight: 600, color: 'text.primary' }}>
                  USS Surveillance
                </Typography>
                <Chip 
                  label="ON-PREMISES" 
                  size="small" 
                  color="success" 
                  variant="outlined" 
                  sx={{ ml: 2, fontSize: '0.65rem', height: 20, borderColor: '#00e5ff', color: '#00e5ff' }} 
                />
              </Box>

              <Box display="flex" alignItems="center">
                <Box sx={{ textAlign: 'right', mr: 2 }}>
                  <Typography variant="body2" sx={{ fontWeight: 600, color: 'text.primary' }}>
                    {user}
                  </Typography>
                  <Typography variant="caption" color="textSecondary" sx={{ display: 'block', textTransform: 'capitalize' }}>
                    Role: {role}
                  </Typography>
                </Box>
                <Button 
                  variant="outlined" 
                  color="error" 
                  size="small" 
                  startIcon={<LogoutIcon />} 
                  onClick={logout}
                  sx={{ textTransform: 'none' }}
                >
                  Log Out
                </Button>
              </Box>
            </Toolbar>
          </AppBar>

          {/* 3-Column Body */}
          <Box display="flex" flexGrow={1} overflow="hidden">
            {/* Left Sidebar */}
            <Box 
              width={leftOpen ? 320 : 0} 
              sx={{ 
                transition: 'width 0.2s', 
                overflow: 'hidden', 
                borderRight: leftOpen ? '1px solid #233554' : 'none', 
                bgcolor: 'background.paper',
                display: 'flex',
                flexDirection: 'column'
              }}
            >
              <Box sx={{ borderBottom: 1, borderColor: '#233554' }}>
                <Tabs value={tabVal} onChange={(e, val) => setTabVal(val)} variant="fullWidth">
                  <Tab label="Fleet" sx={{ textTransform: 'none', fontWeight: 600 }} />
                  <Tab label="History" sx={{ textTransform: 'none', fontWeight: 600 }} />
                </Tabs>
              </Box>

              <Box p={2} flexGrow={1} overflow="auto">
                {tabVal === 0 ? (
                  <Box>
                    <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
                      ACTIVE HUBS
                    </Typography>
                    <Paper sx={{ p: 1.5, mb: 2, border: '1px solid #233554', bgcolor: '#0a192f' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600 }}>Dock Alpha</Typography>
                      <Typography variant="caption" color="success.main" sx={{ display: 'block' }}>● Online (Doors Closed)</Typography>
                      <Typography variant="caption" color="textSecondary">Temp: 24.5°C | Charging Contact: Active</Typography>
                    </Paper>

                    <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
                      DRONE FLEET
                    </Typography>
                    <Paper sx={{ p: 1.5, mb: 2, border: '1px solid #233554', bgcolor: '#0a192f' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600 }}>Drone-01 (M300 RTK)</Typography>
                      <Typography variant="caption" color="success.main" sx={{ display: 'block' }}>● Docked</Typography>
                      <Typography variant="caption" color="textSecondary">Battery: 92% | GPS: 3D Fix</Typography>
                    </Paper>

                    {polygonPoints.length > 0 && (
                      <>
                        <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
                          ACTIVE BOUNDARY POLYGON
                        </Typography>
                        <Paper sx={{ p: 1.5, border: '1px solid #233554', bgcolor: '#0a192f', mb: 2 }}>
                          <Typography variant="caption" color="secondary" sx={{ fontWeight: 600, display: 'block', mb: 1 }}>
                            WGS84 Coordinates ({polygonPoints.length} vertices):
                          </Typography>
                          <Box sx={{ maxHeight: 150, overflow: 'auto', bgcolor: '#061325', p: 1, borderRadius: 1 }}>
                            {polygonPoints.map((pt, i) => (
                              <Typography key={i} variant="caption" sx={{ fontFamily: 'monospace', display: 'block', color: 'text.primary' }}>
                                [{i}]: {pt.lat.toFixed(7)}, {pt.lng.toFixed(7)}
                              </Typography>
                            ))}
                          </Box>
                        </Paper>
                      </>
                    )}

                    {/* Fleet Allocation Suggestion Card */}
                    {suggestion && (
                      <>
                        <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
                          OPTIMAL ALLOCATION
                        </Typography>
                        <Paper sx={{ p: 1.5, border: '1px solid #00e5ff', bgcolor: '#061325' }}>
                          <Typography variant="body2" sx={{ fontWeight: 600, color: '#00e5ff' }}>
                            {suggestion.recommended_drone}
                          </Typography>
                          <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mt: 0.5 }}>
                            Launch Dock: <strong>{suggestion.recommended_dock}</strong>
                          </Typography>
                          <Typography variant="caption" color="textSecondary" sx={{ display: 'block' }}>
                            Distance to Centroid: {suggestion.distance_meters.toFixed(1)} m
                          </Typography>
                        </Paper>
                      </>
                    )}
                  </Box>
                ) : (
                  <Box>
                    <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
                      COMPLETED MISSIONS
                    </Typography>
                    <Typography variant="body2" color="textSecondary">
                      No logs found. Run a mission to generate replay data.
                    </Typography>
                  </Box>
                )}
              </Box>
            </Box>

            {/* Toggle Left Sidebar Button */}
            <Box display="flex" alignItems="center" bgcolor="background.default">
              <IconButton 
                size="small" 
                onClick={() => setLeftOpen(!leftOpen)} 
                sx={{ 
                  color: 'primary.main', 
                  transform: leftOpen ? 'rotate(0deg)' : 'rotate(180deg)',
                  transition: 'transform 0.2s',
                  m: 0.5 
                }}
              >
                <MenuOpenIcon />
              </IconButton>
            </Box>

            {/* Center Map Panel */}
            <Box flexGrow={1} display="flex" flexDirection="column" position="relative" bgcolor="#0a192f">
              <Box p={2} sx={{ borderBottom: '1px solid #233554', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, color: 'text.primary' }}>
                  2D Active Patrol Map
                </Typography>
                
                <Box display="flex" gap={1}>
                  {/* Drawing Control Controls (Only for non-viewer roles) */}
                  {role !== 'viewer' && (
                    <>
                      {polygonPoints.length === 0 ? (
                        <Button 
                          variant="outlined" 
                          size="small" 
                          color="secondary" 
                          startIcon={<GestureIcon />}
                          onClick={startDrawingBoundary}
                          sx={{ textTransform: 'none' }}
                        >
                          {isDrawing ? "Cancel Draw" : "Draw Area"}
                        </Button>
                      ) : (
                        <Button 
                          variant="outlined" 
                          size="small" 
                          color="error" 
                          startIcon={<DeleteIcon />}
                          onClick={clearGeofenceBoundary}
                          sx={{ textTransform: 'none' }}
                        >
                          Clear Area
                        </Button>
                      )}
                    </>
                  )}

                  {role !== 'viewer' && (
                    <Button variant="outlined" size="small" color="primary" startIcon={<LayersIcon />} sx={{ textTransform: 'none' }}>
                      Toggle 3D View
                    </Button>
                  )}
                </Box>
              </Box>

              {/* Leaflet Map DOM Canvas */}
              <Box flexGrow={1} position="relative" sx={{ width: '100%', height: '100%' }}>
                <div ref={mapRef} style={{ width: '100%', height: '100%', outline: 'none' }} />
                
                {/* Visual state indicator overlay */}
                {isDrawing && (
                  <Box 
                    position="absolute" 
                    top={12} 
                    left="50%" 
                    sx={{ 
                      transform: 'translateX(-50%)', 
                      bgcolor: 'background.paper', 
                      px: 2, 
                      py: 0.75, 
                      borderRadius: 1, 
                      border: '1px solid #00e5ff',
                      pointerEvents: 'none',
                      zIndex: 1000
                    }}
                  >
                    <Typography variant="caption" sx={{ color: '#00e5ff', fontWeight: 600 }}>
                      [Drawing Active] Click map to add vertices. Click first vertex to close path.
                    </Typography>
                  </Box>
                )}

                {/* Weather Alert Banner */}
                {weatherInfo && !weatherInfo.safe && (
                  <Box 
                    position="absolute" 
                    top={12} 
                    left="50%" 
                    sx={{ 
                      transform: 'translateX(-50%)', 
                      bgcolor: 'error.main', 
                      px: 2.5, 
                      py: 1, 
                      borderRadius: 1, 
                      boxShadow: '0 4px 15px rgba(255, 61, 0, 0.45)',
                      pointerEvents: 'none',
                      zIndex: 1000,
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1
                    }}
                  >
                    <WarningIcon sx={{ fontSize: 18, color: '#ffffff' }} />
                    <Typography variant="caption" sx={{ color: '#ffffff', fontWeight: 700 }}>
                      WEATHER SUSPENDED: Wind Speed {weatherInfo.wind_speed} m/s ({weatherInfo.precipitation})!
                    </Typography>
                  </Box>
                )}

                {/* Restricted Airspace Warning Banner */}
                {plannerSafe === false && (
                  <Box 
                    position="absolute" 
                    top={12} 
                    left="50%" 
                    sx={{ 
                      transform: 'translateX(-50%)', 
                      bgcolor: 'error.main', 
                      px: 2.5, 
                      py: 1, 
                      borderRadius: 1, 
                      boxShadow: '0 4px 15px rgba(255, 61, 0, 0.45)',
                      pointerEvents: 'none',
                      zIndex: 1000,
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1
                    }}
                  >
                    <WarningIcon sx={{ fontSize: 18, color: '#ffffff' }} />
                    <Typography variant="caption" sx={{ color: '#ffffff', fontWeight: 700 }}>
                      LAUNCH LOCK: Restricted No-Fly Zone Airspace Intersection!
                    </Typography>
                  </Box>
                )}
              </Box>
            </Box>

            {/* Toggle Right Sidebar Button */}
            <Box display="flex" alignItems="center" bgcolor="background.default">
              <IconButton 
                size="small" 
                onClick={() => setRightOpen(!rightOpen)} 
                sx={{ 
                  color: 'primary.main', 
                  transform: rightOpen ? 'rotate(180deg)' : 'rotate(0deg)',
                  transition: 'transform 0.2s',
                  m: 0.5 
                }}
              >
                <MenuOpenIcon />
              </IconButton>
            </Box>

            {/* Right Sidebar */}
            <Box 
              width={rightOpen ? 380 : 0} 
              sx={{ 
                transition: 'width 0.2s', 
                overflow: 'hidden', 
                borderLeft: rightOpen ? '1px solid #233554' : 'none', 
                bgcolor: 'background.paper',
                display: 'flex',
                flexDirection: 'column'
              }}
            >
              <Box p={2} sx={{ borderBottom: '1px solid #233554' }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 600 }}>
                  LIVE VIDEO MONITORING
                </Typography>
              </Box>
              <Box p={2}>
                <Paper 
                  sx={{ 
                    height: 200, 
                    display: 'flex', 
                    flexDirection: 'column', 
                    justifyContent: 'center', 
                    alignItems: 'center', 
                    bgcolor: '#0a192f',
                    border: '1px dashed #233554',
                    position: 'relative'
                  }}
                >
                  <VideocamIcon sx={{ fontSize: 40, color: 'text.secondary', mb: 1 }} />
                  <Typography variant="caption" color="textSecondary">WebRTC Stream Playout Idle</Typography>
                </Paper>
              </Box>

              <Box p={2} sx={{ borderTop: '1px solid #233554' }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1.5 }}>
                  DRONE TELEMETRY
                </Typography>
                <Grid container spacing={1}>
                  <Grid item xs={6}>
                    <Paper p={1} sx={{ p: 1, bgcolor: '#0a192f', border: '1px solid #233554' }}>
                      <Typography variant="caption" color="textSecondary">Altitude</Typography>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>0.0 m</Typography>
                    </Paper>
                  </Grid>
                  <Grid item xs={6}>
                    <Paper p={1} sx={{ 
                      p: 1, 
                      bgcolor: '#0a192f', 
                      border: weatherInfo && !weatherInfo.safe ? '1px solid #ff3d00' : '1px solid #233554',
                      color: weatherInfo && !weatherInfo.safe ? '#ff3d00' : 'text.primary'
                    }}>
                      <Typography variant="caption" color="textSecondary">Wind Speed</Typography>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
                        {weatherInfo ? `${weatherInfo.wind_speed.toFixed(1)} m/s` : '1.2 m/s'}
                      </Typography>
                    </Paper>
                  </Grid>
                  <Grid item xs={6}>
                    <Paper p={1} sx={{ p: 1, bgcolor: '#0a192f', border: '1px solid #233554' }}>
                      <Typography variant="caption" color="textSecondary">Battery</Typography>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>92%</Typography>
                    </Paper>
                  </Grid>
                  <Grid item xs={6}>
                    <Paper p={1} sx={{ p: 1, bgcolor: '#0a192f', border: '1px solid #233554' }}>
                      <Typography variant="caption" color="textSecondary">Ambient Temp</Typography>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
                        {weatherInfo ? `${weatherInfo.temp.toFixed(1)} °C` : '31.2 °C'}
                      </Typography>
                    </Paper>
                  </Grid>
                </Grid>
              </Box>

              <Box p={2} sx={{ borderTop: '1px solid #233554', mt: 'auto' }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1.5 }}>
                  EMERGENCY MANUAL OVERRIDES
                </Typography>
                <Grid container spacing={1}>
                  <Grid item xs={12}>
                    <Button 
                      fullWidth 
                      variant="contained" 
                      color="warning" 
                      startIcon={<PauseIcon />}
                      disabled={role === 'viewer' || (weatherInfo && !weatherInfo.safe) || (plannerSafe === false)}
                      sx={{ textTransform: 'none', py: 1.2 }}
                    >
                      Pause Flight
                    </Button>
                  </Grid>
                  <Grid item xs={12}>
                    <Button 
                      fullWidth 
                      variant="contained" 
                      color="error" 
                      startIcon={<HomeIcon />}
                      disabled={role === 'viewer' || (weatherInfo && !weatherInfo.safe) || (plannerSafe === false)}
                      sx={{ textTransform: 'none', py: 1.2 }}
                    >
                      Return-To-Home (RTH)
                    </Button>
                  </Grid>
                </Grid>
              </Box>
            </Box>
          </Box>
        </Box>
      )}
    </ThemeProvider>
  );
}

export default App;
