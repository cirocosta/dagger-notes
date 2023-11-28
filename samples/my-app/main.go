package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func run(ctx context.Context) error {
	srv := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				log.Println("responding")
				fmt.Fprintln(w, "OK")
			},
		),
	}

	defer srv.Shutdown(ctx)

	srvC := make(chan error, 1)
	go func() {
		log.Println("listening")
		srvC <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("done")
		return nil
	case err := <-srvC:
		return err
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt,
	)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
