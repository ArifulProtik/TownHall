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
	if lastID != "" {
		ctx := r.Context()
		cursor := encodeCursor(lastID)
		for {
			page, err := h.svc.List(ctx, uid, maxListLimit, cursor)
			if err != nil || len(page.Notifications) == 0 {
				break
			}
			// List is newest-first for the bell; replay oldest-first so
			// a reconnecting client sees missed events in order.
			for i := len(page.Notifications) - 1; i >= 0; i-- {
				item := page.Notifications[i]
				_ = writeSSE(w, item.ID, "notification", item)
				flush()
			}
			if !page.HasMore {
				break
			}
			cursor = page.NextCursor
		}
	}

	ch, unsub := h.svc.Subscribe(uid)
	defer unsub()
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
			_ = writeSSE(w, e.ID, "notification", e)
			flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flush()
		}
	}
}
