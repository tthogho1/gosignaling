# WebRTC WASM ビルドスクリプト

Write-Host "Building WebRTC WASM module..." -ForegroundColor Green

# webrtc-wasmディレクトリに移動
Set-Location -Path (Join-Path $PSScriptRoot "webrtc-wasm")

# wasm-packがインストールされているか確認
if (!(Get-Command wasm-pack -ErrorAction SilentlyContinue)) {
    Write-Host "wasm-pack not found. Installing..." -ForegroundColor Yellow
    cargo install wasm-pack
}

# ビルド
Write-Host "Running wasm-pack build..." -ForegroundColor Cyan
wasm-pack build --target web --release

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild successful!" -ForegroundColor Green
    Write-Host "Generated files in: webrtc-wasm/pkg/" -ForegroundColor Cyan
    Write-Host "`nTo test, run:" -ForegroundColor Yellow
    Write-Host "  python -m http.server 8080" -ForegroundColor White
    Write-Host "Then open: http://localhost:8080/rustwasm.html" -ForegroundColor White
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}

# 元のディレクトリに戻る
Set-Location -Path $PSScriptRoot
