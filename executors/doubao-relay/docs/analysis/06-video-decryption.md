# 06 · 视频无水印 token 本地解密

## 为什么要这块

豆包返回视频结果时,**普通播放 URL 带水印**。无水印版本有两种给法:

1. **直链**:`url?lr=unwatermarked`(简单模式,直接下载即可)
2. **加密 token**:响应里的 `fallback_url` 是 base64 字符串,配合 `key_seed` 用 AES-128-CBC 解密出真实 URL

**第二种更常见**。豆包用这个做反爬 — 服务端对无水印 URL 加密,只有豆包客户端自己(以及这个工具)知道怎么解。

## 算法总览

```
输入:
  fallback_url (base64 加密 token)
  key_seed (字符串,从页面响应 JSON 的 key_seed 字段提取)

输出:
  真实的无水印视频 URL (https://xxx.volces.com/.../video.mp4)

流程:
  1. base64Decode(key_seed[:32])           → seedBytes
  2. SHA-512(seedBytes)                    → seedHash (64 bytes)

  3. seedHash[0:16] / seedHash[16:32]      → key1 / iv1 (直接对半分)
     seedHash[16:32] / seedHash[0:16]      → key2 / iv2 (交叉)

  4. SHA-512(base64Decode(key_seed))       → seedFull
     SHA-512(seedFull + salt)              → combined
     combined[0:16] / combined[16:32]      → key3 / iv3
     其中 salt 是固定常量 '4dd4c2e6b83162090e52b3c7a6733ba4' 的 hex 解码

  5. payload = base64Decode(fallback_url)  (支持 $@# ↔ _+/= 字符替换)

  6. 如果 payload[0:4] == A8 00 01 00:
       剥掉 4 字节 magic
       for 每个 (key, iv) in [(key3,iv3), (key1,iv1), (key2,iv2), ...]:
           plain = AES_CBC_Decrypt(payload[4:], key, iv)
           如果 plain 是合法的 https URL: return plain
           # 可能嵌套:plain 本身又是 base64 + AES-CBC
           inner = 再解一次
           如果 inner 是合法 URL: return inner
```

## 核心源码还原

### 常量

```js
const SALT_HEX = '4dd4c2e6b83162090e52b3c7a6733ba4';
// 32 字节的额外盐,从预签名的常量提取
const EXTRA_PEPPER = 'TdTC5rgxYgkOUrPHpnM7pByyRiuCmrWKGWs521cXdST0m69/COjWjSanLjfBqVovHwWlGJKu8pSXMrYqOKrdWA==';
```

**这两个常量极可能是从豆包客户端 JS 逆向出来的**。字节升级客户端后,这两个常量可能会变,届时整条链就断。

### 容错 base64 解码(_0x410227)

```js
function base64DecodeFlexible(input) {
    const raw = String(input || '').trim();

    // 豆包可能用三种 base64 变体
    const variants = [
        raw,
        raw.replace(/[$@#]/g, c => ({ '$': '_', '@': '/', '#': '.' }[c])),
        raw.replace(/[$@#]/g, c => ({ '$': '+', '@': '/', '#': '=' }[c]))
    ];

    const tried = new Set();
    for (const candidate of variants) {
        if (!candidate || tried.has(candidate)) continue;
        tried.add(candidate);

        try {
            // 转为标准 base64
            const normalized = padBase64(candidate)
                .replace(/-/g, '+').replace(/_/g, '/');
            const bin = atob(normalized);
            const bytes = new Uint8Array(bin.length);
            for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
            return bytes;
        } catch { }
    }
    return null;
}

function padBase64(s) {
    const padding = (4 - s.length % 4) % 4;
    return s + '='.repeat(padding);
}
```

**三种 base64 变体**:
1. 标准 base64 (+/=)
2. URL-safe base64 (-_, 无填充) → 字符 $@# 替换为 _/.
3. 另一种混淆变体 → $@# 替换为 +/=

**学习点**:对抗爬虫时,token 字符可能被恶意改成特殊字符。工具"穷举"三种可能性,哪个解出来是 printable 字符就用哪个。

### 可打印字符校验(_0x848206)

```js
function bytesToText(bytes) {
    if (!bytes || !bytes.length) return '';
    // 校验:全部字节必须是可打印 ASCII (或 \t \n \r)
    for (const b of bytes) {
        if (b !== 0x09 && b !== 0x0A && b !== 0x0D && (b < 0x20 || b > 0x7E)) return '';
    }
    return new TextDecoder().decode(bytes);
}
```

