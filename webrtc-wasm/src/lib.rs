use wasm_bindgen::prelude::*;
use wasm_bindgen_futures::JsFuture;
use web_sys::{
    MediaStream, MediaStreamConstraints, RtcConfiguration, RtcIceServer, RtcPeerConnection,
    RtcSdpType, RtcSessionDescriptionInit, RtcPeerConnectionIceEvent, RtcTrackEvent,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::rc::Rc;
use std::cell::RefCell;
use js_sys::{Object, Reflect, Array};

#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = console)]
    fn log(s: &str);
}

macro_rules! console_log {
    ($($t:tt)*) => (log(&format_args!($($t)*).to_string()))
}

#[derive(Debug, Serialize, Deserialize)]
struct Message {
    #[serde(rename = "type")]
    msg_type: String,
    payload: serde_json::Value,
}

#[derive(Debug, Serialize, Deserialize)]
struct SdpPayload {
    sdp: String,
    client_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct IceCandidatePayload {
    candidate: String,
    #[serde(rename = "sdpMid")]
    sdp_mid: Option<String>,
    #[serde(rename = "sdpMLineIndex")]
    sdp_mline_index: Option<u16>,
    client_id: String,
}

#[wasm_bindgen]
pub struct WebRTCClient {
    local_stream: Option<MediaStream>,
    peer_connections: Rc<RefCell<HashMap<String, RtcPeerConnection>>>,
    on_status_change: Option<js_sys::Function>,
    on_remote_stream: Option<js_sys::Function>,
    on_ice_candidate: Option<js_sys::Function>,
}

#[wasm_bindgen]
impl WebRTCClient {
    #[wasm_bindgen(constructor)]
    pub fn new() -> Self {
        console_log!("[WebRTCClient] Initializing...");
        Self {
            local_stream: None,
            peer_connections: Rc::new(RefCell::new(HashMap::new())),
            on_status_change: None,
            on_remote_stream: None,
            on_ice_candidate: None,
        }
    }

    #[wasm_bindgen(js_name = setOnStatusChange)]
    pub fn set_on_status_change(&mut self, callback: js_sys::Function) {
        self.on_status_change = Some(callback);
    }

    #[wasm_bindgen(js_name = setOnRemoteStream)]
    pub fn set_on_remote_stream(&mut self, callback: js_sys::Function) {
        self.on_remote_stream = Some(callback);
    }

    #[wasm_bindgen(js_name = setOnIceCandidate)]
    pub fn set_on_ice_candidate(&mut self, callback: js_sys::Function) {
        self.on_ice_candidate = Some(callback);
    }

    fn update_status(&self, message: &str) {
        console_log!("[Status] {}", message);
        if let Some(ref callback) = self.on_status_change {
            let _ = callback.call1(&JsValue::NULL, &JsValue::from_str(message));
        }
    }

    #[wasm_bindgen(js_name = getLocalStream)]
    pub async fn get_local_stream(&mut self) -> Result<MediaStream, JsValue> {
        console_log!("[WebRTCClient] Getting local stream...");
        
        let window = web_sys::window().ok_or("No window")?;
        let navigator = window.navigator();
        let media_devices = navigator.media_devices()?;

        let mut constraints = MediaStreamConstraints::new();
        
        // Video constraints
        let video_constraints = Object::new();
        Reflect::set(&video_constraints, &"width".into(), &640.into())?;
        Reflect::set(&video_constraints, &"height".into(), &480.into())?;
        constraints.video(&video_constraints.into());
        
        // Audio constraints
        constraints.audio(&JsValue::TRUE);

        let stream_promise = media_devices.get_user_media_with_constraints(&constraints)?;
        let stream = JsFuture::from(stream_promise).await?;
        let media_stream: MediaStream = stream.dyn_into()?;

        self.local_stream = Some(media_stream.clone());
        self.update_status("カメラとマイクを取得しました");

        Ok(media_stream)
    }

