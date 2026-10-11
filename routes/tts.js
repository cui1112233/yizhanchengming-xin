const express = require('express');
const http = require('http');
const { apiAuth } = require('../middleware/auth');

const TTS_UPSTREAM = Object.freeze({
  origin: 'http://tts3.121w.com',
  hostname: 'tts3.121w.com',
  path: '/v1/audio/speech'
});
const router = express.Router();
router.use(apiAuth);

// POST /api/tts — 文本转语音代理（转发到 tts3.121w.com）
router.post('/', async (req, res) => {
  try {
    const { input, voice, speed, pitch, style } = req.body;
    if (!input || !input.trim()) {
      return res.status(400).json({ error: '输入文本不能为空' });
    }

    const payload = JSON.stringify({
      input: input.trim(),
      voice: voice || 'zh-CN-XiaoxiaoNeural',
      speed: Number.isFinite(parseFloat(speed)) ? parseFloat(speed) : 1.8,
      pitch: pitch !== undefined ? String(pitch) : '10',
      style: style || 'general'
    });

    const ttsReq = http.request(`${TTS_UPSTREAM.origin}${TTS_UPSTREAM.path}`, {
      method: 'POST',
      headers: {
        'Host': 'tts3.121w.com',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Accept': '*/*',
        'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
        'Accept-Encoding': 'identity',
        'Referer': 'http://tts3.121w.com/',
        'Origin': 'http://tts3.121w.com',
        'Connection': 'keep-alive',
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(payload)
      },
      timeout: 60000
    }, ttsRes => {
      const chunks = [];
      ttsRes.on('data', chunk => chunks.push(chunk));
      ttsRes.on('end', () => {
        const buffer = Buffer.concat(chunks);
        if (res.headersSent) return;
        if (ttsRes.statusCode >= 400) {
          const text = buffer.toString('utf8').slice(0, 500);
          res.status(502).json({ error: 'TTS 服务错误 (' + ttsRes.statusCode + '): ' + text });
          return;
        }
        if (buffer.length === 0) {
          res.status(502).json({ error: 'TTS 服务返回空音频，请稍后重试' });
          return;
        }
        const contentType = ttsRes.headers['content-type'] || '';
        if (contentType.includes('application/json')) {
          const text = buffer.toString('utf8').slice(0, 500);
          res.status(502).json({ error: 'TTS 服务返回错误: ' + text });
          return;
        }
        res.status(200);
        if (contentType) res.setHeader('Content-Type', contentType);
        res.setHeader('Content-Length', buffer.length);
        res.end(buffer);
      });
      ttsRes.on('error', error => {
        console.error('TTS upstream response error:', error);
        if (!res.headersSent) {
          res.status(502).json({ error: 'TTS 服务响应异常: ' + error.message });
        }
      });
    });

    ttsReq.on('error', error => {
      console.error('TTS upstream error:', error);
      if (!res.headersSent) {
        res.status(502).json({ error: 'TTS 服务请求失败: ' + error.message });
      }
    });

    ttsReq.write(payload);
    ttsReq.end();
  } catch (error) {
    console.error('TTS proxy error:', error);
    res.status(500).json({ error: error.message || 'Internal server error' });
  }
});

module.exports = router;
module.exports.TTS_UPSTREAM = TTS_UPSTREAM;
