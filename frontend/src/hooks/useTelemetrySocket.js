import { useEffect, useRef, useState } from 'react';

const HEARTBEAT_INTERVAL_MS = 3000;

/**
 * Owns the telemetry WebSocket connection lifecycle (connect while
 * authenticated, 1 Hz message parsing, client heartbeat ping every 3s per
 * AD-5). Extracted out of App.jsx so the connection isn't tied to a
 * specific component tree position - a retro action item from epic 3
 * flagged the inline version as hard to reuse if routing is ever added.
 */
export default function useTelemetrySocket(isAuthenticated) {
  const [telemetry, setTelemetry] = useState(null);
  const [leaseholder, setLeaseholder] = useState("");
  const [hubDoors, setHubDoors] = useState("closed");
  const wsRef = useRef(null);

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
        setLeaseholder(data.leaseholder || "");
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

    const heartbeatInterval = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping" }));
      }
    }, HEARTBEAT_INTERVAL_MS);

    return () => {
      clearInterval(heartbeatInterval);
      ws.close();
      wsRef.current = null;
    };
  }, [isAuthenticated]);

  const resetTelemetry = () => setTelemetry(null);

  return { telemetry, leaseholder, hubDoors, resetTelemetry };
}
