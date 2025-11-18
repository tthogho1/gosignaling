# WebRTC Signaling Server (Go)

Go 言語で実装された WebRTC シグナリングサーバです。[singo](https://github.com/tockn/singo)ライブラリを参考に実装されています。

## 機能

- **ルーム管理**: クライアントが接続時にルームを作成し、同じルーム内のユーザ間でプロトコル交換を行います
- **自動リソース管理**: ルームから全ユーザが退出した時に、自動的にルームのリソースを削除します
- **WebSocket ベース**: WebSocket を使用した双方向通信
- **フルメッシュ P2P**: 複数ユーザ間でのフルメッシュ P2P 通信をサポート

## アーキテクチャ

```
gosignaling/
├── main.go              # エントリーポイント
├── server.go            # HTTPサーバ設定
├── handler/
│   └── handler.go       # WebSocket接続とメッセージハンドリング
├── manager/
│   └── room.go          # ルーム管理ロジック
├── model/
│   ├── room.go          # ルームとクライアントのモデル
│   └── message.go       # メッセージ型定義
└── repository/
    ├── room.go          # リポジトリインターフェース
    └── mem/
        └── room.go      # インメモリリポジトリ実装
```

## インストール

### 前提条件

- Go 1.21 以上

### セットアップ

```bash
# リポジトリをクローン（または作成）
cd gosignaling

# 依存関係をインストール
go mod download
```

## 使い方

### サーバの起動

```bash
# デフォルト設定で起動（0.0.0.0:5000）
go run .

# カスタムアドレスとポートで起動
go run . -addr 127.0.0.1 -port 8080
```

### コマンドラインオプション

- `-addr`: サーバのアドレス（デフォルト: `0.0.0.0`）
- `-port`: サーバのポート（デフォルト: `5000`）

### ビルド

```bash
# 実行ファイルをビルド
go build -o signaling-server

# 実行
./signaling-server
```

Windows の場合:

```powershell
go build -o signaling-server.exe
.\signaling-server.exe
```

## WebSocket API

### エンドポイント

- `ws://localhost:5000/connect` - WebSocket 接続エンドポイント

### メッセージタイプ

#### クライアント → サーバ

**1. ルームに参加**

```json
{
  "type": "join",
  "payload": {
    "room_id": "room123"
  }
}
```

**2. SDP Offer を送信**

```json
{
  "type": "offer",
  "payload": {
    "sdp": "v=0\r\no=- ...",
    "client_id": "target_client_id"
  }
}
```

**3. SDP Answer を送信**

```json
{
  "type": "answer",
  "payload": {
    "sdp": "v=0\r\no=- ...",
    "client_id": "target_client_id"
  }
}
```

#### サーバ → クライアント

**1. クライアント ID 通知**

```json
{
  "type": "notify-client-id",
  "payload": {
    "client_id": "unique_client_id"
  }
}
```

**2. 新規クライアント通知**

```json
{
  "type": "new-client",
  "payload": {
    "client_id": "new_client_id"
  }
}
```

**3. クライアント退出通知**

```json
{
  "type": "leave-client",
  "payload": {
    "client_id": "leaving_client_id"
  }
}
```

**4. SDP Offer 受信**

```json
{
  "type": "offer",
  "payload": {
    "client_id": "sender_client_id",
    "sdp": "v=0\r\no=- ..."
  }
}
```

**5. SDP Answer 受信**

```json
{
  "type": "answer",
  "payload": {
    "client_id": "sender_client_id",
    "sdp": "v=0\r\no=- ..."
  }
}
```

## 処理フロー

1. **接続確立**

   - クライアントが WebSocket で `/connect` に接続
   - サーバが一意のクライアント ID を生成して通知

2. **ルーム参加**

   - クライアントが `join` メッセージを送信
   - ルームが存在しない場合は自動作成
   - 既存のクライアントに新規参加を通知

3. **WebRTC 接続確立**

   - 新規参加者が Offer を作成して既存クライアントに送信
   - 既存クライアントが Answer を返信
   - ICE 候補交換（SDP に含まれる）

4. **ルーム退出**
   - クライアント切断時に自動的にルームから削除
   - 他のクライアントに退出を通知
   - ルームが空になった場合、自動的にリソースを削除

## 技術スタック

- **Go**: プログラミング言語
- **gorilla/websocket**: WebSocket 実装
- **rs/xid**: 一意な ID 生成

## ライセンス

MIT License

## 参考

このプロジェクトは [tockn/singo](https://github.com/tockn/singo) を参考に実装されています。
