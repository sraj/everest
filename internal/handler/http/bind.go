package http

import (
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/pkg/bind"
)

func bindBody(c *fiber.Ctx, target any) error {
	if err := bind.Body(c, target); err != nil {
		if bind.IsValidation(err) {
			return apperror.ValidationError(err)
		}
		return apperror.BadRequest("invalid request: %v", err)
	}
	return nil
}

func bindQuery(c *fiber.Ctx, target any) error {
	if err := bind.Query(c, target); err != nil {
		if bind.IsValidation(err) {
			return apperror.ValidationError(err)
		}
		return apperror.BadRequest("invalid query: %v", err)
	}
	return nil
}

func readMultipartBody(c *fiber.Ctx) (*MultipartBody, error) {
	body := &MultipartBody{}

	body.Title = c.FormValue("title")
	if body.Title == "" {
		body.Title = "Untitled Document"
	}

	file, err := c.FormFile("file")
	if err != nil {
		body.Content = []byte(c.FormValue("content"))
		body.ContentType = "text/html"
	} else {
		f, err := file.Open()
		if err != nil {
			return nil, apperror.BadRequest("failed to open uploaded file")
		}
		defer f.Close()

		body.Content, err = io.ReadAll(f)
		if err != nil {
			return nil, apperror.BadRequest("failed to read uploaded file")
		}

		body.ContentType = file.Header.Get("Content-Type")
		if body.ContentType == "" {
			body.ContentType = "application/octet-stream"
		}

		if body.Title == "Untitled Document" && file.Filename != "" {
			body.Title = file.Filename
		}
		body.FileName = file.Filename
	}

	return body, nil
}

func isMultipart(c *fiber.Ctx) bool {
	return strings.HasPrefix(string(c.Request().Header.ContentType()), "multipart/form-data")
}