**判别技巧**:AES 解密如果 key/iv 错,结果是乱码(0x00-0xFF 均匀分布)。如果结果全是可打印 ASCII,说明 key/iv 对了,这就是真实数据。

### key_seed 提取(_0x363561)

key_seed 不一定在固定字段,可能嵌在:
- JSON 字段 `key_seed`
- URL 的 query `?key_seed=...`
- 字符串里内嵌的 `"key_seed":"..."`

```js
function extractKeySeed(value, depth = 0) {
    if (depth > 10 || value == null) return '';

    if (typeof value === 'string') {
        let m = value.match(/(?:^|[?&])key_seed=([^&"'<>\\\s]+)/i);
        if (m) return decodeURIComponent(m[1]);
        m = value.match(/["']key_seed["']\s*:\s*["']([^"']+)/i);
        return m ? decodeURIComponent(m[1]) : '';
    }

    if (typeof value !== 'object') return '';

    // 直接字段
    if (typeof value.key_seed === 'string' && value.key_seed.trim()) {
        return value.key_seed.trim();
    }

    // 递归子对象
    for (const child of Object.values(value)) {
        const found = extractKeySeed(child, depth + 1);
        if (found) return found;
    }
    return '';
}
```

## 主解密函数

```js
async function decryptToken(tokenBase64, keySeed = '') {
    // 快速路径:token 本身就是合法 URL(未加密)
    if (isValidHttpsUrl(tokenBase64)) return tokenBase64;

    // 路径 1: 仅用 token 解
    const simple = decodeWithPassthrough(tokenBase64);
    if (simple) return simple;

    // 路径 2: token 以 'qAAB' 开头(magic + 空 padding) + keySeed → 完整解密
    if (tokenBase64.startsWith('qAAB') && keySeed) {
        return await decryptWithKeySeed(tokenBase64, keySeed);
    }

    return '';
}

function decodeWithPassthrough(token) {
    const bytes = base64DecodeFlexible(token);
    if (!bytes) return '';
    const text = bytesToText(bytes);
    return isValidHttpsUrl(text) ? text : '';
}
```

### 完整解密(_0xe2f081)

```js
async function decryptWithKeySeed(tokenBase64, keySeed) {
    const payloadBytes = base64DecodeFlexible(tokenBase64);
    const seedBytes = base64DecodeFlexible(keySeed);
    if (!payloadBytes || !seedBytes) return '';

    // 第一组密钥:取 seedBytes 前 32 字节做 SHA-512
    const seedHash32 = await sha512(seedBytes.slice(0, 32));

    // concat extra pepper 常量
    const extraSalt = hexDecode(SALT_HEX);                   // 16 字节
    const pepperBytes = base64DecodeFlexible(EXTRA_PEPPER);  // 32 字节

    // 第二组密钥:seedBytes 全长 → SHA-512 → concat salt → SHA-512
    const seedFull = await sha512(seedBytes);
    const combined = await sha512(concat(seedFull, extraSalt));

    const key1 = seedHash32.slice(0, 16);
    const iv1  = seedHash32.slice(16, 32);
    const key2 = combined.slice(0, 16);     // 来自 combined (加了 salt 的)
    const iv2  = combined.slice(16, 32);

    // 第三组密钥:keySeed 全部 bytes → SHA-512 → concat pepper → SHA-512
    const seedShaFull = await sha512(seedBytes);
    const key3 = (await sha512(concat(seedShaFull, pepperBytes))).slice(0, 16);
    const iv3  = (await sha512(concat(seedShaFull, pepperBytes))).slice(16, 32);

    // 识别 magic 'qAAB' (A8 00 01 00)
    const candidates = [];
    if (payloadBytes.length >= 4 &&
        payloadBytes[0] === 0xA8 && payloadBytes[1] === 0x00 &&
        payloadBytes[2] === 0x01 && payloadBytes[3] === 0x00) {

        // 多路尝试
        candidates.push({ payload: payloadBytes.slice(4), key: key3, iv: iv3 });
        candidates.push({ payload: payloadBytes.slice(4), key: key1, iv: iv1 });
        candidates.push({ payload: payloadBytes.slice(4), key: key2, iv: iv2 });

        // 嵌套变体: offset 0x24 后的部分可能有新 iv
        if (payloadBytes.length > 0x24) {
            candidates.push({
                payload: payloadBytes.slice(0x24),
                key: key1,
                iv: payloadBytes.slice(0x14, 0x24)
            });
            candidates.push({ payload: payloadBytes.slice(0x24), key: key1, iv: iv2 });
        }
    } else {
        // 无 magic,整个 payload 直接当密文
        candidates.push({ payload: payloadBytes, key: key1, iv: iv2 });
    }

    for (const c of candidates) {
        const result = await decryptAesCbc(c.payload, c.key, c.iv);
        if (result) return result;
    }
    return '';
}
```

