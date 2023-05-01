package auth

import "github.com/labstack/echo/v4"

type AccessFunc func(echo.Context, []byte) bool

func All(checks ...AccessFunc) AccessFunc {
	return func(ctx echo.Context, b []byte) bool {
		for _, check := range checks {
			if !check(ctx, b) {
				return false
			}
		}
		return true
	}
}

func Any(checks ...AccessFunc) AccessFunc {
	return func(ctx echo.Context, b []byte) bool {
		for _, check := range checks {
			if check(ctx, b) {
				return true
			}
		}
		return false
	}
}

func (check AccessFunc) Or(other AccessFunc) AccessFunc {
	return func(ctx echo.Context, b []byte) bool {
		return check(ctx, b) || other(ctx, b)
	}
}

func (check AccessFunc) And(other AccessFunc) AccessFunc {
	return func(ctx echo.Context, b []byte) bool {
		return check(ctx, b) && other(ctx, b)
	}
}

func CheckAccess(c echo.Context, doc []byte, checks ...AccessFunc) bool {
	allowed := false
	for _, check := range checks {
		if check(c, doc) {
			allowed = true
			break
		}
	}
	return allowed
}
