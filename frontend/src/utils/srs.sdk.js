// srs.sdk.js - Lightweight SRS WebRTC SDK for SDP Negotiations

export class SrsRtcPlayer {
  constructor() {
    this.pc = null;
  }

  async play(streamUrl) {
    this.close();

    // Setup standard WebRTC connection with passive transceivers
    this.pc = new RTCPeerConnection();
    this.pc.addTransceiver("audio", { direction: "recvonly" });
    this.pc.addTransceiver("video", { direction: "recvonly" });

    const offer = await this.pc.createOffer();
    await this.pc.setLocalDescription(offer);

    const srsPayload = {
      api: "http://localhost:1985/rtc/v1/play/",
      streamurl: streamUrl,
      sdp: offer.sdp
    };

    // Dispatch SDP offer exchange to SRS Media Server
    const response = await fetch("http://localhost:1985/rtc/v1/play/", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(srsPayload)
    });

    if (response.status !== 200) {
      throw new Error(`SRS media server returned error code ${response.status}`);
    }

    const data = await response.json();
    if (data.code !== 0) {
      throw new Error(`SRS connection refused: ${data.msg}`);
    }

    const answer = new RTCSessionDescription({
      type: "answer",
      sdp: data.sdp
    });

    await this.pc.setRemoteDescription(answer);
    return this.pc;
  }

  close() {
    if (this.pc) {
      this.pc.close();
      this.pc = null;
    }
  }
}
