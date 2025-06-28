package reqcontext

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestContext() (*ReqContext, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("chiParam", "chiValue")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return New(w, req), w
}

func TestCloseBody(t *testing.T) {
	c, _ := setupTestContext()
	err := c.CloseBody()
	assert.NoError(t, err)
}

func TestCloseBodyError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Body = &mockReadCloser{closeErr: errors.New("close error")}

	c := New(httptest.NewRecorder(), req)
	err := c.CloseBody()
	assert.Error(t, err)
}

type mockReadCloser struct {
	closeErr error
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) { return 0, io.EOF }
func (m *mockReadCloser) Close() error                     { return m.closeErr }

func TestURL(t *testing.T) {
	c, _ := setupTestContext()
	u := c.URL()
	require.NotNil(t, u)
	assert.Equal(t, "/test?param=value", u.RequestURI())
}

func TestQuery(t *testing.T) {
	c, _ := setupTestContext()
	query := c.Query()
	require.NotNil(t, query)
	assert.Equal(t, "value", query.Get("param"))
}

func TestGetParam(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, "value", c.GetParam("param"))
	assert.Empty(t, c.GetParam("nonexistent"))
}

func TestGetChiParam(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, "chiValue", c.GetChiParam("chiParam"))
	assert.Empty(t, c.GetChiParam("nonexistent"))
}

func TestGetChiParamFromCtx(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, "chiValue", c.GetChiParamFromCtx("chiParam"))
	assert.Empty(t, c.GetChiParamFromCtx("nonexistent"))
}

func TestRequest(t *testing.T) {
	c, _ := setupTestContext()
	req := c.Request()
	require.NotNil(t, req)
	assert.Equal(t, http.MethodGet, req.Method)
}

func TestResponceWriter(t *testing.T) {
	c, w := setupTestContext()
	rw := c.ResponceWriter()
	require.NotNil(t, rw)
	assert.Equal(t, w, rw)
}

func TestRequestID(t *testing.T) {
	c, _ := setupTestContext()

	ctx := context.WithValue(c.Context(), middleware.RequestIDKey, "test-id")
	c.r = c.r.WithContext(ctx)
	assert.Equal(t, "test-id", c.RequestID())
}

func TestHeader(t *testing.T) {
	c, _ := setupTestContext()
	headers := c.Header()
	require.NotNil(t, headers)
	headers.Set("Test-Header", "value")
	assert.Equal(t, "value", headers.Get("Test-Header"))
}

func TestContext(t *testing.T) {
	c, _ := setupTestContext()
	ctx := c.Context()
	require.NotNil(t, ctx)
	assert.Equal(t, c.r.Context(), ctx)
}

func TestWriteHeader(t *testing.T) {
	c, w := setupTestContext()
	c.WriteHeader(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, w.Code)
}

func TestDeadline(t *testing.T) {
	c, _ := setupTestContext()
	deadline, ok := c.Deadline()
	assert.False(t, ok)
	assert.True(t, deadline.IsZero())
}

func TestDone(t *testing.T) {
	c, _ := setupTestContext()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.r = c.r.WithContext(ctx)

	done := c.Done()
	require.NotNil(t, done, "Done channel should not be nil")

	select {
	case <-done:
		t.Error("Done channel should not be closed yet")
	default:

	}

	cancel()
	select {
	case <-done:

	case <-time.After(100 * time.Millisecond):
		t.Error("Done channel should be closed after cancellation")
	}
}

func TestForm(t *testing.T) {
	t.Run("nil for GET requests", func(t *testing.T) {
		c, _ := setupTestContext()
		form := c.Form()
		assert.Nil(t, form, "Form should be nil for GET requests")
	})

	t.Run("populated for POST requests", func(t *testing.T) {

		body := strings.NewReader("key=value")
		req := httptest.NewRequest("POST", "/", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		err := req.ParseForm()
		require.NoError(t, err)

		c := New(w, req)
		form := c.Form()
		require.NotNil(t, form, "Form should not be nil after ParseForm")
		assert.Equal(t, "value", form.Get("key"))
	})
}

func TestErr(t *testing.T) {
	c, _ := setupTestContext()
	assert.Nil(t, c.Err())
}

func TestValue(t *testing.T) {
	c, _ := setupTestContext()
	key := "test-key"
	val := "test-value"
	ctx := context.WithValue(c.Context(), key, val)
	c.r = c.r.WithContext(ctx)
	assert.Equal(t, val, c.Value(key))
}

func TestBody(t *testing.T) {
	c, _ := setupTestContext()
	body := c.Body()
	require.NotNil(t, body)
}

func TestMethod(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, http.MethodGet, c.Method())
}

func TestURI(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, "/test?param=value", c.URI())
}

func TestUserAgent(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, "", c.UserAgent())
}

func TestHost(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, "example.com", c.Host())
}

func TestTLS(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := New(w, r)
		assert.NotNil(t, c.TLS())
	}))
	defer ts.Close()

	client := ts.Client()
	_, err := client.Get(ts.URL)
	assert.NoError(t, err)
}

func TestProto(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, "HTTP/1.1", c.Proto())
}

func TestProtoMajor(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, 1, c.ProtoMajor())
}

func TestProtoMinor(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, 1, c.ProtoMinor())
}