    fn create_rtc_config(&self) -> Result<RtcConfiguration, JsValue> {
        let config = RtcConfiguration::new();
        let ice_servers = Array::new();

        // STUN servers
        let stun_urls = vec![
            "stun:stun.l.google.com:19302",
            "stun:stun1.l.google.com:19302",
            "stun:stun2.l.google.com:19302",
            "stun:stun3.l.google.com:19302",
            "stun:stun4.l.google.com:19302",
        ];

        for url in stun_urls {
            let mut server = RtcIceServer::new();
            let urls = Array::new();
            urls.push(&JsValue::from_str(url));
            server.urls(&urls.into());
            ice_servers.push(&server);
        }

        // TURN servers
        let turn_configs = vec![
            ("turn:openrelay.metered.ca:80", "openrelayproject", "openrelayproject"),
            ("turn:openrelay.metered.ca:443", "openrelayproject", "openrelayproject"),
            ("turn:openrelay.metered.ca:443?transport=tcp", "openrelayproject", "openrelayproject"),
        ];

        for (url, username, credential) in turn_configs {
            let mut server = RtcIceServer::new();
            let urls = Array::new();
            urls.push(&JsValue::from_str(url));
            server.urls(&urls.into());
            server.username(username);
            server.credential(credential);
            ice_servers.push(&server);
        }

        config.set_ice_servers(&ice_servers);
        Ok(config)
    }

    fn create_peer_connection(&self, target_client_id: &str) -> Result<RtcPeerConnection, JsValue> {
        console_log!("[WebRTCClient] Creating peer connection for: {}", target_client_id);
        
        let config = self.create_rtc_config()?;
        let pc = RtcPeerConnection::new_with_configuration(&config)?;

        // Add local tracks
        if let Some(ref stream) = self.local_stream {
            let tracks = stream.get_tracks();
            for i in 0..tracks.length() {
                if let Some(track) = tracks.get(i).dyn_into::<web_sys::MediaStreamTrack>().ok() {
                    // Use add_track with proper parameters via JavaScript interop
                    if let Ok(add_track_fn) = js_sys::Reflect::get(&pc, &"addTrack".into()) {
                        let _ = js_sys::Reflect::apply(
                            &add_track_fn.dyn_into::<js_sys::Function>().unwrap(),
                            &pc,
                            &Array::of2(&track, stream),
                        );
                    }
                }
            }
        }

        // Setup ontrack handler
        let target_id = target_client_id.to_string();
        let on_remote_stream = self.on_remote_stream.clone();
        let ontrack_callback = Closure::wrap(Box::new(move |event: RtcTrackEvent| {
            console_log!("[WebRTCClient] Received track from: {}", target_id);
            let streams = event.streams();
            if streams.length() > 0 {
                if let Some(stream) = streams.get(0).dyn_into::<MediaStream>().ok() {
                    console_log!("[WebRTCClient] Got MediaStream with {} tracks", stream.get_tracks().length());
                    if let Some(ref callback) = on_remote_stream {
                        let _ = callback.call2(&JsValue::NULL, &JsValue::from_str(&target_id), &stream);
                    }
                }
            } else {
                console_log!("[WebRTCClient] No streams in track event");
            }
        }) as Box<dyn FnMut(_)>);
        
        pc.set_ontrack(Some(ontrack_callback.as_ref().unchecked_ref()));
        ontrack_callback.forget();

        // Setup onicecandidate handler
        let target_id_ice = target_client_id.to_string();
        let on_ice_callback = self.on_ice_candidate.clone();
        let onicecandidate_callback = Closure::wrap(Box::new(move |event: RtcPeerConnectionIceEvent| {
            if let Some(candidate) = event.candidate() {
                console_log!("[WebRTCClient] New ICE candidate for: {}", target_id_ice);
                if let Some(ref callback) = on_ice_callback {
                    let candidate_str = candidate.candidate();
                    let sdp_mid = candidate.sdp_mid();
                    let sdp_mline_index = candidate.sdp_m_line_index();
                    
                    let payload = IceCandidatePayload {
                        candidate: candidate_str,
                        sdp_mid,
                        sdp_mline_index,
                        client_id: target_id_ice.clone(),
                    };
                    
                    if let Ok(payload_json) = serde_json::to_value(&payload) {
                        let message = Message {
                            msg_type: "ice-candidate".to_string(),
                            payload: payload_json,
                        };
                        if let Ok(json) = serde_json::to_string(&message) {
                            let _ = callback.call1(&JsValue::NULL, &JsValue::from_str(&json));
                        }
                    }
                }
            } else {
                console_log!("[WebRTCClient] ICE gathering complete for: {}", target_id_ice);
            }
        }) as Box<dyn FnMut(_)>);
        
        pc.set_onicecandidate(Some(onicecandidate_callback.as_ref().unchecked_ref()));
        onicecandidate_callback.forget();

        // Setup onconnectionstatechange
        let target_id2 = target_client_id.to_string();
        let onconnectionstatechange_callback = Closure::wrap(Box::new(move |_event: JsValue| {
            console_log!("[WebRTCClient] Connection state changed for: {}", target_id2);
        }) as Box<dyn FnMut(_)>);
        
        pc.set_onconnectionstatechange(Some(onconnectionstatechange_callback.as_ref().unchecked_ref()));
        onconnectionstatechange_callback.forget();

        self.peer_connections.borrow_mut().insert(target_client_id.to_string(), pc.clone());

        Ok(pc)
    }

