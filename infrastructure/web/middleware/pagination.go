package middleware

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func Pagination(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Set default values
		page := 1
		perPage := 10

		// Parse query parameters
		// Parse and validate the "page" query parameter
		if p := c.QueryParam("page"); p != "" {
			if val, err := strconv.Atoi(p); err != nil {
				// Log or handle invalid "page" parameter
				return echo.NewHTTPError(400, "Invalid 'page' parameter")
			} else if val > 0 {
				page = val
			}
		}

		// Parse and validate the "per_page" query parameter
		if pp := c.QueryParam("per_page"); pp != "" {
			if val, err := strconv.Atoi(pp); err != nil {
				// Log or handle invalid "per_page" parameter
				return echo.NewHTTPError(400, "Invalid 'per_page' parameter")
			} else if val > 0 && val <= 100 {
				perPage = val
			}
		}

		// Store values in context
		c.Set("page", page)
		c.Set("per_page", perPage)

		return next(c)
	}
}
