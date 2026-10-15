# -*- coding: utf-8 -*-
"""Generate giant-material-executor icon.ico (stdlib only, supersampled BMP-ICO)."""
import struct, math, os

def rounded_rect_mask(x, y, x0, y0, x1, y1, r):
    if x < x0 or x > x1 or y < y0 or y > y1:
        return False
    cx = min(max(x, x0 + r), x1 - r)
    cy = min(max(y, y0 + r), y1 - r)
    return (x - cx) ** 2 + (y - cy) ** 2 <= r * r or (x0 + r <= x <= x1 - r) or (y0 + r <= y <= y1 - r)

def sample(u, v):
    """u,v in 0..1 over the 256x256 logical canvas. Returns (r,g,b,a) 0..255."""
    x, y = u * 256, v * 256
    # transparent margin
    if not rounded_rect_mask(x, y, 8, 8, 248, 248, 44):
        return (0, 0, 0, 0)
    # dark navy rounded background
    r, g, b, a = 13, 34, 48, 255
    # white document x=64..192 y=40..216, rounded r=12
    if rounded_rect_mask(x, y, 64, 40, 192, 216, 12):
        r, g, b = 244, 250, 252
        # text lines (dark navy bars)
        for ly in (78, 112, 146):
            if ly - 5 <= y <= ly + 5 and 88 <= x <= 168:
                r, g, b = 13, 34, 48
        # mint scan beam y=166..190 with soft edge
        if 166 <= y <= 190 and 64 <= x <= 192:
            edge = min(y - 166, 190 - y)
            if edge > 6:
                r, g, b = 100, 223, 192
            else:
                t = edge / 6.0
                r = int(13 + (100 - 13) * t)
                g = int(34 + (223 - 34) * t)
                b = int(48 + (192 - 48) * t)
    # mint border ring around whole icon
    if rounded_rect_mask(x, y, 8, 8, 248, 248, 44) and not rounded_rect_mask(x, y, 20, 20, 236, 236, 34):
        r, g, b = 100, 223, 192
    return (r, g, b, a)

def render(size):
    ss = 4  # supersample factor
    rows = []
    for py in range(size):
        row = bytearray()
        for px in range(size):
            r = g = b = a = 0
            for sy in range(ss):
                for sx in range(ss):
                    u = (px * ss + sx + 0.5) / (size * ss)
                    v = (py * ss + sy + 0.5) / (size * ss)
                    pr, pg, pb, pa = sample(u, v)
                    r += pr; g += pg; b += pb; a += pa
            n = ss * ss
            # BGRA, straight alpha (keep simple: premultiply not required for our opaque design)
            row += bytes((r // n, g // n, b // n, a // n))
        rows.append(row)
    return rows  # top-down BGRA rows

def bmp_payload(rows):
    h = len(rows)
    w = len(rows[0]) // 4
    header = struct.pack('<IiiHHIIiiII', 40, w, h * 2, 1, 32, 0, w * h * 4 + h * ((w + 31) // 32) * 4, 0, 0, 0, 0)
    xor = b''.join(rows[::-1])  # bottom-up
    stride = ((w + 31) // 32) * 4
    and_mask = bytes(stride) * h  # opaque everywhere (alpha channel governs)
    return header + xor + and_mask

def write_ico(path, sizes):
    images = [bmp_payload(render(s)) for s in sizes]
    out = struct.pack('<HHH', 0, 1, len(sizes))
    offset = 6 + 16 * len(sizes)
    for s, data in zip(sizes, images):
        out += struct.pack('<BBBBHHII', s % 256, s % 256, 0, 0, 1, 32, len(data), offset)
        offset += len(data)
    for data in images:
        out += data
    with open(path, 'wb') as f:
        f.write(out)
    print(f'wrote {path} ({len(out)} bytes)')

out_dir = os.path.dirname(os.path.abspath(__file__))
write_ico(os.path.join(out_dir, 'icon.ico'), [32, 256])
