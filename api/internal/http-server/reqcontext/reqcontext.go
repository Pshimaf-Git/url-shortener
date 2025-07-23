package reqcontext

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ajg/form"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ReqContext struct {
	w    http.ResponseWriter
	r    *http.Request
	body io.ReadCloser
}

func New(w http.ResponseWriter, r *http.Request) *ReqContext {
	return &ReqContext{
		w:    w,
		r:    r,
		body: r.Body,
	}
}

func (c *ReqContext) URL() *url.URL                 { return c.r.URL }
func (c *ReqContext) Query() url.Values             { return c.URL().Query() }
func (c *ReqContext) GetParam(key string) string    { return c.Query().Get(key) }
func (c *ReqContext) GetChiParam(key string) string { return chi.URLParam(c.r, key) }
func (c *ReqContext) GetChiParamFromCtx(key string) string {
	return chi.URLParamFromCtx(c.Context(), key)
}

func (c *ReqContext) JSON(status int, v any) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(c.w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Header().Set("Content-Type", "application/json")
	c.WriteHeader(status)
	_, _ = c.Write(buf.Bytes())
}

func (c *ReqContext) XML(status int, v any) {
	buf := &bytes.Buffer{}
	enc := xml.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		http.Error(c.w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Header().Set("Content-Type", "application/xml")
	c.WriteHeader(status)
	_, _ = c.Write(buf.Bytes())
}

func (c *ReqContext) FORM(status int, v any) {
	buf := &bytes.Buffer{}
	enc := form.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		http.Error(c.w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	c.WriteHeader(status)
	_, _ = c.Write(buf.Bytes())
}

func (c *ReqContext) Request() *http.Request              { return c.r }
func (c *ReqContext) ResponceWriter() http.ResponseWriter { return c.w }

func (c *ReqContext) RequestID() string                  { return middleware.GetReqID(c.Context()) }
func (c *ReqContext) Header() http.Header                { return c.w.Header() }
func (c *ReqContext) Write(v []byte) (int, error)        { return c.w.Write(v) }
func (c *ReqContext) Context() context.Context           { return c.r.Context() }
func (c *ReqContext) WriteHeader(status int)             { c.w.WriteHeader(status) }
func (c *ReqContext) SetHeader(key string, value string) { c.w.Header().Set(key, value) }

func (c *ReqContext) Deadline() (time.Time, bool) { return c.r.Context().Deadline() }
func (c *ReqContext) Done() <-chan struct{}       { return c.r.Context().Done() }
func (c *ReqContext) Err() error                  { return c.r.Context().Err() }
func (c *ReqContext) Value(v any) any             { return c.r.Context().Value(v) }

func (c *ReqContext) Body() io.ReadCloser { return c.r.Body }
func (c *ReqContext) DecodeJSON(v any) error {
	defer c.CloseBody() //nolint:errcheck
	return json.NewDecoder(c.Body()).Decode(v)
}
func (c *ReqContext) DecodeXML(v any) error {
	defer c.CloseBody() //nolint:errcheck
	return xml.NewDecoder(c.Body()).Decode(v)
}
func (c *ReqContext) DecodeForm(v any) error {
	defer c.CloseBody() //nolint:errcheck
	return form.NewDecoder(c.Body()).Decode(v)
}

func (c *ReqContext) CloseBody() error { return c.body.Close() }

func (c *ReqContext) Form() url.Values          { return c.r.Form }
func (c *ReqContext) Method() string            { return c.r.Method }
func (c *ReqContext) URI() string               { return c.r.RequestURI }
func (c *ReqContext) UserAgent() string         { return c.r.UserAgent() }
func (c *ReqContext) Host() string              { return c.r.Host }
func (c *ReqContext) Pattern() string           { return c.r.Pattern }
func (c *ReqContext) TLS() *tls.ConnectionState { return c.r.TLS }
func (c *ReqContext) ContentLength() int64      { return c.r.ContentLength }

func (c *ReqContext) Proto() string   { return c.r.Proto }
func (c *ReqContext) ProtoMajor() int { return c.r.ProtoMajor }
func (c *ReqContext) ProtoMinor() int { return c.r.ProtoMinor }
