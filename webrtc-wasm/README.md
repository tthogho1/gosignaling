# WebRTC WASM Client

A WebRTC client implemented in Rust WebAssembly

## Required Tools

```powershell
# Install Rust (if not already installed)
# Install from https://rustup.rs/

# Install wasm-pack
cargo install wasm-pack

# Or
npm install -g wasm-pack
```

## Build Instructions

```powershell
cd webrtc-wasm
wasm-pack build --target web
```

This will generate the following files in the `pkg/` directory:

- `webrtc_wasm.js` - JavaScript bindings
- `webrtc_wasm_bg.wasm` - WebAssembly binary
- `webrtc_wasm.d.ts` - TypeScript type definitions

## Usage

1. Build the WASM module:

```powershell
cd webrtc-wasm
wasm-pack build --target web
```

2. Start an HTTP server (WASM does not work with the file:// protocol):

```powershell
# For Python
python -m http.server 8080

# Or for Node.js
npx http-server -p 8080
```

3. Open `http://localhost:8080/rustwasm.html` in your browser

## Implemented Features

### Rust (WASM) Side

- ✅ WebRTC PeerConnection management
- ✅ Media stream acquisition
- ✅ SDP Offer/Answer generation
- ✅ ICE candidate collection
- ✅ Remote stream reception
- ✅ STUN/TURN server configuration
- ✅ Trickle ICE support

### JavaScript Side

- ✅ WebSocket communication
- ✅ Signaling message routing
- ✅ UI control
- ✅ Video element management

## Architecture

```
┌─────────────────────────────────────────┐
│         rustwasm.html (UI)              │
│  - WebSocket communication              │
│  - Video element management             │
└────────────┬────────────────────────────┘
             │
             │ JavaScript Bridge
             ▼
┌─────────────────────────────────────────┐
│   webrtc-wasm (Rust/WASM)               │
│  - WebRTC PeerConnection                │
│  - Media stream processing              │
│  - SDP exchange logic                   │
└─────────────────────────────────────────┘
```

## Performance

- **Initial Load**: Slight overhead for loading WebAssembly
- **Execution Speed**: Near-native performance
- **Binary Size**: Approximately 100-200KB with optimized build

## Development

### Debug Build

```powershell
wasm-pack build --target web --dev
```

### Release Build (Optimized)

```powershell
wasm-pack build --target web --release
```

### Size Optimization

```powershell
# Further optimize using wasm-opt
wasm-opt pkg/webrtc_wasm_bg.wasm -O3 -o pkg/webrtc_wasm_bg.wasm
```

## Troubleshooting

### CORS Error

WASM files must be served via an HTTP server.

### Build Error

```powershell
# Update dependencies
cargo update

# Clean build
cargo clean
wasm-pack build --target web
```

## Future Improvements

1. **Complete WebSocket Integration** - Manage WebSocket on the Rust side
2. **Data Channels** - Support for file transfers, etc.
3. **Recording Feature** - MediaRecorder API integration
4. **Screen Sharing** - getDisplayMedia API support
5. **Statistics** - WebRTC statistics retrieval and display
