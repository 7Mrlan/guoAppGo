// 在系统 webview 里做一次特性检测，把结果从页面回报到 Go 进程并落盘。
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/egoist/mygo"
)

const page = `<!doctype html>
<meta charset="utf-8">
<title>mse-probe</title>
<body>probing…</body>
<script>
const out = {
  ua: navigator.userAgent,
  hasMediaSource: ('MediaSource' in window),
  hasManagedMediaSource: ('ManagedMediaSource' in window),
  hasSourceBuffer: ('SourceBuffer' in window),
  mediaSourceType: typeof MediaSource,
  managedType: typeof window.ManagedMediaSource,
  mp4Avc: null,
  mp4Hevc: null,
  nativeHls: document.createElement('video').canPlayType('application/vnd.apple.mpegurl'),
  newMediaSourceOK: false,
  newMediaSourceErr: null,
  mse_audioVideo: null
};
try {
  const ms = new MediaSource();
  out.newMediaSourceOK = true;
  out.mse_audioVideo = ms.constructor.name || String(ms);
  out.mp4Avc = MediaSource.isTypeSupported('video/mp4; codecs="avc1.42E01E"');
  out.mp4Hevc = MediaSource.isTypeSupported('video/mp4; codecs="hvc1.1.6.L93.B0"');
} catch (e) {
  out.newMediaSourceErr = String(e && e.message || e);
}
fetch('/report?d=' + encodeURIComponent(JSON.stringify(out))).then(() => {});
</script>`

func main() {
	outPath := os.Getenv("PROBE_OUT")
	if outPath == "" {
		outPath = "probe-result.json"
	}

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
	mux.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		body := r.URL.Query().Get("d")
		_ = os.WriteFile(outPath, []byte(body), 0o644)
		fmt.Println("REPORT:", body)
		w.WriteHeader(http.StatusNoContent)
		go func() { time.Sleep(300 * time.Millisecond); os.Exit(0) }()
	})
	go func() { _ = http.Serve(ln, mux) }()

	// 兜底：15 秒内没回报就按超时退出
	go func() {
		time.Sleep(15 * time.Second)
		_ = os.WriteFile(outPath, []byte(`{"error":"timeout: page never reported"}`), 0o644)
		fmt.Println("TIMEOUT")
		os.Exit(2)
	}()

	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:  "mse-probe",
			URL:    "http://" + addr + "/probe.html",
			Width:  640,
			Height: 420,
		})
	})
	if err := mygo.App.Run(); err != nil {
		fmt.Println("app FAIL:", err)
		os.Exit(1)
	}
}