**魔术头 `0xA8 00 01 00`**(即 base64 字符串以 `qAAB` 开头):
- 这是豆包自己定义的协议标志,表示"这是加密 payload"。
- 剥掉这 4 字节后才是真正的 AES 密文。

### AES-128-CBC 解密 + 嵌套

```js
async function decryptAesCbc(ciphertext, key, iv) {
    if (!ciphertext.length || ciphertext.length % 16 !== 0) return '';
    try {
        const cryptoKey = await crypto.subtle.importKey(
            'raw', key, 'AES-CBC', false, ['decrypt']);
        const plainBytes = new Uint8Array(await crypto.subtle.decrypt(
            { name: 'AES-CBC', iv }, cryptoKey, ciphertext));

        // 直接解:也许就是 URL
        const direct = bytesToText(plainBytes);
        if (isValidHttpsUrl(direct)) return direct;

        // 嵌套:也许要去掉 PKCS7 padding 后再读
        const unpadded = pkcs7Unpad(plainBytes);
        const inner = bytesToText(unpadded);
        return isValidHttpsUrl(inner) ? inner : '';
    } catch {
        return '';
    }
}

function pkcs7Unpad(bytes) {
    if (!bytes || !bytes.length) return new Uint8Array();
    const pad = bytes[bytes.length - 1];
    if (pad < 1 || pad > 16 || pad > bytes.length) return bytes;
    // 校验最后 pad 个字节都是 pad 值
    for (let i = bytes.length - pad; i < bytes.length; i++) {
        if (bytes[i] !== pad) return bytes;
    }
    return bytes.slice(0, bytes.length - pad);
}
```

**AES-128-CBC 的三个要素**:
- `key`:16 字节
- `iv`:16 字节
- `ciphertext`:必须是 16 的倍数

**PKCS7 padding 的校验**:最后一个字节的值 N(1≤N≤16),表示末尾有 N 个值都为 N 的字节。工具手工实现 unpad(而不是让 WebCrypto 自己处理)是因为 WebCrypto 的 AES-CBC 必须用标准 padding,但豆包的 payload 可能没加 padding 或用了别的格式。

## 入口函数

```js
async function decode(response, apiUrl) {
    // 1. 从 response 里挖 video_info 或 video_list
    const info = extractVideoInfo(response);
    // 2. 挑一个最高质量的 entry
    const entry = pickMainUrlEntry(info);
    if (!entry) return { ok: false, reason: 'fallback_main_url_missing' };

    // 3. 拿 key_seed (可能在 response 顶层或嵌套)
    const keySeed = extractKeySeed(response) || extractKeySeed(apiUrl);

    // 4. 解密
    const url = await decryptToken(entry.token, keySeed);
    return url ? { ok: true, url } : { ok: false, reason: 'fallback_url_decode_failed' };
}
```

### pickMainUrlEntry

从 video_list 里挑最佳 entry(可能有多个清晰度):

```js
function pickMainUrlEntry(data) {
    const videos = data?.video_list;
    const entries = videos && typeof videos === 'object' && Object.keys(videos).length
                    ? Object.values(videos) : [data];

    let best = null;
    for (const entry of entries) {
        if (!entry || typeof entry !== 'object') continue;
        const token = entry.main_url || entry.play_url || '';
        if (typeof token !== 'string' || !token.trim()) continue;

        const codec = String(entry.codec_type ?? '').toLowerCase();
        // 评分:H.264 优先 + 比特率 + 分辨率
        const score =
            (/h264|avc/.test(codec) || entry.codec_type === 0 ? 1_000_000_000 : 0) +
            Number(entry.bitrate || entry.real_bitrate || 0) +
            Number(entry.vwidth || entry.width || 0) *
            Number(entry.vheight || entry.height || 0);

        if (!best || score > best.score) {
            best = { token: token.trim(), score, entry };
        }
    }
    return best;
}
```

**为什么优先 H.264**:兼容性最好,所有播放器都认。更新的 H.265/HEVC 可能有的播放器放不出来。

## 完整的字段路径

sellback_url 的 JSON 响应结构(简化):

