package handlers

import (
	"bytes"
	"context"
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

type Context struct {
	w    http.ResponseWriter
	r    *http.Request
	body io.ReadCloser
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		w:    w,
		r:    r,
		body: r.Body,
	}
}

func (c *Context) URL() *url.URL                        { return c.r.URL }
func (c *Context) Query() url.Values                    { return c.URL().Query() }
func (c *Context) GetParam(key string) string           { return c.Query().Get(key) }
func (c *Context) GetChiParam(key string) string        { return chi.URLParam(c.r, key) }
func (c *Context) GetChiParamRfomCtx(key string) string { return chi.URLParamFromCtx(c.Context(), key) }

func (c *Context) JSON(status int, v any) {
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

func (c *Context) XML(status int, v any) {
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

func (c *Context) FORM(status int, v any) {
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

func (c *Context) Request() *http.Request              { return c.r }
func (c *Context) ResponceWriter() http.ResponseWriter { return c.w }

func (c *Context) RequestID() string           { return middleware.GetReqID(c.Context()) }
func (c *Context) Header() http.Header         { return c.w.Header() }
func (c *Context) Write(v []byte) (int, error) { return c.w.Write(v) }
func (c *Context) Context() context.Context    { return c.r.Context() }
func (c *Context) WriteHeader(status int)      { c.w.WriteHeader(status) }

func (c *Context) Deadline() (time.Time, bool) { return c.r.Context().Deadline() }
func (c *Context) Done() <-chan struct{}       { return c.r.Context().Done() }
func (c *Context) Err() error                  { return c.r.Context().Err() }
func (c *Context) Value(v any) any             { return c.r.Context().Value(v) }

func (c *Context) Body() io.ReadCloser { return c.body }
func (c *Context) DecodeJSON(v any) error {
	defer c.CloseBody()
	return json.NewDecoder(c.Body()).Decode(v)
}
func (c *Context) DecodeXML(v any) error {
	defer c.CloseBody()
	return xml.NewDecoder(c.Body()).Decode(v)
}
func (c *Context) DecodeForm(v any) error {
	defer c.CloseBody()
	return form.NewDecoder(c.Body()).Decode(v)
}

func (c *Context) CloseBody() error { return c.body.Close() }
