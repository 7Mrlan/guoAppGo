package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/MoeForge/moejs"
)

// 最小验证：moejs 在 Go 进程内编译/执行 ES 模块、调用导出函数、把结果拿回 Go，
// 并通过 NativeFunc 反向调用宿主 Go 函数。
const pluginSrc = `
export function summarize(items) {
  return items.reduce((a, b) => a + b, 0);
}

export function upper(s) {
  return s.toUpperCase() + " @ " + now();
}

export function boom() {
  throw new Error("kaboom");
}

export function spin() { for (;;) {} }
`

func hostNow(r *moejs.Realm, this moejs.Value, args []moejs.Value) (moejs.Value, error) {
	return moejs.String("host-time"), nil
}

func main() {
	mod, err := moejs.Compile("plugin.js", pluginSrc)
	if err != nil {
		fmt.Println("compile FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("compile OK")

	rt := moejs.NewRuntime(moejs.Options{})
	if err := rt.SetGlobal("now", moejs.NativeFunc(hostNow)); err != nil {
		fmt.Println("SetGlobal FAIL:", err)
		os.Exit(1)
	}
	if err := rt.Load(mod); err != nil {
		fmt.Println("load FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("load OK")

	hookSum, err := mod.Hook("summarize")
	if err != nil {
		fmt.Println("hook FAIL:", err)
		os.Exit(1)
	}
	arg, err := rt.ParseJSON([]byte(`[1,2,3,4]`))
	if err != nil {
		fmt.Println("ParseJSON FAIL:", err)
		os.Exit(1)
	}
	res, err := rt.Call(hookSum, arg)
	if err != nil {
		fmt.Println("call FAIL:", err)
		os.Exit(1)
	}
	out, err := rt.AppendJSON(nil, res)
	rt.ReleaseCallData()
	if err != nil {
		fmt.Println("AppendJSON FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("call summarize([1,2,3,4]) ->", string(out))

	// 导出函数反向调用宿主 Go 函数
	hookUpper, err := mod.Hook("upper")
	if err != nil {
		fmt.Println("hook upper FAIL:", err)
		os.Exit(1)
	}
	sres, err := rt.Call(hookUpper, moejs.String("abc"))
	if err != nil {
		fmt.Println("call upper FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("call upper('abc') ->", sres.String())

	// JS 抛异常：Go 侧拿到 error 而不是 panic
	hookBoom, _ := mod.Hook("boom")
	if _, err := rt.Call(hookBoom); err != nil {
		fmt.Println("boom -> error (expected):", err)
	}

	// 死循环可被 context 超时打断
	hookSpin, _ := mod.Hook("spin")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	stop := context.AfterFunc(ctx, func() { rt.Interrupt(context.Cause(ctx)) })
	_, err = rt.Call(hookSpin)
	stop()
	cancel()
	rt.ClearInterrupt()
	fmt.Println("spin interrupt err:", err)

	// 调用耗时：同一 runtime 连调 1000 次
	start := time.Now()
	const n = 1000
	arg2, _ := rt.ParseJSON([]byte(`[1,2,3]`))
	for i := 0; i < n; i++ {
		if _, err := rt.Call(hookSum, arg2); err != nil {
			fmt.Println("bench call FAIL:", err)
			os.Exit(1)
		}
	}
	rt.ReleaseCallData()
	el := time.Since(start)
	fmt.Printf("bench: %d calls in %v = %v/call\n", n, el, el/n)
}
