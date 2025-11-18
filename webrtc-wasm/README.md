# WebRTC WASM Client

Rust WebAssembly で実装された WebRTC クライアント

## 必要なツール

```powershell
# Rust のインストール (まだの場合)
# https://rustup.rs/ からインストール

# wasm-pack のインストール
cargo install wasm-pack

# または
npm install -g wasm-pack
```

## ビルド方法

```powershell
cd webrtc-wasm
wasm-pack build --target web
```

これにより `pkg/` ディレクトリに以下のファイルが生成されます:

- `webrtc_wasm.js` - JavaScript バインディング
- `webrtc_wasm_bg.wasm` - WebAssembly バイナリ
- `webrtc_wasm.d.ts` - TypeScript 型定義

## 使用方法

1. WASM モジュールをビルド:

```powershell
cd webrtc-wasm
wasm-pack build --target web
```

2. HTTP サーバーを起動 (WASM は file://プロトコルでは動作しません):

```powershell
# Pythonの場合
python -m http.server 8080

# または Node.jsの場合
npx http-server -p 8080
```

3. ブラウザで `http://localhost:8080/rustwasm.html` を開く

## 実装されている機能

### Rust (WASM) 側

- ✅ WebRTC PeerConnection 管理
- ✅ メディアストリーム取得
- ✅ SDP Offer/Answer 生成
- ✅ ICE 候補収集
- ✅ リモートストリームの受信
- ✅ STUN/TURN サーバー設定
- ✅ Trickle ICE サポート

### JavaScript 側

- ✅ WebSocket 通信
- ✅ シグナリングメッセージのルーティング
- ✅ UI 制御
- ✅ ビデオ要素の管理

## アーキテクチャ

```
┌─────────────────────────────────────────┐
│         rustwasm.html (UI)              │
│  - WebSocket通信                         │
│  - ビデオ要素管理                        │
└────────────┬────────────────────────────┘
             │
             │ JavaScript Bridge
             ▼
┌─────────────────────────────────────────┐
│   webrtc-wasm (Rust/WASM)               │
│  - WebRTC PeerConnection                │
│  - メディアストリーム処理                 │
│  - SDP交換ロジック                       │
└─────────────────────────────────────────┘
```

## パフォーマンス

- **初期ロード**: WebAssembly のロードにわずかなオーバーヘッド
- **実行速度**: ネイティブに近いパフォーマンス
- **バイナリサイズ**: 最適化ビルドで約 100-200KB

## 開発

### デバッグビルド

```powershell
wasm-pack build --target web --dev
```

### リリースビルド (最適化)

```powershell
wasm-pack build --target web --release
```

### サイズ最適化

```powershell
# wasm-optを使用してさらに最適化
wasm-opt pkg/webrtc_wasm_bg.wasm -O3 -o pkg/webrtc_wasm_bg.wasm
```

## トラブルシューティング

### CORS エラー

WASM ファイルは HTTP サーバー経由で提供する必要があります。

### ビルドエラー

```powershell
# 依存関係を更新
cargo update

# クリーンビルド
cargo clean
wasm-pack build --target web
```

## 今後の改善案

1. **完全な WebSocket 統合** - WebSocket も Rust 側で管理
2. **データチャネル** - ファイル転送などのサポート
3. **録画機能** - MediaRecorder API の統合
4. **画面共有** - getDisplayMedia API のサポート
5. **統計情報** - WebRTC 統計の取得と表示
