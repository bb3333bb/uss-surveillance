import React, { useState, useEffect, useRef } from 'react';
import { 
  ThemeProvider, CssBaseline, Button, Box, Typography, Paper, Avatar, 
  Tabs, Tab, Grid, AppBar, Toolbar, Chip, IconButton,
  Dialog, DialogTitle, DialogContent, DialogActions, Stepper, Step, StepLabel,
  LinearProgress, Slider
} from '@mui/material';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import FlightTakeoffIcon from '@mui/icons-material/FlightTakeoff';
import LogoutIcon from '@mui/icons-material/Logout';
import MenuOpenIcon from '@mui/icons-material/MenuOpen';
import VideocamIcon from '@mui/icons-material/Videocam';
import PauseIcon from '@mui/icons-material/Pause';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import HomeIcon from '@mui/icons-material/Home';
import LayersIcon from '@mui/icons-material/Layers';
import GestureIcon from '@mui/icons-material/Gesture';
import DeleteIcon from '@mui/icons-material/Delete';
import WarningIcon from '@mui/icons-material/Warning';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';

import L from 'leaflet';
import axios from 'axios';
import { useAuth } from './context/AuthContext.jsx';
import theme from './theme.js';
import VideoPlayer from './components/VideoPlayer.jsx';

function App() {
  const { isAuthenticated, user, role, login, logout, loading } = useAuth();
  const [tabVal, setTabVal] = useState(0);
  const [leftOpen, setLeftOpen] = useState(true);
  const [rightOpen, setRightOpen] = useState(true);

  // Map & drawing refs
  const mapRef = useRef(null);
  const leafletMapRef = useRef(null);
  const flightPathLayerRef = useRef(null);
  const droneMarkerRef = useRef(null);
  const wsRef = useRef(null);

  const [isDrawing, setIsDrawing] = useState(false);
  const [polygonPoints, setPolygonPoints] = useState([]);
  const [weatherInfo, setWeatherInfo] = useState(null);
  const [suggestion, setSuggestion] = useState(null);
  const [flightPath, setFlightPath] = useState(null);
  const [plannerSafe, setPlannerSafe] = useState(null);
  const [plannerMessage, setPlannerMessage] = useState(null);

  // Telemetry & Control mutex leases
  const [telemetry, setTelemetry] = useState(null);
  const [leaseholder, setLeaseholder] = useState("");
  const [hubDoors, setHubDoors] = useState("closed");

  // Historical Replay Mode states
  const [missions, setMissions] = useState([]);
  const [replayMode, setReplayMode] = useState(false);
  const [activeReplay, setActiveReplay] = useState(null);
  const [replayIndex, setReplayIndex] = useState(0);
  const [isReplayPlaying, setIsReplayPlaying] = useState(false);

  // Takeoff / Launch states
  const [showLaunchModal, setShowLaunchModal] = useState(false);
  const [launching, setLaunching] = useState(false);
  const [launchStep, setLaunchStep] = useState(0);
  const [launchMessage, setLaunchMessage] = useState("");

  // Long-press manual override controls (AD-4: requires 1.5s hold)
  const [activePress, setActivePress] = useState(null);
  const [pressProgress, setPressProgress] = useState(0);
  const pressTimerRef = useRef(null);
  const pressIntervalRef = useRef(null);

  const dispatchOverrideCommand = async (cmd) => {
    try {
      const response = await axios.post('http://localhost:8080/api/operator/command', {
        command: cmd
      });
      alert(response.data.message);
    } catch (err) {
      const errMsg = err.response?.data?.message || `Failed to dispatch ${cmd} override command.`;
      alert(errMsg);
    }
  };

  const startHoldGesture = (type) => {
    setActivePress(type);
    setPressProgress(0);

    const startTime = Date.now();
    const duration = 1500;

    pressIntervalRef.current = setInterval(() => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min((elapsed / duration) * 100, 100);
      setPressProgress(Math.floor(progress));
    }, 30);

    pressTimerRef.current = setTimeout(() => {
      clearInterval(pressIntervalRef.current);
      setActivePress(null);
      setPressProgress(0);
      dispatchOverrideCommand(type);
    }, duration);
  };

  const cancelHoldGesture = () => {
    if (pressTimerRef.current) {
      clearTimeout(pressTimerRef.current);
      pressTimerRef.current = null;
    }
    if (pressIntervalRef.current) {
      clearInterval(pressIntervalRef.current);
      pressIntervalRef.current = null;
    }
    setActivePress(null);
    setPressProgress(0);
  };

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

  // Telemetry WebSocket lifecycle connection
  useEffect(() => {
    if (!isAuthenticated) {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      return;
    }

    const token = sessionStorage.getItem('jwt');
    const ws = new WebSocket(`ws://localhost:8080/api/operator/telemetry?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log("Telemetry WebSocket connection opened.");
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setTelemetry(data);
        if (data.leaseholder) {
          setLeaseholder(data.leaseholder);
        } else {
          setLeaseholder("");
        }
        if (data.hub_doors) {
          setHubDoors(data.hub_doors);
        }
      } catch (err) {
        console.error("Failed to parse telemetry event payload:", err);
      }
    };

    ws.onclose = () => {
      console.log("Telemetry WebSocket connection closed.");
    };

    // Client heartbeat watch interval (AD-5: Sends client ping packets every 3s)
    const heartbeatInterval = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping" }));
      }
    }, 3000);

    return () => {
      clearInterval(heartbeatInterval);
      ws.close();
      wsRef.current = null;
    };
  }, [isAuthenticated]);

  // Render active or historical drone positions dynamically on Leaflet map
  useEffect(() => {
    const map = leafletMapRef.current;
    if (!map) return;

    if (droneMarkerRef.current) {
      map.removeLayer(droneMarkerRef.current);
      droneMarkerRef.current = null;
    }

    if (!replayMode && telemetry && telemetry.is_flying) {
      const marker = L.circleMarker([telemetry.lat, telemetry.lng], {
        radius: 8,
        color: '#ff9100', // Neon Orange active drone
        fillColor: '#0a192f',
        fillOpacity: 1,
        weight: 3
      }).bindTooltip("Drone-01 (M300 RTK) - ACTIVE").addTo(map);

      droneMarkerRef.current = marker;
    } else if (replayMode && activeReplay && activeReplay.telemetry && activeReplay.telemetry.length > replayIndex) {
      const pt = activeReplay.telemetry[replayIndex];
      const marker = L.circleMarker([pt.lat, pt.lng], {
        radius: 8,
        color: '#00e5ff', // Neon Cyan historical replay drone
        fillColor: '#0a192f',
        fillOpacity: 1,
        weight: 3
      }).bindTooltip(`Drone-01 (Replay) - Point ${replayIndex}`).addTo(map);

      droneMarkerRef.current = marker;
    }
  }, [telemetry, replayMode, activeReplay, replayIndex]);

  // Fetch completed missions when the History tab (tab index 1) is active
  const fetchMissions = async () => {
    try {
      const token = sessionStorage.getItem('jwt');
      const res = await fetch('http://localhost:8080/api/operator/missions', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (res.ok) {
        const data = await res.json();
        setMissions(data);
      }
    } catch (err) {
      console.error("Failed to fetch historical missions:", err);
    }
  };

  useEffect(() => {
    if (tabVal === 1 && isAuthenticated) {
      fetchMissions();
    }
  }, [tabVal, isAuthenticated]);

  // Map historical replay coordinates to map flightPath line layer
  useEffect(() => {
    if (replayMode && activeReplay) {
      const historicalPath = activeReplay.telemetry.map(pt => ({ lat: pt.lat, lng: pt.lng }));
      setFlightPath(historicalPath);
    } else {
      setFlightPath(null);
    }
  }, [replayMode, activeReplay]);

  // Reset playback parameters on replay target changes
  useEffect(() => {
    setReplayIndex(0);
    setIsReplayPlaying(false);
  }, [activeReplay, replayMode]);

  // Historical replay playback timer ticker (Task 4)
  useEffect(() => {
    if (!replayMode || !isReplayPlaying || !activeReplay) {
      return;
    }
    const maxLen = activeReplay.telemetry ? activeReplay.telemetry.length : 0;
    if (maxLen <= 1) return;

    const interval = setInterval(() => {
      setReplayIndex((prev) => {
        if (prev >= maxLen - 1) {
          setIsReplayPlaying(false);
          return prev;
        }
        return prev + 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [replayMode, isReplayPlaying, activeReplay]);

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
      
      if (response.data.safe) {
        fetchSuggestion({ lat, lng });
        fetchFlightPlan(vertices);
      }
    } catch (err) {
      console.error("Failed to query weather alerts:", err);
    }
  };

  // Dispatch launch command over interlocks
  const triggerDroneLaunch = async () => {
    setShowLaunchModal(false);
    setLaunching(true);
    setLaunchStep(0);
    setLaunchMessage("Initiating automated launch protocol...");

    const t1 = setTimeout(() => {
      setLaunchStep(1);
      setLaunchMessage("Commanding Drone Hub cover doors to open...");
    }, 600);

    const t2 = setTimeout(() => {
      setLaunchStep(2);
      setLaunchMessage("Doors open confirmed. Verifying landing gear interlocks...");
    }, 1500);

    const t3 = setTimeout(() => {
      setLaunchStep(3);
      setLaunchMessage("Ignition sequence started. Hovering takeoff...");
    }, 2400);

    try {
      const response = await axios.post('http://localhost:8080/api/operator/launch');
      clearTimeout(t1);
      clearTimeout(t2);
      clearTimeout(t3);
      setLaunchStep(4);
      setLaunchMessage(response.data.message);
    } catch (err) {
      clearTimeout(t1);
      clearTimeout(t2);
      clearTimeout(t3);
      setLaunching(false);
      setLaunchStep(0);
      const errMsg = err.response?.data?.message || "Launch sequence failed due to mechanical interlock timeout.";
      alert(errMsg);
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

    setTimeout(() => {
      map.invalidateSize();
    }, 250);

    // Setup drawing listeners
    const handleMapClick = (e) => {
      const state = drawStateRef.current;
      if (!state.isDrawing) return;

      const latlng = e.latlng;
      const vertex = [latlng.lat, latlng.lng];
      state.vertices.push(vertex);

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
    if (droneMarkerRef.current) {
      map.removeLayer(droneMarkerRef.current);
      droneMarkerRef.current = null;
    }
    state.vertices = [];
    setPolygonPoints([]);
    setWeatherInfo(null);
    setSuggestion(null);
    setFlightPath(null);
    setPlannerSafe(null);
    setPlannerMessage(null);
    setLaunching(false);
    setLaunchStep(0);
    setTelemetry(null);
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

  // Lock status variables for exclusive lease warnings
  const isLeaseLocked = leaseholder && leaseholder !== user;

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
                {/* Control lease notification banner */}
                {leaseholder && (
                  <Chip 
                    label={isLeaseLocked ? `MUTEX LOCKED: ${leaseholder}` : "MUTEX LEASE: Active (You)"}
                    color={isLeaseLocked ? "warning" : "info"}
                    variant="outlined"
                    sx={{ mr: 2, fontSize: '0.7rem', fontWeight: 600 }}
                  />
                )}

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
                      <Typography variant="caption" color={hubDoors === 'recharging' ? "info.main" : "success.main"} sx={{ display: 'block', fontWeight: 600 }}>
                        ● Online ({
                          hubDoors === 'recharging' ? 'Closed & Recharging' :
                          hubDoors === 'opening' ? 'Cover Opening...' :
                          hubDoors === 'open' ? 'Cover Open' :
                          hubDoors === 'closing' ? 'Cover Closing...' : 'Doors Closed'
                        })
                      </Typography>
                      
                      {hubDoors === 'recharging' ? (
                        <Box sx={{ mt: 1, display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Box sx={{ width: '100%' }}>
                            <LinearProgress 
                              variant="determinate" 
                              value={telemetry ? telemetry.battery : 92} 
                              sx={{ 
                                height: 6, 
                                borderRadius: 3, 
                                bgcolor: 'rgba(0, 229, 255, 0.1)', 
                                '& .MuiLinearProgress-bar': {
                                  bgcolor: '#00e5ff',
                                  animation: 'pulse 1.2s infinite'
                                } 
                              }} 
                            />
                          </Box>
                          <Typography variant="caption" sx={{ color: '#00e5ff', fontFamily: 'monospace', fontWeight: 700 }}>
                            ⚡{telemetry ? telemetry.battery.toFixed(0) : "92"}%
                          </Typography>
                        </Box>
                      ) : (
                        <Typography variant="caption" color="textSecondary">Temp: 24.5°C | Charging Contact: Active</Typography>
                      )}
                    </Paper>

                    <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
                      DRONE FLEET
                    </Typography>
                    <Paper sx={{ p: 1.5, mb: 2, border: '1px solid #233554', bgcolor: '#0a192f' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600 }}>Drone-01 (M300 RTK)</Typography>
                      <Typography variant="caption" color="success.main" sx={{ display: 'block' }}>● {telemetry && telemetry.is_flying ? "Flying" : "Docked"}</Typography>
                      <Typography variant="caption" color="textSecondary">Battery: {telemetry ? telemetry.battery.toFixed(0) : "92"}% | GPS: 3D Fix</Typography>
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
                        <Paper sx={{ p: 1.5, border: '1px solid #00e5ff', bgcolor: '#061325', mb: 2 }}>
                          <Typography variant="body2" sx={{ fontWeight: 600, color: '#00e5ff' }}>
                            {suggestion.recommended_drone}
                          </Typography>
                          <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mt: 0.5 }}>
                            Launch Dock: <strong>{suggestion.recommended_dock}</strong>
                          </Typography>
                          <Typography variant="caption" color="textSecondary" sx={{ display: 'block' }}>
                            Distance to Centroid: {suggestion.distance_meters.toFixed(1)} m
                          </Typography>

                          {/* Confirm Mission Action Trigger */}
                          {plannerSafe && role !== 'viewer' && !isLeaseLocked && !telemetry?.is_flying && (
                            <Button 
                              fullWidth
                              variant="contained"
                              color="success"
                              startIcon={<FlightTakeoffIcon />}
                              onClick={() => setShowLaunchModal(true)}
                              sx={{ mt: 2, fontWeight: 600, py: 1 }}
                            >
                              Confirm Mission
                            </Button>
                          )}

                          {isLeaseLocked && (
                            <Typography variant="caption" color="warning.main" sx={{ display: 'block', mt: 2, fontWeight: 600, textAlign: 'center' }}>
                              ⚠️ Control locked by leaseholder: {leaseholder}
                            </Typography>
                          )}
                        </Paper>
                      </>
                    )}
                  </Box>
                ) : (
                  <Box>
                    <Typography variant="subtitle2" color="textSecondary" sx={{ mb: 1.5, fontWeight: 600 }}>
                      COMPLETED MISSIONS
                    </Typography>
                    {missions.length === 0 ? (
                      <Typography variant="body2" color="textSecondary">
                        No logs found. Run a mission to generate replay data.
                      </Typography>
                    ) : (
                      missions.map((mission) => {
                        const m = Math.floor(mission.duration_sec / 60);
                        const s = Math.floor(mission.duration_sec % 60);
                        const durationStr = `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
                        const dateStr = new Date(mission.start_time).toLocaleString();
                        const isReplayingThis = replayMode && activeReplay?.id === mission.id;

                        return (
                          <Paper 
                            key={mission.id} 
                            sx={{ 
                              p: 1.5, 
                              mb: 2, 
                              border: isReplayingThis ? '1.5px solid #00e5ff' : '1px solid #233554', 
                              bgcolor: '#0a192f',
                              transition: 'all 0.3s'
                            }}
                          >
                            <Typography variant="body2" sx={{ fontWeight: 600, color: isReplayingThis ? '#00e5ff' : 'text.primary' }}>
                              {mission.id}
                            </Typography>
                            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mt: 0.5 }}>
                              Date: {dateStr}
                            </Typography>
                            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 1.5 }}>
                              Duration: {durationStr} | Points: {mission.telemetry ? mission.telemetry.length : 0}
                            </Typography>
                            
                            {isReplayingThis ? (
                              <Button 
                                fullWidth
                                size="small"
                                variant="contained"
                                color="error"
                                onClick={() => {
                                  setReplayMode(false);
                                  setActiveReplay(null);
                                }}
                                sx={{ textTransform: 'none' }}
                              >
                                Stop Replay
                              </Button>
                            ) : (
                              <Button 
                                fullWidth
                                size="small"
                                variant="outlined"
                                color="info"
                                onClick={() => {
                                  setReplayMode(true);
                                  setActiveReplay(mission);
                                }}
                                sx={{ textTransform: 'none', borderColor: '#00e5ff', color: '#00e5ff', '&:hover': { borderColor: '#00b0ff' } }}
                              >
                                View Replay
                              </Button>
                            )}
                          </Paper>
                        );
                      })
                    )}
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
                  {role !== 'viewer' && !isLeaseLocked && !telemetry?.is_flying && (
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

                  {role !== 'viewer' && !isLeaseLocked && (
                    <Button variant="outlined" size="small" color="primary" startIcon={<LayersIcon />} sx={{ textTransform: 'none' }}>
                      Toggle 3D View
                    </Button>
                  )}
                </Box>
              </Box>

              {/* Leaflet Map DOM Canvas */}
              <Box 
                flexGrow={1} 
                position="relative" 
                sx={{ 
                  width: '100%', 
                  height: '100%',
                  minHeight: 0,
                  border: telemetry?.alerts?.length > 0 ? '2px solid #ff3d00' : 'none',
                  boxShadow: telemetry?.alerts?.length > 0 ? '0 0 15px rgba(255, 61, 0, 0.4)' : 'none',
                  transition: 'all 0.3s'
                }}
              >
                <div ref={mapRef} style={{ position: 'absolute', top: 0, bottom: 0, left: 0, right: 0, outline: 'none' }} />
                
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

                {/* Historical Replay Mode Banner Overlay */}
                {replayMode && (
                  <Box 
                    position="absolute" 
                    top={12} 
                    left="50%" 
                    sx={{ 
                      transform: 'translateX(-50%)', 
                      bgcolor: 'rgba(10, 25, 47, 0.95)', 
                      border: '1.5px solid #00e5ff', 
                      boxShadow: '0 0 15px rgba(0, 229, 255, 0.4)',
                      borderRadius: 1, 
                      px: 2, 
                      py: 1, 
                      zIndex: 1000, 
                      display: 'flex', 
                      alignItems: 'center', 
                      gap: 2
                    }}
                  >
                    <Typography variant="body2" sx={{ color: '#00e5ff', fontWeight: 600, display: 'flex', alignItems: 'center', gap: 1 }}>
                      <span>⚠️</span> VIEWING HISTORICAL REPLAY: {activeReplay?.id}
                    </Typography>
                    <Button 
                      variant="contained" 
                      size="small" 
                      color="error" 
                      onClick={() => {
                        setReplayMode(false);
                        setActiveReplay(null);
                      }}
                      sx={{ textTransform: 'none', height: 28 }}
                    >
                      Exit Replay
                    </Button>
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

                {/* Historical Replay Bottom Scrubber & Playback Controls (Task 1 & 4) */}
                {replayMode && activeReplay && activeReplay.telemetry && (
                  <Box 
                    position="absolute" 
                    bottom={20} 
                    left="50%" 
                    sx={{ 
                      transform: 'translateX(-50%)', 
                      width: '85%', 
                      bgcolor: 'rgba(10, 25, 47, 0.95)', 
                      border: '1.5px solid #233554', 
                      borderRadius: 2, 
                      px: 3, 
                      py: 1.5, 
                      zIndex: 1000, 
                      display: 'flex', 
                      alignItems: 'center', 
                      gap: 2,
                      boxShadow: '0 4px 20px rgba(0, 0, 0, 0.5)'
                    }}
                  >
                    <IconButton 
                      onClick={() => setIsReplayPlaying(!isReplayPlaying)} 
                      color="primary"
                      sx={{ bgcolor: 'rgba(0, 229, 255, 0.1)', '&:hover': { bgcolor: 'rgba(0, 229, 255, 0.2)' } }}
                    >
                      {isReplayPlaying ? <PauseIcon sx={{ color: '#00e5ff' }} /> : <PlayArrowIcon sx={{ color: '#00e5ff' }} />}
                    </IconButton>

                    <Box sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column' }}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                        <Typography variant="caption" sx={{ color: 'text.secondary', fontWeight: 600 }}>
                          Timeline Playout
                        </Typography>
                        <Typography variant="caption" sx={{ color: '#00e5ff', fontFamily: 'monospace', fontWeight: 700 }}>
                          Point {replayIndex + 1} / {activeReplay.telemetry.length}
                        </Typography>
                      </Box>
                      <Slider 
                        value={replayIndex} 
                        min={0} 
                        max={activeReplay.telemetry.length - 1} 
                        onChange={(e, val) => setReplayIndex(val)} 
                        sx={{ 
                          color: '#00e5ff', 
                          height: 6,
                          '& .MuiSlider-thumb': {
                            width: 14,
                            height: 14,
                            backgroundColor: '#0a192f',
                            border: '2px solid #00e5ff',
                            '&:focus, &:hover, &.Mui-active, &.Mui-focusVisible': {
                              boxShadow: 'inherit',
                            },
                          },
                          '& .MuiSlider-track': {
                            border: 'none',
                          },
                          '& .MuiSlider-rail': {
                            opacity: 0.28,
                            backgroundColor: '#233554',
                          },
                        }} 
                      />
                    </Box>
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
                <VideoPlayer 
                  isFlying={replayMode || telemetry?.is_flying} 
                  telemetry={replayMode && activeReplay && activeReplay.telemetry ? activeReplay.telemetry[replayIndex] : telemetry} 
                  alerts={replayMode ? [] : telemetry?.alerts} 
                  replayMode={replayMode}
                  replayIndex={replayIndex}
                  totalPoints={activeReplay?.telemetry?.length || 0}
                />
              </Box>

              <Box p={2} sx={{ borderTop: '1px solid #233554' }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1.5 }}>
                  DRONE TELEMETRY
                </Typography>
                <Grid container spacing={1}>
                  <Grid item xs={6}>
                    <Paper p={1} sx={{ p: 1, bgcolor: '#0a192f', border: '1px solid #233554' }}>
                      <Typography variant="caption" color="textSecondary">Altitude</Typography>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
                        {telemetry ? `${telemetry.altitude.toFixed(1)} m` : "0.0 m"}
                      </Typography>
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
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
                        {telemetry ? `${telemetry.battery.toFixed(0)}%` : "92%"}
                      </Typography>
                    </Paper>
                  </Grid>
                  <Grid item xs={6}>
                    <Paper p={1} sx={{ p: 1, bgcolor: '#0a192f', border: '1px solid #233554' }}>
                      <Typography variant="caption" color="textSecondary">Speed</Typography>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
                        {telemetry ? `${telemetry.speed.toFixed(1)} m/s` : "0.0 m/s"}
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
                      disabled={role === 'viewer' || (weatherInfo && !weatherInfo.safe) || (plannerSafe === false) || isLeaseLocked}
                      onMouseDown={() => startHoldGesture("pause")}
                      onMouseUp={cancelHoldGesture}
                      onMouseLeave={cancelHoldGesture}
                      onTouchStart={() => startHoldGesture("pause")}
                      onTouchEnd={cancelHoldGesture}
                      sx={{ textTransform: 'none', py: 1.2, position: 'relative', overflow: 'hidden' }}
                    >
                      {activePress === 'pause' ? `Hold Pause... ${pressProgress}%` : "Pause Flight"}
                      {activePress === 'pause' && (
                        <LinearProgress 
                          variant="determinate" 
                          value={pressProgress} 
                          color="inherit" 
                          sx={{ position: 'absolute', bottom: 0, left: 0, width: '100%', height: 4, opacity: 0.8 }} 
                        />
                      )}
                    </Button>
                  </Grid>
                  <Grid item xs={12}>
                    <Button 
                      fullWidth 
                      variant="contained" 
                      color="error" 
                      startIcon={<HomeIcon />}
                      disabled={role === 'viewer' || (weatherInfo && !weatherInfo.safe) || (plannerSafe === false) || isLeaseLocked}
                      onMouseDown={() => startHoldGesture("rth")}
                      onMouseUp={cancelHoldGesture}
                      onMouseLeave={cancelHoldGesture}
                      onTouchStart={() => startHoldGesture("rth")}
                      onTouchEnd={cancelHoldGesture}
                      sx={{ textTransform: 'none', py: 1.2, position: 'relative', overflow: 'hidden' }}
                    >
                      {activePress === 'rth' ? `Hold RTH... ${pressProgress}%` : "Return-To-Home (RTH)"}
                      {activePress === 'rth' && (
                        <LinearProgress 
                          variant="determinate" 
                          value={pressProgress} 
                          color="inherit" 
                          sx={{ position: 'absolute', bottom: 0, left: 0, width: '100%', height: 4, opacity: 0.8 }} 
                        />
                      )}
                    </Button>
                  </Grid>
                </Grid>
              </Box>
            </Box>
          </Box>
        </Box>
      )}

      {/* Confirmation Mission Modal overlay */}
      <Dialog 
        open={showLaunchModal} 
        onClose={() => setShowLaunchModal(false)}
        PaperProps={{
          sx: {
            bgcolor: 'background.paper',
            border: '1px solid #233554',
            backgroundImage: 'none',
            maxWidth: 450,
            borderRadius: 2
          }
        }}
      >
        <DialogTitle sx={{ fontWeight: 600, color: 'text.primary', borderBottom: '1px solid #233554', py: 2 }}>
          LAUNCH MISSION CONFIRMATION
        </DialogTitle>
        <DialogContent sx={{ py: 3 }}>
          <Typography variant="body2" sx={{ color: 'text.secondary', mb: 2 }}>
            Confirming launch will trigger Drone Hub mechanical door releases and coordinate flight plans over local MQTT channels.
          </Typography>
          <Paper sx={{ p: 2, bgcolor: '#0a192f', border: '1px solid #233554' }}>
            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 0.5 }}>
              Assigned Drone: <strong>Drone-01 (M300 RTK)</strong>
            </Typography>
            <Typography variant="caption" color="textSecondary" sx={{ display: 'block', mb: 0.5 }}>
              Launch Dock: <strong>Dock Alpha</strong>
            </Typography>
            <Typography variant="caption" color="textSecondary" sx={{ display: 'block' }}>
              Flight Boundary: <strong>{polygonPoints.length} WGS84 vertices mapped</strong>
            </Typography>
          </Paper>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3, pt: 0, justifyContent: 'space-between' }}>
          <Button 
            variant="outlined" 
            color="secondary" 
            onClick={() => setShowLaunchModal(false)}
            sx={{ textTransform: 'none' }}
          >
            Cancel
          </Button>
          <Button 
            variant="contained" 
            color="success" 
            startIcon={<FlightTakeoffIcon />}
            onClick={triggerDroneLaunch}
            sx={{ textTransform: 'none', fontWeight: 600 }}
          >
            Confirm & Launch
          </Button>
        </DialogActions>
      </Dialog>

      {/* Takeoff Interlock Progress Stepper overlay */}
      <Dialog 
        open={launching} 
        PaperProps={{
          sx: {
            bgcolor: '#0a192f',
            border: '1px solid #00e5ff',
            backgroundImage: 'none',
            minWidth: 420,
            borderRadius: 2
          }
        }}
      >
        <DialogContent sx={{ py: 4, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
          {launchStep === 4 ? (
            <CheckCircleIcon sx={{ fontSize: 60, color: 'success.main', mb: 2 }} />
          ) : (
            <Box 
              sx={{ 
                width: 50, 
                height: 50, 
                borderRadius: '50%', 
                border: '3px solid #00e5ff', 
                borderTopColor: 'transparent',
                animation: 'spin 1s linear infinite',
                mb: 2
              }} 
            />
          )}

          <Typography variant="h6" color={launchStep === 4 ? "success.main" : "text.primary"} sx={{ fontWeight: 600, mb: 1, textAlign: 'center' }}>
            {launchStep === 4 ? "LAUNCH SUCCESSFUL!" : "DISPATCHING LAUNCH SEQUENCE"}
          </Typography>
          <Typography variant="caption" color="textSecondary" sx={{ mb: 4, display: 'block', textAlign: 'center', px: 2 }}>
            {launchMessage}
          </Typography>

          <Box sx={{ width: '100%', px: 2 }}>
            <Stepper activeStep={launchStep} alternativeLabel sx={{ '& .MuiStepLabel-label': { fontSize: '0.7rem' } }}>
              <Step>
                <StepLabel>Open Hub</StepLabel>
              </Step>
              <Step>
                <StepLabel>Doors Open</StepLabel>
              </Step>
              <Step>
                <StepLabel>Interlocks</StepLabel>
              </Step>
              <Step>
                <StepLabel>Takeoff</StepLabel>
              </Step>
            </Stepper>
          </Box>

          {launchStep === 4 && (
            <Button 
              variant="contained" 
              color="primary" 
              onClick={() => setLaunching(false)}
              sx={{ mt: 4, px: 4, textTransform: 'none', fontWeight: 600 }}
            >
              Close
            </Button>
          )}
        </DialogContent>
      </Dialog>

      {/* Append keyframes style block directly */}
      <style>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </ThemeProvider>
  );
}

export default App;
