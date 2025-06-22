package sl

import "log/slog"

// Error returns slog.Attr where key = ""error" and value = slog.StringValue(err.Error())
func Error(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