func TestContentLength(t *testing.T) {
	c, _ := setupTestContext()
	assert.Equal(t, int64(0), c.ContentLength())
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name           string
		inputData      interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "simple object",
			inputData: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"name":"John","age":30}` + "\n",
		},
		{
			name:           "empty object",
			inputData:      struct{}{},
			expectedStatus: http.StatusCreated,
			expectedBody:   "{}\n",
		},
		{
			name: "hard object",
			inputData: struct {
				Name string `json:"name"`
				Addr struct {
					City    string `json:"city"`
					Country string `json:"country"`
				} `json:"user_address"`
				Hobbies []struct {
					Hobby string `json:"hobby"`
					Num   int    `json:"num"`
				} `json:"hobbies"`
			}{
				Name: "John",
				Addr: struct {
					City    string `json:"city"`
					Country string `json:"country"`
				}{
					City:    "Moscow",
					Country: "Russia",
				},
				Hobbies: []struct {
					Hobby string `json:"hobby"`
					Num   int    `json:"num"`
				}{
					{
						Hobby: "programing",
						Num:   1000,
					},
					{
						Hobby: "art",
						Num:   1,
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"name":"John","user_address":{"city":"Moscow","country":"Russia"},"hobbies":[{"hobby":"programing","num":1000},{"hobby":"art","num":1}]}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			c := New(rr, req)
			c.JSON(tt.expectedStatus, tt.inputData)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
		})
	}
}

func TestXML(t *testing.T) {
	type testStruct struct {
		XMLName xml.Name `xml:"test"`
		Field   string   `xml:"field"`
	}

	tests := []struct {
		name        string
		input       interface{}
		status      int
		expectXML   string
		expectError bool
	}{
		{
			name:      "simple struct",
			input:     testStruct{Field: "value"},
			status:    http.StatusOK,
			expectXML: `<test><field>value</field></test>`, // Note: no newline
		},
		{
			name:      "empty struct",
			input:     testStruct{},
			status:    http.StatusOK,
			expectXML: `<test><field></field></test>`, // Empty field is still present
		},
		{
			name:        "invalid type",
			input:       make(chan int),
			status:      http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			c := New(w, req)

			c.XML(tt.status, tt.input)

			if tt.expectError {
				assert.NotEqual(t, http.StatusOK, w.Code)
				return
			}

			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectXML, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestFORM(t *testing.T) {
	tests := []struct {
		name           string
		inputData      interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "simple form data",
			inputData: struct {
				Name string `form:"name"`
				Age  int    `form:"age"`
			}{
				Name: "Alice",
				Age:  25,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "age=25&name=Alice",
		},
		{
			name: "nested form data",
			inputData: struct {
				User struct {
					Login string `form:"username"`
					ID    int    `form:"user_id"`
				} `form:"user"`
				Active bool `form:"is_active"`
			}{
				User: struct {
					Login string `form:"username"`
					ID    int    `form:"user_id"`
				}{
					Login: "admin",
					ID:    1,
				},
				Active: true,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "is_active=true&user.user_id=1&user.username=admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			c := New(rr, req)
			c.FORM(tt.expectedStatus, tt.inputData)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/x-www-form-urlencoded", rr.Header().Get("Content-Type"))

			expected, err := url.ParseQuery(tt.expectedBody)
			assert.NoError(t, err)
			actual, err := url.ParseQuery(rr.Body.String())
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name          string
		inputBody     string
		target        interface{}
		expectedError bool
		expectedValue interface{}
	}{
		{
			name:      "valid json",
			inputBody: `{"name":"John","age":30}`,
			target: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{},
			expectedError: false,
			expectedValue: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
		},
		{
			name:          "invalid json",
			inputBody:     `{"name": "John", "age": }`,
			target:        &struct{}{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.inputBody))

			c := New(rr, req)
			err := c.DecodeJSON(tt.target)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, tt.target)
			}
		})
	}
}

func TestDecodeForm(t *testing.T) {
	tests := []struct {
		name          string
		inputBody     string
		target        interface{}
		expectedError bool
		expectedValue interface{}
	}{
		{
			name:      "valid form data",
			inputBody: "name=John&age=30",
			target: &struct {
				Name string `form:"name"`
				Age  int    `form:"age"`
			}{},
			expectedError: false,
			expectedValue: &struct {
				Name string `form:"name"`
				Age  int    `form:"age"`
			}{
				Name: "John",
				Age:  30,
			},
		},
		{
			name:      "nested form data",
			inputBody: "user.name=Admin&user.id=1",
			target: &struct {
				User struct {
					Name string `form:"name"`
					ID   int    `form:"id"`
				} `form:"user"`
			}{},
			expectedError: false,
			expectedValue: &struct {
				User struct {
					Name string `form:"name"`
					ID   int    `form:"id"`
				} `form:"user"`
			}{
				User: struct {
					Name string `form:"name"`
					ID   int    `form:"id"`
				}{
					Name: "Admin",
					ID:   1,
				},
			},
		},
		{
			name:          "invalid form data",
			inputBody:     "name=John&age=thirty",
			target:        &struct{}{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			c := New(rr, req)
			err := c.DecodeForm(tt.target)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, tt.target)
			}
		})
	}
}

func TestDecodeXML(t *testing.T) {
	type testStruct struct {
		XMLName xml.Name `xml:"test"`
		Field   string   `xml:"field"`
	}

	tests := []struct {
		name        string
		input       string
		output      testStruct
		expectError bool
	}{
		{
			name:  "valid xml",
			input: `<test><field>value</field></test>`,
			output: testStruct{
				XMLName: xml.Name{Local: "test"},
				Field:   "value",
			},
		},
		{
			name:  "empty xml",
			input: `<test></test>`,
			output: testStruct{
				XMLName: xml.Name{Local: "test"},
			},
		},
		{
			name:        "malformed xml",
			input:       `<test><field>value</test>`,
			expectError: true,
		},
		{
			name:        "empty body",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.input)
			req := httptest.NewRequest("POST", "/", body)
			req.Header.Set("Content-Type", "application/xml")
			w := httptest.NewRecorder()
			c := New(w, req)

			var result testStruct
			err := c.DecodeXML(&result)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.output, result)
		})
	}
}
