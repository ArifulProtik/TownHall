package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ArifulProtik/TownHall/pkg/response"

	"github.com/labstack/echo/v5"
)

// ServiceAPI is what the handler needs. *Service satisfies it, tests mock it.
type ServiceAPI interface {
	List(ctx context.Context, userID string, limit int, cursor string) (*ListResponse, error)
	ListSince(ctx context.Context, userID, lastID string, limit int) ([]Item, error)
	UnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, userID, id string) error
	MarkAllRead(ctx context.Context, userID string) error
	Subscribe(userID string) (<-chan Event, func())
}

type Handler struct {
	svc ServiceAPI
}

func NewHandler(svc ServiceAPI) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterRoutes(public *echo.Group, protected *echo.Group) {
	protected.GET("/notifications", h.List)
	protected.GET("/notifications/unread-count", h.UnreadCount)
	protected.POST("/notifications/:id/read", h.MarkRead)
	protected.POST("/notifications/read-all", h.MarkAllRead)
	// Header-auth SSE: frontend uses fetch with Authorization, never ?token=.
	protected.GET("/stream", h.Stream)
}

func listParams(c *echo.Context) (int, string) {
	limit := defaultListLimit
	if raw := strings.TrimSpace(c.QueryParam("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	return clampLimit(limit), c.QueryParam("cursor")
}

func (h *Handler) List(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	limit, cursor := listParams(c)
	resp, err := h.svc.List(c.Request().Context(), uid, limit, cursor)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UnreadCount(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	n, err := h.svc.UnreadCount(c.Request().Context(), uid)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, UnreadResponse{Count: n})
}

func (h *Handler) MarkRead(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		return response.BadRequest(c, "id is required")
	}
	if err := h.svc.MarkRead(c.Request().Context(), uid, id); err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, response.Map{"status": "ok"})
}

func (h *Handler) MarkAllRead(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	if err := h.svc.MarkAllRead(c.Request().Context(), uid); err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, response.Map{"status": "ok"})
}

func writeSSE(w http.ResponseWriter, id, event string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if id != "" {
		_, _ = fmt.Fprintf(w, "id: %s\n", id)
	}
	if event != "" {
		_, _ = fmt.Fprintf(w, "event: %s\n", event)
	}
	_, _ = fmt.Fprintf(w, "data: %s\n\n", raw)
	return nil
}

// Stream replays missed events since Last-Event-ID, then streams live.
// It ends on client disconnect (request context cancel).
func (h *Handler) Stream(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	r := c.Request()
	w := c.Response()
	hdr := w.Header()
	hdr.Set("Content-Type", "text/event-stream")
	hdr.Set("Cache-Control", "no-cache")
	hdr.Set("Connection", "keep-alive")
	hdr.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	flush := func() {
		if flusher != nil {
			flusher.Flush()
		}
	}
	flush()

	lastID := r.Header.Get("Last-Event-ID")
	if lastID == "" {
		lastID = r.URL.Query().Get("lastEventId")
	}
	// Subscribe BEFORE replay: events published during the replay query
	// land in the channel instead of falling in the gap. Anything the
	// replay also returns is skipped in the live loop via replayedThrough
	// (UUIDv7 ids sort as creation order, so string order is time order).
	ch, unsub := h.svc.Subscribe(uid)
	defer unsub()
	replayedThrough := ""
	if lastID != "" {
		items, err := h.svc.ListSince(r.Context(), uid, lastID, maxListLimit)
		if err == nil {
			for _, item := range items {
				_ = writeSSE(w, item.ID, "notification", item)
				flush()
				replayedThrough = item.ID
			}
		}
	}

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return nil
		case e, ok := <-ch:
			if !ok {
				return nil
			}
			if replayedThrough != "" && e.ID <= replayedThrough {
				continue
			}
			_ = writeSSE(w, e.ID, "notification", e)
			flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flush()
		}
	}
}
