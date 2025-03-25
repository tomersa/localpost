package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ParseRequest reads and parses a request YAML file into a Request struct.
func ParseRequest(filePath string) (Request, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Request{}, fmt.Errorf("error reading %s: %v", filePath, err)
	}

	var req Request
	if err := yaml.Unmarshal(data, &req); err != nil {
		return Request{}, fmt.Errorf("error parsing %s: %v", filePath, err)
	}

	if req.URL == "" {
		return Request{}, fmt.Errorf("url is required in %s", filePath)
	}

	return req, nil
}

// ExecuteRequest executes an HTTP request and returns a Response struct.
func ExecuteRequest(req Request) (Response, error) {
	env, err := LoadEnv()
	if err != nil {
		return Response{}, fmt.Errorf("error loading env: %v", err)
	}

	finalURL, err := replacePlaceholders(req.URL, env.Vars)
	if err != nil {
		return Response{}, err
	}

	if !strings.HasPrefix(finalURL, "http://") && !strings.HasPrefix(finalURL, "https://") {
		return Response{}, fmt.Errorf("invalid URL: %s (must resolve to http:// or https://)", finalURL)
	}

	for key, value := range req.Headers {
		req.Headers[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in header %s: %v", key, err)
		}
	}

	for key, value := range req.Body.Json {
		if strVal, ok := value.(string); ok {
			req.Body.Json[key], err = replacePlaceholders(strVal, env.Vars)
			if err != nil {
				return Response{}, fmt.Errorf("error replacing placeholders in JSON body %s: %v", key, err)
			}
		}
	}
	for key, value := range req.Body.FormUrlEncoded {
		req.Body.FormUrlEncoded[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in form-urlencoded %s: %v", key, err)
		}
	}
	for key, value := range req.Body.Form.Fields {
		req.Body.Form.Fields[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in form fields %s: %v", key, err)
		}
	}
	for key, filePath := range req.Body.Form.Files {
		req.Body.Form.Files[key], err = replacePlaceholders(filePath, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in form files %s: %v", key, err)
		}
	}
	req.Body.Text, err = replacePlaceholders(req.Body.Text, env.Vars)
	if err != nil {
		return Response{}, fmt.Errorf("error replacing placeholders in text body: %v", err)
	}

	var body io.Reader
	contentType := req.Headers["Content-Type"]
	var reqBody string

	switch contentType {
	case "application/json", "":
		if len(req.Body.Json) > 0 {
			bodyBytes, err := json.Marshal(req.Body.Json)
			if err != nil {
				return Response{}, fmt.Errorf("error marshaling JSON body: %v", err)
			}
			reqBody = string(bodyBytes)
			body = strings.NewReader(reqBody)
			if contentType == "" {
				contentType = "application/json"
			}
		}
	case "application/x-www-form-urlencoded":
		if len(req.Body.FormUrlEncoded) > 0 {
			data := make(url.Values)
			for k, v := range req.Body.FormUrlEncoded {
				data.Set(k, v)
			}
			reqBody = data.Encode()
			body = strings.NewReader(reqBody)
		}
	case "multipart/form-data":
		if len(req.Body.Form.Fields) > 0 || len(req.Body.Form.Files) > 0 {
			bodyBuffer := &bytes.Buffer{}
			writer := multipart.NewWriter(bodyBuffer)

			for k, v := range req.Body.Form.Fields {
				writer.WriteField(k, v)
			}
			for k, filePath := range req.Body.Form.Files {
				file, err := os.Open(filePath)
				if err != nil {
					return Response{}, fmt.Errorf("error opening file %s: %v", filePath, err)
				}
				defer file.Close()
				part, err := writer.CreateFormFile(k, filepath.Base(filePath))
				if err != nil {
					return Response{}, fmt.Errorf("error creating form file %s: %v", k, err)
				}
				_, err = io.Copy(part, file)
				if err != nil {
					return Response{}, fmt.Errorf("error writing file %s to form: %v", k, err)
				}
			}
			err := writer.Close()
			if err != nil {
				return Response{}, fmt.Errorf("error closing form writer: %v", err)
			}
			reqBody = bodyBuffer.String()
			body = bodyBuffer
			contentType = writer.FormDataContentType()
		}
	case "text/plain":
		if req.Body.Text != "" {
			reqBody = req.Body.Text
			body = strings.NewReader(reqBody)
		}
	default:
		if len(req.Body.Json) > 0 || len(req.Body.FormUrlEncoded) > 0 || len(req.Body.Form.Fields) > 0 || len(req.Body.Form.Files) > 0 || req.Body.Text != "" {
			return Response{}, fmt.Errorf("unsupported or missing Content-Type for body: %s", contentType)
		}
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest(req.Method, finalURL, body)
	if err != nil {
		return Response{}, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Add stored cookies to the request
	if len(env.Cookies) > 0 {
		cookieHeader := ""
		for name, value := range env.Cookies {
			if cookieHeader != "" {
				cookieHeader += "; "
			}
			cookieHeader += fmt.Sprintf("%s=%s", name, value)
		}
		httpReq.Header.Set("Cookie", cookieHeader)
	}

	start := time.Now()
	resp, err := client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()
	duration := time.Since(start)

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("error reading response: %v", err)
	}
	respBody := string(respBodyBytes)

	req.URL = finalURL

	// Process set-env-var if present
	if len(req.SetEnv) > 0 {
		for varName, source := range req.SetEnv {
			var value string
			if source.Header != "" {
				if val, ok := resp.Header[source.Header]; ok && len(val) > 0 {
					value = val[0]
				}
			} else if source.Body != "" {
				var data map[string]interface{}
				if err := json.Unmarshal(respBodyBytes, &data); err == nil {
					if val, ok := data[source.Body]; ok {
						value = fmt.Sprintf("%v", val)
					}
				}
			}
			if value != "" {
				if err := SetEnvVar(varName, value); err != nil {
					return Response{}, fmt.Errorf("error setting env var %s: %v", varName, err)
				}
			}
		}
	}

	// Process Set-Cookie headers and save to config.yaml
	if setCookies, ok := resp.Header["Set-Cookie"]; ok && len(setCookies) > 0 {
		for _, cookie := range setCookies {
			parts := strings.SplitN(cookie, ";", 2)
			if len(parts) > 0 {
				kv := strings.SplitN(parts[0], "=", 2)
				if len(kv) == 2 {
					name := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					if err := SetCookie(name, value); err != nil {
						return Response{}, fmt.Errorf("error setting cookie %s: %v", name, err)
					}
				}
			}
		}
	}

	return Response{
		ReqURL:      req.URL,
		ReqHeaders:  req.Headers,
		ReqBody:     reqBody,
		Status:      resp.Status,
		RespHeaders: resp.Header,
		RespBody:    respBody,
		Duration:    duration,
	}, nil
}
