package main

import (
	"context"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"skygate/internal/grpcapi"
	"skygate/internal/metrics"
	"skygate/internal/rest"
	"skygate/internal/sim"
	"skygate/internal/store"
	"skygate/internal/web"
	"skygate/internal/ws"
)

func main() {
	port := getenv("PORT", "8080")
	grpcPort := getenv("GRPC_PORT", "9090")

	st := store.New()
	engine := sim.NewEngine(st, sim.RealClock, time.Now().UnixNano())
	engine.Seed()

	hub := ws.NewHub()
	engine.Subscribe(hub.Broadcast)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go engine.Run(ctx, time.Second)

	version := getenv("VERSION", "dev")
	color := getenv("COLOR", "blue")
	// buildVersion is the actual software release baked into the image at
	// build time - kept distinct from version, which routing-canary demo
	// deployments override with a "v1"/"v2" label to compare instances.
	buildVersion := getenv("BUILD_VERSION", "dev")
	panicEnabled := getenv("CHAOS_PANIC_ENABLED", "false") == "true"
	mux, chaos := rest.NewMux(rest.Deps{
		Store: st, Version: version, Color: color, BuildVersion: buildVersion, PanicEnabled: panicEnabled,
	})
	mux.Handle("GET /ws/flights", ws.Handler(st, hub))

	staticFS, err := fs.Sub(web.Static, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /", http.FileServerFS(staticFS))

	// Wrap the fully assembled mux exactly once, here, after every route -
	// REST, WebSocket, and static UI - has been registered on it. See the
	// doc comment on rest.NewMux for why wrapping any earlier breaks GET
	// routing once a method-specific pattern like "GET /" is added.
	handler := metrics.Instrument(version, chaos.Middleware(mux))
	srv := &http.Server{Addr: ":" + port, Handler: handler}
	go func() {
		log.Printf("skygate REST listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	grpcServer := grpcapi.NewServer(st)
	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Printf("skygate gRPC listening on :%s", grpcPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Print("shutting down")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
	grpcServer.GracefulStop()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
