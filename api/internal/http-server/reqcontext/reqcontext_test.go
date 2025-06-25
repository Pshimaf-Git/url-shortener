package reqcontext

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	c := New(w, r)

	assert.NotNil(t, c)
	assert.Equal(t, w, c.w)
	assert.Equal(t, r, c.r)
	assert.Equal(t, r.Body, c.body)
}

func TestRequest(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)
		assert.Equal(t, r, c.Request())   // compare pointers
		assert.Equal(t, *r, *c.Request()) // compare values
	})
}

func TestResponceWriter(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)
		assert.Equal(t, w, c.ResponceWriter())
	})
}

func TestHeader(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)

		assert.Equal(t, w.Header(), c.Header())
	})
}

func TestWriteHeader(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)
		c.WriteHeader(http.StatusAccepted)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})
}

func TestContext(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)
		assert.Equal(t, r.Context(), c.Context())
	})
}

func TestProto(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)

		assert.Equal(t, r.Proto, c.Proto())
	})
}

func TestProtoMajor(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)

		assert.Equal(t, r.ProtoMajor, c.ProtoMajor())
	})
}

func TestProtoMinor(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		c := New(w, r)

		assert.Equal(t, r.ProtoMinor, c.ProtoMinor())
	})
}

func TestPattern(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		pattern := "/blabla"
		router := chi.NewRouter()

		r := httptest.NewRequest(http.MethodOptions, pattern, nil)
		w := httptest.NewRecorder()

		router.Get(pattern, func(w http.ResponseWriter, r *http.Request) {
			c := New(w, r)
			assert.Equal(t, r.Pattern, c.Pattern())
		})

		router.ServeHTTP(w, r)
	})
}

func TestURL(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		expected string
	}{
		{
			name:     "root path",
			inputURL: "/",
			expected: "/",
		},
		{
			name:     "with query",
			inputURL: "/path?key=value",
			expected: "/path?key=value",
		},
		{
			name:     "with fragment",
			inputURL: "/path#section",
			expected: "/path%23section", // in url symbol `#` coding as %23
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.inputURL, nil)

			c := New(w, r)
			assert.Equal(t, tt.expected, c.URL().String())
		})
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		expected url.Values
	}{
		{
			name:     "no query",
			inputURL: "/",
			expected: url.Values{},
		},
		{
			name:     "single query",
			inputURL: "/?key=value",
			expected: url.Values{"key": []string{"value"}},
		},
		{
			name:     "multiple queries",
			inputURL: "/?key1=value1&key2=value2",
			expected: url.Values{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.inputURL, nil)

			c := New(w, r)
			assert.Equal(t, tt.expected, c.Query())
		})
	}
}

func TestGetParam(t *testing.T) {
	tests := []struct {
		name     string
		inputURL string
		key      string
		expected string
	}{
		{
			name:     "existing param",
			inputURL: "/?key=value",
			key:      "key",
			expected: "value",
		},
		{
			name:     "missing param",
			inputURL: "/",
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.inputURL, nil)

			c := New(w, r)
			assert.Equal(t, tt.expected, c.GetParam(tt.key))
		})
	}
}

func TestGetChiParamFromCtx(t *testing.T) {
	tests := []struct {
		name          string
		routePattern  string
		requestPath   string
		paramName     string
		expectedValue string
		setupContext  func(context.Context) context.Context // Optional context setup
	}{
		{
			name:          "basic chi param from context",
			routePattern:  "/users/{id}",
			requestPath:   "/users/123",
			paramName:     "id",
			expectedValue: "123",
		},
		{
			name:          "param not in context",
			routePattern:  "/posts/{slug}",
			requestPath:   "/posts/hello-world",
			paramName:     "missing",
			expectedValue: "",
		},
		{
			name:          "with additional context values",
			routePattern:  "/orgs/{orgID}/teams/{teamID}",
			requestPath:   "/orgs/42/teams/7",
			paramName:     "teamID",
			expectedValue: "7",
			setupContext: func(ctx context.Context) context.Context {
				// Add additional context values if needed
				return context.WithValue(ctx, "requestID", "abc123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get(tt.routePattern, func(w http.ResponseWriter, r *http.Request) {
				// Apply context setup if provided
				if tt.setupContext != nil {
					r = r.WithContext(tt.setupContext(r.Context()))
				}

				c := New(w, r)
				value := c.GetChiParamFromCtx(tt.paramName)
				assert.Equal(t, tt.expectedValue, value)
			})

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
		})
	}
}

func TestGetChiParam(t *testing.T) {
	tests := []struct {
		name          string
		routePattern  string
		requestPath   string
		paramName     string
		expectedValue string
	}{
		{
			name:          "simple param",
			routePattern:  "/users/{id}",
			requestPath:   "/users/123",
			paramName:     "id",
			expectedValue: "123",
		},
		{
			name:          "multiple params",
			routePattern:  "/posts/{year}/{month}",
			requestPath:   "/posts/2023/10",
			paramName:     "month",
			expectedValue: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get(tt.routePattern, func(w http.ResponseWriter, r *http.Request) {
				c := New(w, r)
				assert.Equal(t, tt.expectedValue, c.GetChiParam(tt.paramName))
			})

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
		})
	}
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
			req := httptest.NewRequest("GET", "/", nil)

			c := New(rr, req)
			c.JSON(tt.expectedStatus, tt.inputData)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
		})
	}
}

func TestXML(t *testing.T) {
	tests := []struct {
		name           string
		inputData      interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "simple object",
			inputData: struct {
				XMLName xml.Name `xml:"person"`
				Name    string   `xml:"name"`
				Age     int      `xml:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `<person><name>John</name><age>30</age></person>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			c := New(rr, req)
			c.XML(tt.expectedStatus, tt.inputData)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/xml", rr.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectedBody, strings.TrimSpace(rr.Body.String()))
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
			req := httptest.NewRequest("GET", "/", nil)

			c := New(rr, req)
			c.FORM(tt.expectedStatus, tt.inputData)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/x-www-form-urlencoded", rr.Header().Get("Content-Type"))

			// Form encoding order isn't guaranteed, so we need to parse and compare
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

func TestDecodeXML(t *testing.T) {
	tests := []struct {
		name          string
		inputBody     string
		target        interface{}
		expectedError bool
		expectedValue interface{}
	}{
		{
			name:      "valid xml",
			inputBody: `<person><name>John</name><age>30</age></person>`,
			target: &struct {
				XMLName xml.Name `xml:"person"`
				Name    string   `xml:"name"`
				Age     int      `xml:"age"`
			}{},
			expectedError: false,
			expectedValue: &struct {
				XMLName xml.Name `xml:"person"`
				Name    string   `xml:"name"`
				Age     int      `xml:"age"`
			}{
				Name: "John",
				Age:  30,
			},
		},
		{
			name:          "invalid xml",
			inputBody:     `<person><name>John</name><age>30</person>`, // Missing closing age tag
			target:        &struct{}{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-Type", "application/xml")

			c := New(rr, req)
			err := c.DecodeXML(tt.target)

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
			inputBody:     "name=John&age=thirty", // age should be int
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
