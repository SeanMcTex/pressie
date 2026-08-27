package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/store"
	"github.com/SeanMcTex/pressie/internal/web"
)

// cmdServe starts the pressie web server.
func cmdServe(args []string) {
	if wantsHelp(args) {
		helpServe()
		return
	}

	port := 7612
	authToken := ""

	// Parse serve-specific flags manually (not all flags apply).
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--port", "-p":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "pressie: --port requires a value")
				os.Exit(1)
			}
			p, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "pressie: invalid port: %s\n", args[i])
				os.Exit(1)
			}
			port = p
		case "--auth-token":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "pressie: --auth-token requires a value")
				os.Exit(1)
			}
			authToken = args[i]
		case "--auth-token=":
			authToken = arg[len("--auth-token="):]
		case "-h", "--help":
			helpServe()
			return
		default:
			if len(arg) > 12 && arg[:12] == "--auth-token=" {
				authToken = arg[12:]
				continue
			}
			if len(arg) > 7 && arg[:7] == "--port=" {
				p, err := strconv.Atoi(arg[7:])
				if err != nil {
					fmt.Fprintf(os.Stderr, "pressie: invalid port: %s\n", arg[7:])
					os.Exit(1)
				}
				port = p
				continue
			}
		}
	}

	giftsDir := config.DefaultGiftsDir()

	// Verify the gifts directory is initialized.
	if _, err := store.LoadIndex(giftsDir); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: no gifts directory found at %s\n", giftsDir)
		fmt.Fprintf(os.Stderr, "Run `pressie init %s` first.\n", giftsDir)
		os.Exit(1)
	}

	// Check for auth token in index if not provided via flag.
	if authToken == "" {
		idx, _ := store.LoadIndex(giftsDir)
		if idx.Sync != nil && idx.Sync.Remote != "" {
			// Could check for web config in index here in the future.
			_ = idx
		}
	}

	srv := web.NewServer(giftsDir, authToken)
	addr := srv.ListenAddr(port)

	fmt.Printf("Pressie web server starting at http://%s\n", addr)
	if authToken == "" {
		fmt.Println("No auth token set — localhost only. Use --auth-token to expose on network.")
	} else {
		fmt.Println("Auth enabled — accessible from network.")
	}
	fmt.Println("Press Ctrl+C to stop.")

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		fmt.Fprintf(os.Stderr, "pressie: %s\n", err)
		os.Exit(1)
	}
}