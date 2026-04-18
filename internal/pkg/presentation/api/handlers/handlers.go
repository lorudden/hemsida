package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
)

func NewSSEHandler(appContext context.Context, version string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetFromContext(ctx)

		flusher, ok := w.(http.Flusher)
		if !ok {
			logger.Warn("streaming not supported for this response writer")
			http.Error(w, "unable to start event stream", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		const eventFmt string = "event: %s\ndata: %s\n\n"

		logger.Info("comparing versions", "client", r.PathValue("version"), "mine", version)
		waitingForUpgrade := false

		if strings.Compare(r.PathValue("version"), version) != 0 {
			waitingForUpgrade = true

			logger.Info("sending upgrade to client")
			fmt.Fprintf(w, eventFmt, "upgrade2", version)
			flusher.Flush()

			time.Sleep(10 * time.Second)

			logger.Info("sending goodbye to client")
			fmt.Fprintf(w, eventFmt, "goodbye", version)
			flusher.Flush()
		} else {
			logger.Info("sse client successfully connected")
		}

		defer func() { logger.Info("exiting sse handler") }()

		tmr := time.NewTicker(time.Second)

		for {
			select {
			case t := <-tmr.C:
				if !waitingForUpgrade {
					fmt.Fprintf(w, eventFmt, "tick", t.Format(time.RFC3339Nano))
				}
			case <-ctx.Done():
				logger.Info("sse client connection closed")
				return
			case <-appContext.Done():
				logger.Info("we are shutting down")
				fmt.Fprintf(w, eventFmt, "goodbye", version)
				flusher.Flush()
				return
			}

			flusher.Flush()
		}
	})
}