```json
{
  "data": {
    "video_info": {
      "video_list": {
        "video_1": {
          "main_url": "qAAB<base64 加密 token>",
          "key_seed": "<base64 seed>",       // 可能在 entry 内
          "codec_type": 0,                   // 0 = H.264
          "bitrate": 2500000,
          "vwidth": 1920,
          "vheight": 1080
        },
        "video_2": { ... }
      }
    }
  },
  "key_seed": "<base64 seed>"                // 可能在顶层
}
```

工具用 `extractVideoInfo` 和 `extractKeySeed` 两个递归搜索把这些字段挖出来。

## 用 Rust 重写

### 推荐 crates

```toml
[dependencies]
aes = "0.8"
cbc = { version = "0.1", features = ["alloc"] }
sha2 = "0.10"
base64 = "0.21"
```

### 等价 Rust 代码

```rust
use aes::Aes128;
use base64::{Engine, engine::general_purpose};
use cbc::Decryptor;
use cbc::cipher::{BlockDecryptMut, KeyIvInit};
use sha2::{Digest, Sha512};

const SALT_HEX: &str = "4dd4c2e6b83162090e52b3c7a6733ba4";
const EXTRA_PEPPER: &str = "TdTC5rgxYgkOUrPHpnM7pByyRiuCmrWKGWs521cXdST0m69/COjWjSanLjfBqVovHwWlGJKu8pSXMrYqOKrdWA==";

fn base64_decode_flexible(input: &str) -> Option<Vec<u8>> {
    let input = input.trim();
    let variants = [
        input.to_string(),
        input.replace(|c| matches!(c, '$' | '@' | '#'), |c| match c {
            '$' => "_", '@' => "/", '#' => ".",
            _ => unreachable!(),
        }.to_string()),
        // ... 第三种
    ];
    for v in variants {
        let padded = pad_base64(&v);
        let normalized = padded.replace('-', "+").replace('_', "/");
        if let Ok(bytes) = general_purpose::STANDARD.decode(&normalized) {
            return Some(bytes);
        }
    }
    None
}

fn sha512(input: &[u8]) -> Vec<u8> {
    Sha512::digest(input).to_vec()
}

fn aes_cbc_decrypt(ciphertext: &[u8], key: &[u8], iv: &[u8]) -> Option<Vec<u8>> {
    if ciphertext.len() % 16 != 0 || key.len() != 16 || iv.len() != 16 { return None; }

    let decryptor = Decryptor::<Aes128>::new_from_slices(key, iv).ok()?;
    let mut buf = ciphertext.to_vec();
    // 自己 unpad 而不是让 Pkcs7 自动处理(可能 padding 格式不标准)
    let n = buf.len();
    decryptor.decrypt_padded::<cbc::cipher::block_padding::NoPadding>(&mut buf).ok()?;
    Some(buf)
}

async fn decrypt_token(token: &str, key_seed: &str) -> Option<String> {
    // ... 对标 JS 版本 ...
}
```

**Rust 的好处**:
- 原生并发 — 可以同时试多个 (key, iv) 组合
- 零依赖 crypto — `aes` + `cbc` + `sha2` 都是 RustCrypto 组织的,纯 Rust 实现
- 没有 WebCrypto 的 PKCS7 强制依赖,可以直接用 `NoPadding` 自己 unpad

### 测试策略

Rust 实现写完后,**务必抓一个真实样本做端到端测试**:
1. 用沈浪跑一次视频生成,保留 OutputDirectory 里的 doubao-result-capture-report.json
2. 从 report 里拿到 fallback_url 和 key_seed(可能需要在 PS 脚本里加日志记录)
3. 单测:Rust 版本解出来的 URL 是否和原版本一致、是否能下载

**不要信"算法移植成功就能用"**,豆包的 token 格式细节可能和你想的不一样。

## 维护风险

**这块代码最脆弱**。字节升级豆包客户端时可能:

| 变化 | 影响 |
|---|---|
| 换 salt 常量 | key2/iv2 失效,但其他路径可能仍工作 → 功能部分可用 |
| 换 extra pepper 常量 | key3/iv3 失效 |
| 换 magic 头 (`0xA8 00 01 00` → 别的) | 识别不出加密 payload,整个解密链断 |
| 改 base64 字符替换规则 | 容错 base64 失效 |
| 改 AES mode (CBC → GCM) | 解密算法根本不对 |
| 完全重做协议(比如用 ECC/ChaCha20) | 全部重写 |

**监控策略**:
- 服务跑的时候记录"视频 fallback 解密成功率"
- 突然大幅下降 → 推送告警
- 自动对比沈浪新版本和当前 doubao-video-decoder.js 的 hash

## 下一步

[07-native-shadow.md](./07-native-shadow.md) — 看影子进程 doubao-tracker.exe 怎么用。
