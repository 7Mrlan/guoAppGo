// A 段验证：本地 fMP4 HLS + 自起 HTTP 服务，在系统 webview 里用 hls.js 播放，
// 量首帧与 seek 恢复时间。不访问任何外网、不涉及任何真实源。
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

const page = `<!doctype html>
<meta charset="utf-8">
<title>mse-chain</title>
<style>body{margin:0;background:#111}video{width:640px}</style>
<script>
// 抓住控制台错误，hls.js 的失败原因基本只在 console 里
window.__err = [];
const _ce = console.error, _cw = console.warn;
console.error = function(){ window.__err.push(Array.from(arguments).map(String).join(' ')); _ce.apply(console, arguments); };
console.warn = function(){ window.__err.push(Array.from(arguments).map(String).join(' ')); _cw.apply(console, arguments); };
window.onerror = function(m){ window.__err.push('onerror:' + m); };
window.onunhandledrejection = function(e){ window.__err.push('unhandled:' + e.reason); };
</script>
<video id="v" muted playsinline></video>
<script src="/hls.min.js"></script>
<script>
const out = {
  ua: navigator.userAgent,
  href: location.href,
  baseURI: document.baseURI,
  resolvedSeg: new URL('/hls/init.mp4', document.baseURI).href,
  hasBaseTag: !!document.querySelector('base'),
  hasMSE: 'MediaSource' in window,
  hlsJSAvailable: typeof Hls !== 'undefined',
  hlsJSVersion: (typeof Hls !== 'undefined' && Hls.version) || null,
  directFetch: null,
  fragURL: null,
  firstFrameMs: null,
  seekRecoverMs: null,
  duration: null,
  console: [],
  errors: [],
  events: []
};
const report = () => {
  out.console = window.__err.slice(0, 8);
  return fetch('/report?d=' + encodeURIComponent(JSON.stringify(out))).then(() => {});
};
const v = document.getElementById('v');
const log = (k) => out.events.push(k + '@' + Math.round(performance.now()));
let hls = null;

function measure() {
  let t0 = 0, firstDone = false, seekStart = 0;
  hls.on(Hls.Events.FRAG_LOADING, (e, d) => { out.fragURL = (d.frag && d.frag.url) || 'none'; });
  hls.on(Hls.Events.ERROR, (e, data) => {
    out.errors.push('hls:' + data.type + '/' + data.details + (data.fatal ? '/fatal' : '') +
      ' url=' + (data.url || '') + ' code=' + (data.response && data.response.code));
  });
  hls.on(Hls.Events.MANIFEST_PARSED, () => {
    out.duration = hls.media ? hls.media.duration : null;
    log('manifest');
    t0 = performance.now();
    v.play().catch(err => out.errors.push('play:' + err.message));
  });
  v.addEventListener('playing', () => {
    if (firstDone) return;
    firstDone = true;
    out.firstFrameMs = Math.round(performance.now() - t0);
    out.duration = v.duration;
    log('playing');
    setTimeout(() => {
      const target = Math.min(20, (v.duration || 30) * 0.6);
      out.events.push('seekTo:' + target.toFixed(1));
      seekStart = performance.now();
      v.currentTime = target;
    }, 3000);
  });
  v.addEventListener('seeked', () => {
    out.seekRecoverMs = Math.round(performance.now() - seekStart);
    log('seeked');
    setTimeout(report, 10000);
  });
  v.addEventListener('error', () => out.errors.push('video error: ' + (v.error && v.error.code)));
  hls.loadSource('/proxy/index.m3u8');
  hls.attachMedia(v);
}

if (typeof Hls === 'undefined' || !Hls.isSupported()) {
  out.errors.push('hls.js not supported: ' + (typeof Hls === 'undefined' ? 'Hls undefined' : 'isSupported false'));
  report();
} else {
  hls = new Hls({ enableWorker: false, lowLatencyMode: false });
  // 先确认分片本身能取到，把「网络」和「hls.js」两层问题分开
  fetch('/hls/init.mp4').then(r => { out.directFetch = r.status; })
    .catch(e => { out.directFetch = 'err:' + e.message; })
    .finally(measure);
}
setTimeout(() => { if (out.seekRecoverMs === null) report(); }, 25000);
</script>`

type statusRecorder struct {
	http.ResponseWriter
	status int
	n      int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	n, err := s.ResponseWriter.Write(b)
	s.n += n
	return n, err
}

func main() {
	b := os.Getenv("PROBE_DIR")
	if b == "" {
		b = "."
	}
	hlsFile := os.Getenv("HLS_JS")
	if hlsFile == "" {
		hlsFile = filepath.Join(b, "web", "hls.min.js")
	}
	outPath := filepath.Join(b, "mse-chain.json")
	hlsDir := filepath.Join(b, "hls")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("listen FAIL:", err)
		os.Exit(1)
	}
	addr := ln.Addr().String()

	mux := http.NewServeMux()
	mux.HandleFunc("/probe.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page)
	})
	mux.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir(hlsDir))))
	// 模拟产品里的代理：播放列表改写为自己的 URL，分片按块转发而不是一次性吐出。
	mux.HandleFunc("/proxy/index.m3u8", func(w http.ResponseWriter, r *http.Request) {
		raw, err := os.ReadFile(filepath.Join(hlsDir, "index.m3u8"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		out := make([]string, 0, 32)
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			line = strings.TrimSpace(line)
			switch {
			case line == "":
				continue
			case strings.HasPrefix(line, "#EXT-X-MAP:"):
				line = strings.Replace(line, `URI="`, `URI="/proxy/`, 1)
			case !strings.HasPrefix(line, "#"):
				line = "/proxy/" + line
			}
			out = append(out, line)
		}
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		fmt.Fprint(w, strings.Join(out, "\n")+"\n")
	})
	mux.HandleFunc("/proxy/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/proxy/")
		f, err := os.Open(filepath.Join(hlsDir, filepath.Base(name)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "video/mp4")
		buf := make([]byte, 8<<10)
		flusher, _ := w.(http.Flusher)
		for {
			n, rerr := f.Read(buf)
			if n > 0 {
				if _, werr := w.Write(buf[:n]); werr != nil {
					return
				}
				if flusher != nil {
					flusher.Flush()
				}
				time.Sleep(200 * time.Microsecond)
			}
			if rerr != nil {
				return
			}
		}
	})
	mux.HandleFunc("/hls.min.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, hlsFile)
	})
	mux.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		body := r.URL.Query().Get("d")
		_ = os.WriteFile(outPath, []byte(body), 0o644)
		fmt.Println("REPORT:", body)
		w.WriteHeader(http.StatusNoContent)
		go func() { time.Sleep(300 * time.Millisecond); os.Exit(0) }()
	})

	logged := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		mux.ServeHTTP(rec, r)
		if r.URL.Path != "/report" {
			fmt.Printf("REQ %s %s -> %d (%d bytes)\n", r.Method, r.URL.Path, rec.status, rec.n)
		}
	})
	go func() { _ = http.Serve(ln, logged) }()

	go func() {
		time.Sleep(40 * time.Second)
		_ = os.WriteFile(outPath, []byte(`{"error":"timeout"}`), 0o644)
		fmt.Println("TIMEOUT")
		os.Exit(2)
	}()

	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:  "mse-chain",
			URL:    "http://" + addr + "/probe.html",
			Width:  760,
			Height: 520,
		})
	})
	if err := mygo.App.Run(); err != nil {
		fmt.Println("app FAIL:", err)
		os.Exit(1)
	}
}
