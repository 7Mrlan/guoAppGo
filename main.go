// Command guoapp is the desktop application of guoAppGo.
package main

import (
	"log"

	"github.com/7Mrlan/guoAppGo/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