    #[wasm_bindgen(js_name = handleNewClient)]
    pub async fn handle_new_client(&self, client_id: String) -> Result<String, JsValue> {
        console_log!("[WebRTCClient] Handling new client: {}", client_id);
        self.update_status(&format!("新しいクライアントが参加: {}", client_id));

        let pc = self.create_peer_connection(&client_id)?;
        console_log!("[WebRTCClient] Peer connection created successfully");

        // Create offer
        console_log!("[WebRTCClient] Creating offer...");
        let offer = JsFuture::from(pc.create_offer()).await
            .map_err(|e| {
                console_log!("[WebRTCClient] Error creating offer");
                e
            })?;
        console_log!("[WebRTCClient] Offer promise resolved");
        
        // Convert JsValue to RtcSessionDescriptionInit
        let offer_init = RtcSessionDescriptionInit::from(offer);
        console_log!("[WebRTCClient] Offer ready to set as local description");
        
        JsFuture::from(pc.set_local_description(&offer_init)).await
            .map_err(|e| {
                console_log!("[WebRTCClient] Error setting local description");
                e
            })?;
        console_log!("[WebRTCClient] Offer created and set as local description");

        // Get SDP from local description
        if let Some(local_desc) = pc.local_description() {
            let sdp_string = local_desc.sdp();
            console_log!("[WebRTCClient] Got SDP from local description, length: {}", sdp_string.len());
            
            let payload = SdpPayload {
                sdp: sdp_string,
                client_id: client_id.clone(),
            };
            let message = Message {
                msg_type: "offer".to_string(),
                payload: serde_json::to_value(payload).unwrap(),
            };
            let json = serde_json::to_string(&message).unwrap();
            console_log!("[WebRTCClient] Returning offer message as JSON string, length: {}", json.len());
            return Ok(json);
        }

        console_log!("[WebRTCClient] Failed to get local description");
        Err(JsValue::from_str("Failed to get local description"))
    }

