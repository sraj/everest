package http

import (
	"bufio"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
)

// ChatStream handles GET /api/chat/stream — SSE connection for server-pushed events.
func (h *Handler) chatStream(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	done := c.Context().Done()

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
		w.Flush()
		<-done
	})

	return nil
}

// ChatMessages handles POST /api/chat/messages.
// Accepts both the Message format ({ id, role, content, timestamp }) and
// the fetch-sse format ({ message }). Returns JSON for SSE protocol,
// or SSE stream for fetch-sse protocol.
func (h *Handler) chatMessages(c *fiber.Ctx) error {
	// Accept both body formats
	var body struct {
		ID        string `json:"id"`
		Role      string `json:"role"`
		Content   string `json:"content"`
		Timestamp int64  `json:"timestamp"`
		Message   string `json:"message"`
	}
	if err := bindBody(c, &body); err != nil {
		return err
	}

	content := body.Content
	if content == "" {
		content = body.Message
	}
	if content == "" {
		return apperror.BadRequest("message is required")
	}
	if len(content) > 2000 {
		content = content[:2000]
	}

	acceptSSE := string(c.Request().Header.Peek("Accept")) == "text/event-stream"

	if acceptSSE {
		// fetch-sse protocol: stream SSE events
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			if err := h.chatService.Stream(c.Context(), content, w); err != nil {
				h.log.Error("chat stream failed", "error", err)
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				w.Flush()
			}
		})
		return nil
	}

	// SSE protocol: return complete JSON response
	answer, sources, err := h.chatService.Process(c.Context(), content)
	if err != nil {
		return apperror.Wrap(err, 500, "internal", "failed to process message")
	}

	return c.JSON(fiber.Map{
		"response": fiber.Map{
			"id":        body.ID,
			"role":      "assistant",
			"content":   answer,
			"timestamp": body.Timestamp,
			"metadata": fiber.Map{
				"citations":    sources,
				"streamStatus": "done",
			},
		},
	})
}
