import React, { useRef, useEffect, useState } from 'react';
import { Box, Typography } from '@mui/material';
import VideocamIcon from '@mui/icons-material/Videocam';
import { SrsRtcPlayer } from '../utils/srs.sdk.js';

export default function VideoPlayer({ isFlying, telemetry, alerts, replayMode, replayIndex, totalPoints }) {
  const videoRef = useRef(null);
  const canvasRef = useRef(null);
  const [fallback, setFallback] = useState(false);

  useEffect(() => {
    if (replayMode) {
      setFallback(true);
      return;
    }

    if (!isFlying) {
      setFallback(false);
      return;
    }

    const player = new SrsRtcPlayer();
    const timer = setTimeout(() => {
      console.warn("SRS media connection timeout. Engaging mock Canvas feed.");
      setFallback(true);
    }, 2000);

    player.play("webrtc://localhost:1985/live/drone-stream")
      .then((pc) => {
        clearTimeout(timer);
        pc.ontrack = (event) => {
          if (videoRef.current) {
            videoRef.current.srcObject = event.streams[0];
          }
        };
      })
      .catch(() => {
        clearTimeout(timer);
        setFallback(true);
      });

    return () => {
      player.close();
      clearTimeout(timer);
    };
  }, [isFlying, replayMode]);

  // HTML5 Canvas HUD drawing loop
  useEffect(() => {
    if (!isFlying || !fallback || !canvasRef.current) return;
    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d');
    let animId;
    let scanlineY = 0;
    let targetX = 80;
    let targetY = 60;
    let dx = 0.6;
    let dy = 0.4;

    const renderHUD = () => {
      const w = canvas.width = canvas.offsetWidth;
      const h = canvas.height = canvas.offsetHeight;

      // Dark background grids
      ctx.fillStyle = '#061325';
      ctx.fillRect(0, 0, w, h);

      // Scanline patterns
      ctx.strokeStyle = 'rgba(0, 229, 255, 0.03)';
      ctx.lineWidth = 1;
      for (let y = 0; y < h; y += 10) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(w, y);
        ctx.stroke();
      }

      // Scrolling neon scanline
      scanlineY = (scanlineY + 1) % h;
      ctx.strokeStyle = 'rgba(0, 229, 255, 0.08)';
      ctx.beginPath();
      ctx.moveTo(0, scanlineY);
      ctx.lineTo(w, scanlineY);
      ctx.stroke();

      // Center crosshair overlay
      const cx = w / 2;
      const cy = h / 2;
      ctx.strokeStyle = '#00e5ff';
      ctx.lineWidth = 1.2;
      ctx.beginPath();
      ctx.moveTo(cx - 25, cy); ctx.lineTo(cx - 8, cy);
      ctx.moveTo(cx + 8, cy); ctx.lineTo(cx + 25, cy);
      ctx.moveTo(cx, cy - 25); ctx.lineTo(cx, cy - 8);
      ctx.moveTo(cx, cy + 8); ctx.lineTo(cx, cy + 25);
      ctx.stroke();

      ctx.beginPath();
      ctx.arc(cx, cy, 6, 0, Math.PI * 2);
      ctx.stroke();

      // Object tracking bounding box
      targetX += dx;
      targetY += dy;
      if (targetX < 40 || targetX > w - 60) dx *= -1;
      if (targetY < 40 || targetY > h - 60) dy *= -1;

      ctx.strokeStyle = '#ff9100'; // Target Acquisition Orange
      ctx.lineWidth = 1.5;
      ctx.strokeRect(targetX, targetY, 36, 36);

      ctx.fillStyle = '#ff9100';
      ctx.font = '7.5px monospace';
      ctx.fillText("TRK: Dock Alpha", targetX, targetY - 6);
      ctx.fillText("SAFE LOCK", targetX + 1, targetY + 45);

      // HUD parameters synced with live WebSocket telemetry variables (or replay data)
      ctx.fillStyle = '#00e5ff';
      ctx.font = '8.5px monospace';
      ctx.fillText(replayMode ? "STREAM: HISTORICAL_REPLAY" : "STREAM: AUTO_NAV", 12, 34);
      ctx.fillText(`ALT: ${telemetry ? telemetry.altitude.toFixed(1) : '12.0'} m`, 12, 46);
      ctx.fillText(`BAT: ${telemetry ? telemetry.battery.toFixed(0) : '98'}%`, 12, 58);

      // Coordinates lower right corner
      const latStr = telemetry ? telemetry.lat.toFixed(7) : '10.7626220';
      const lngStr = telemetry ? telemetry.lng.toFixed(7) : '106.6601720';
      ctx.fillText(`LAT: ${latStr}`, w - 105, h - 25);
      ctx.fillText(`LNG: ${lngStr}`, w - 105, h - 13);

      // Draw historical playback status alert banner (AC-3)
      if (replayMode) {
        ctx.fillStyle = 'rgba(0, 229, 255, 0.05)';
        ctx.fillRect(0, 0, w, h);

        const m = Math.floor(replayIndex / 60);
        const s = Math.floor(replayIndex % 60);
        const timeStr = `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;

        const flashPhase = Math.floor(Date.now() / 400) % 2;
        if (flashPhase === 0) {
          ctx.fillStyle = '#00e5ff';
          ctx.fillRect(0, 0, w, 22);

          ctx.fillStyle = '#0a192f';
          ctx.font = 'bold 8px monospace';
          ctx.fillText(`⚠️ REPLAY PLAYBACK: ${timeStr} (Pt ${replayIndex + 1}/${totalPoints})`, 12, 14);
        }
      }

      // Draw neon red flashing warning status overlay (AC-3)
      const hasAlerts = alerts && alerts.length > 0;
      if (!replayMode && hasAlerts) {
        ctx.fillStyle = 'rgba(255, 61, 0, 0.12)';
        ctx.fillRect(0, 0, w, h);

        const flashPhase = Math.floor(Date.now() / 400) % 2;
        if (flashPhase === 0) {
          ctx.fillStyle = '#ff3d00';
          ctx.fillRect(0, 0, w, 22);

          ctx.fillStyle = '#ffffff';
          ctx.font = 'bold 8px monospace';
          ctx.fillText(`⚠️ WARNING: SAFETY BREACH (${alerts.join(", ")})`, 12, 14);
        }
      }

      animId = requestAnimationFrame(renderHUD);
    };

    renderHUD();
    return () => cancelAnimationFrame(animId);
  }, [isFlying, fallback, telemetry, alerts, replayMode, replayIndex, totalPoints]);

  const hasAlerts = alerts && alerts.length > 0;

  if (!isFlying) {
    return (
      <Box 
        sx={{ 
          height: 200, 
          display: 'flex', 
          flexDirection: 'column', 
          justifyContent: 'center', 
          alignItems: 'center', 
          bgcolor: '#0a192f',
          border: '1px dashed #233554',
          borderRadius: 1
        }}
      >
        <VideocamIcon sx={{ fontSize: 40, color: 'text.secondary', mb: 1 }} />
        <Typography variant="caption" color="textSecondary">WebRTC Stream Playout Idle</Typography>
      </Box>
    );
  }

  return (
    <Box 
      sx={{ 
        position: 'relative', 
        width: '100%', 
        height: 200, 
        bgcolor: '#061325', 
        borderRadius: 1, 
        overflow: 'hidden', 
        border: replayMode ? '2px solid #00e5ff' : (hasAlerts ? '2px solid #ff3d00' : '1px solid #233554'),
        boxShadow: replayMode ? '0 0 15px rgba(0, 229, 255, 0.3)' : (hasAlerts ? '0 0 15px rgba(255, 61, 0, 0.4)' : 'none'),
        transition: 'all 0.3s'
      }}
    >
      {fallback ? (
        <canvas ref={canvasRef} style={{ width: '100%', height: '100%', display: 'block' }} />
      ) : (
        <video ref={videoRef} autoPlay playsInline muted style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
      )}
      
      <Box
        sx={{
          position: 'absolute',
          top: replayMode || hasAlerts ? 30 : 8,
          right: 8,
          bgcolor: 'rgba(10, 25, 47, 0.85)',
          px: 1.2,
          py: 0.35,
          borderRadius: 0.5,
          border: '1px solid #00e5ff',
          pointerEvents: 'none'
        }}
      >
        <Typography variant="caption" sx={{ color: '#00e5ff', fontWeight: 700, fontSize: '0.62rem' }}>
          {replayMode ? "● HISTORICAL REPLAY" : "● LIVE WEBRTC"}
        </Typography>
      </Box>
    </Box>
  );
}