    #[wasm_bindgen(js_name = handleOffer)]
    pub async fn handle_offer(&self, sender_id: String, sdp: String) -> Result<String, JsValue> {
        console_log!("[WebRTCClient] Handling offer from: {}", sender_id);
        self.update_status(&format!("{} からOfferを受信", sender_id));

        let pc = self.create_peer_connection(&sender_id)?;

        // Set remote description
        let remote_desc = RtcSessionDescriptionInit::new(RtcSdpType::Offer);
        remote_desc.set_sdp(&sdp);
        JsFuture::from(pc.set_remote_description(&remote_desc)).await?;
        console_log!("[WebRTCClient] Remote description set");

        // Create answer
        let answer = JsFuture::from(pc.create_answer()).await?;
        let answer_init = RtcSessionDescriptionInit::from(answer);
        
        JsFuture::from(pc.set_local_description(&answer_init)).await?;
        console_log!("[WebRTCClient] Answer created and set as local description");

        // Get answer from local description
        if let Some(local_desc) = pc.local_description() {
            let sdp_string = local_desc.sdp();
            console_log!("[WebRTCClient] Got SDP from local description");
            
            let payload = SdpPayload {
                sdp: sdp_string,
                client_id: sender_id.clone(),
            };
            let message = Message {
                msg_type: "answer".to_string(),
                payload: serde_json::to_value(payload).unwrap(),
            };
            let json = serde_json::to_string(&message).unwrap();
            console_log!("[WebRTCClient] Returning answer message as JSON string");
            return Ok(json);
        }

        Err(JsValue::from_str("Failed to get local description"))
    }

    #[wasm_bindgen(js_name = handleAnswer)]
    pub async fn handle_answer(&self, sender_id: String, sdp: String) -> Result<(), JsValue> {
        console_log!("[WebRTCClient] Handling answer from: {}", sender_id);
        self.update_status(&format!("{} からAnswerを受信", sender_id));

        if let Some(pc) = self.peer_connections.borrow().get(&sender_id) {
            let remote_desc = RtcSessionDescriptionInit::new(RtcSdpType::Answer);
            remote_desc.set_sdp(&sdp);
            JsFuture::from(pc.set_remote_description(&remote_desc)).await?;
            console_log!("[WebRTCClient] Remote description set");
        }

        Ok(())
    }

    #[wasm_bindgen(js_name = handleIceCandidate)]
    pub async fn handle_ice_candidate(&self, sender_id: String, candidate: String, sdp_mid: Option<String>, sdp_mline_index: Option<u16>) -> Result<(), JsValue> {
        console_log!("[WebRTCClient] Handling ICE candidate from: {}", sender_id);

        if let Some(pc) = self.peer_connections.borrow().get(&sender_id) {
            let mut candidate_init = web_sys::RtcIceCandidateInit::new(&candidate);
            if let Some(mid) = sdp_mid {
                candidate_init.sdp_mid(Some(&mid));
            }
            if let Some(index) = sdp_mline_index {
                candidate_init.sdp_m_line_index(Some(index));
            }
            
            let ice_candidate = web_sys::RtcIceCandidate::new(&candidate_init)?;
            JsFuture::from(pc.add_ice_candidate_with_opt_rtc_ice_candidate(Some(&ice_candidate))).await?;
            console_log!("[WebRTCClient] ICE candidate added");
        }

        Ok(())
    }

    #[wasm_bindgen(js_name = handleLeaveClient)]
    pub fn handle_leave_client(&self, client_id: String) {
        console_log!("[WebRTCClient] Client left: {}", client_id);
        self.update_status(&format!("{} が退出しました", client_id));

        if let Some(pc) = self.peer_connections.borrow_mut().remove(&client_id) {
            pc.close();
        }
    }

    #[wasm_bindgen(js_name = close)]
    pub fn close(&mut self) {
        console_log!("[WebRTCClient] Closing connections...");
        
        // Close all peer connections
        for (_, pc) in self.peer_connections.borrow_mut().drain() {
            pc.close();
        }

        // Stop local stream tracks
        if let Some(ref stream) = self.local_stream {
            let tracks = stream.get_tracks();
            for i in 0..tracks.length() {
                if let Some(track) = tracks.get(i).dyn_into::<web_sys::MediaStreamTrack>().ok() {
                    track.stop();
                }
            }
        }

        self.local_stream = None;
    }
}
