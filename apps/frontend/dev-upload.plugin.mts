import { existsSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import type { Plugin } from 'vite'

// 개발용: 폰에서 user vault로 이미지 올리기(/__upload?t=토큰). 평소엔 vite.config.mts에서 주석 처리 — REF-test-env.md "폰 사진 올리기"
// test1은 외부에 열려 있으므로 토큰이 없거나 틀리면 404로 숨긴다

const PAGE = `<!doctype html><meta name="viewport" content="width=device-width,initial-scale=1">
<title>upload</title>
<style>body{font:16px sans-serif;padding:16px}input{font-size:18px}li{margin:6px 0}</style>
<p><input id="f" type="file" accept="image/*" multiple></p><ul id="log"></ul>
<script>
const log = (t) => { const li = document.createElement('li'); li.textContent = t; document.getElementById('log').prepend(li) }
// 프록시 본문 한도를 피하려고 900KB 넘으면 JPEG로 줄인다
async function prep(f) {
  if (f.size < 900e3 && /^image\\/(png|jpeg)$/.test(f.type)) return f
  const bmp = await createImageBitmap(f)
  const s = Math.min(1, 2000 / Math.max(bmp.width, bmp.height))
  const c = document.createElement('canvas')
  c.width = Math.round(bmp.width * s); c.height = Math.round(bmp.height * s)
  c.getContext('2d').drawImage(bmp, 0, 0, c.width, c.height)
  let b
  for (const q of [0.9, 0.8, 0.65, 0.5]) { b = await new Promise(r => c.toBlob(r, 'image/jpeg', q)); if (b.size < 900e3) break }
  return b
}
document.getElementById('f').onchange = async (e) => {
  for (const f of e.target.files) {
    try {
      const b = await prep(f)
      const r = await fetch(location.href, { method: 'POST', headers: { 'content-type': b.type }, body: b })
      log(r.ok ? 'OK ' + await r.text() + ' (' + Math.round(b.size / 1e3) + 'KB)' : 'FAIL ' + r.status + ' ' + f.name)
    } catch (err) { log('FAIL ' + f.name + ' ' + err) }
  }
  e.target.value = ''
}
</script>`

const EXT: Record<string, string> = { 'image/png': 'png', 'image/jpeg': 'jpg' }

const stamp = () => {
    const d = new Date(), p = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}_${p(d.getHours())}-${p(d.getMinutes())}-${p(d.getSeconds())}`
}

export const devUpload = (dir: string, token: string): Plugin => ({
    name: 'dev-upload',
    apply: 'serve',
    configureServer(server) {
        server.middlewares.use('/__upload', (req, res) => {
            const t = new URL(req.url ?? '/', 'http://x').searchParams.get('t')
            if (!token || t !== token) {
                res.statusCode = 404
                return res.end()
            }
            if (req.method === 'GET') {
                res.setHeader('content-type', 'text/html; charset=utf-8')
                return res.end(PAGE)
            }
            const ext = EXT[req.headers['content-type'] ?? '']
            if (req.method !== 'POST' || !ext) {
                res.statusCode = 400
                return res.end()
            }
            const chunks: Buffer[] = []
            let size = 0
            req.on('data', (c: Buffer) => {
                size += c.length
                if (size > 20e6) req.destroy()
                else chunks.push(c)
            })
            req.on('end', () => {
                const base = `photo_${stamp()}`
                let name = `${base}.${ext}`
                for (let i = 2; existsSync(join(dir, name)); i++) name = `${base}_${i}.${ext}`
                writeFileSync(join(dir, name), Buffer.concat(chunks))
                res.end(name)
            })
        })
    },
})
