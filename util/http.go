package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/briandowns/spinner"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bombsimon/jtd-infer-go"

	"gopkg.in/yaml.v3"
)

// ReadRequestDefinition reads and parses a request YAML file into a RequestDefinition struct.
func ReadRequestDefinition(fileName string) (RequestDefinition, error) {
	filePath := filepath.Join(RequestsDir, fileName+".yaml")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return RequestDefinition{}, fmt.Errorf("error reading %s: %v", fileName, err)
	}

	var req RequestDefinition
	if err := yaml.Unmarshal(data, &req); err != nil {
		return RequestDefinition{}, fmt.Errorf("error parsing %s: %v", fileName, err)
	}

	parts := strings.SplitN(fileName, "_", 2)
	if len(parts) != 2 {
		return RequestDefinition{}, fmt.Errorf("invalid request name format: %s (expected METHOD_name)", fileName)
	}
	req.Method = strings.ToUpper(parts[0])

	if req.URL == "" {
		return RequestDefinition{}, fmt.Errorf("url is required in %s", fileName)
	}

	return req, nil
}

// processResponse handles post-request actions like set-env-var.
func processResponse(reqDef RequestDefinition, resp Response) error {
	if len(reqDef.SetEnv) > 0 {
		for varName, source := range reqDef.SetEnv {
			var value string
			if source.Header != "" {
				if val, ok := resp.RespHeaders[source.Header]; ok && len(val) > 0 {
					value = val[0]
				}
			} else if source.Body != "" {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(resp.RespBody), &data); err == nil {
					if val, ok := data[source.Body]; ok {
						value = fmt.Sprintf("%v", val)
					}
				}
			}
			if value != "" {
				if err := SetEnvVar(varName, value); err != nil {
					return fmt.Errorf("error setting env var %s: %v", varName, err)
				}
			}
		}
	}

	// Update runtime cookies
	if setCookies, ok := resp.RespHeaders["Set-Cookie"]; ok && len(setCookies) > 0 {
		cookies, err := LoadCookies()
		if err != nil {
			return fmt.Errorf("error loading cookies: %v", err)
		}
		for _, cookie := range setCookies {
			parts := strings.SplitN(cookie, ";", 2)
			if len(parts) > 0 {
				kv := strings.SplitN(parts[0], "=", 2)
				if len(kv) == 2 {
					name := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					cookies[name] = value
				}
			}
		}
		if err := SaveCookies(cookies); err != nil {
			return fmt.Errorf("error saving cookies: %v", err)
		}
	}

	return nil
}

// executeHTTPRequest performs the actual HTTP request and returns the response.
func executeHTTPRequest(reqDef RequestDefinition, fileName string) (Response, error) {
	env, err := LoadEnv()
	if err != nil {
		return Response{}, fmt.Errorf("error loading env: %v", err)
	}
	cookies, err := LoadCookies()
	if err != nil {
		return Response{}, fmt.Errorf("error loading cookies: %v", err)
	}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Prefix = fileName + " "
	s.Start()

	// Replace placeholders in URL (including query params)
	finalURL, err := replacePlaceholders(reqDef.URL, env.Vars)
	if err != nil {
		s.Stop()
		return Response{}, fmt.Errorf("error replacing placeholders in URL: %v", err)
	}

	if !strings.HasPrefix(finalURL, "http://") && !strings.HasPrefix(finalURL, "https://") {
		s.Stop()
		return Response{}, fmt.Errorf("invalid URL after placeholder replacement: %s", finalURL)
	}

	for key, value := range reqDef.Headers {
		reqDef.Headers[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
			s.Stop()
			return Response{}, fmt.Errorf("error replacing placeholders in header %s: %v", key, err)
		}
	}

	var body io.Reader
	contentType := reqDef.Headers["Content-Type"]
	var reqBody string

	switch contentType {
	case "application/json", "":
		if len(reqDef.Body.Json) > 0 {
			for key, value := range reqDef.Body.Json {
				if strVal, ok := value.(string); ok {
					reqDef.Body.Json[key], err = replacePlaceholders(strVal, env.Vars)
					if err != nil {
						s.Stop()
						return Response{}, fmt.Errorf("error replacing placeholders in JSON body %s: %v", key, err)
					}
				}
			}
			bodyBytes, err := json.Marshal(reqDef.Body.Json)
			if err != nil {
				s.Stop()
				return Response{}, fmt.Errorf("error marshaling JSON body: %v", err)
			}
			reqBody = string(bodyBytes)
			body = strings.NewReader(reqBody)
			if contentType == "" {
				contentType = "application/json"
			}
		}
	case "application/x-www-form-urlencoded":
		if len(reqDef.Body.FormUrlEncoded) > 0 {
			data := make(url.Values)
			for k, v := range reqDef.Body.FormUrlEncoded {
				data.Set(k, v)
			}
			reqBody = data.Encode()
			body = strings.NewReader(reqBody)
		}
	case "multipart/form-data":
		if len(reqDef.Body.Form.Fields) > 0 || len(reqDef.Body.Form.Files) > 0 {
			bodyBuffer := &bytes.Buffer{}
			writer := multipart.NewWriter(bodyBuffer)

			for k, v := range reqDef.Body.Form.Fields {
				writer.WriteField(k, v)
			}
			for k, filePath := range reqDef.Body.Form.Files {
				file, err := os.Open(filePath)
				if err != nil {
					s.Stop()
					return Response{}, fmt.Errorf("error opening file %s: %v", filePath, err)
				}
				defer file.Close()
				part, err := writer.CreateFormFile(k, filepath.Base(filePath))
				if err != nil {
					s.Stop()
					return Response{}, fmt.Errorf("error creating form file %s: %v", k, err)
				}
				_, err = io.Copy(part, file)
				if err != nil {
					s.Stop()
					return Response{}, fmt.Errorf("error writing file %s to form: %v", k, err)
				}
			}
			err := writer.Close()
			if err != nil {
				s.Stop()
				return Response{}, fmt.Errorf("error closing form writer: %v", err)
			}
			reqBody = bodyBuffer.String()
			body = bodyBuffer
			contentType = writer.FormDataContentType()
		}
	case "text/plain":
		if reqDef.Body.Text != "" {
			reqBody = reqDef.Body.Text
			body = strings.NewReader(reqBody)
		}
	default:
		if len(reqDef.Body.Json) > 0 || len(reqDef.Body.FormUrlEncoded) > 0 || len(reqDef.Body.Form.Fields) > 0 || len(reqDef.Body.Form.Files) > 0 || reqDef.Body.Text != "" {
			s.Stop()
			return Response{}, fmt.Errorf("unsupported or missing Content-Type for body: %s", contentType)
		}
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest(reqDef.Method, finalURL, body)
	if err != nil {
		s.Stop()
		return Response{}, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range reqDef.Headers {
		httpReq.Header.Set(key, value)
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	if len(cookies) > 0 {
		cookieHeader := ""
		for name, value := range cookies {
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
		s.Stop()
		return Response{}, fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Stop()
		return Response{}, fmt.Errorf("error reading response: %v", err)
	}
	respBody := string(respBodyBytes)

	response := Response{
		ReqURL:      finalURL,
		ReqHeaders:  reqDef.Headers,
		ReqBody:     reqBody,
		StatusCode:  resp.StatusCode,
		RespHeaders: resp.Header,
		RespBody:    respBody,
		Duration:    duration,
	}

	// Auto-login if status matches env.Login.TriggeredBy
	if env.Login != nil {
		for _, status := range env.Login.TriggeredBy {
			if response.StatusCode == status {
				s.Stop()
				_, err := HandleRequest(env.Login.Request)
				if err != nil {
					return Response{}, fmt.Errorf("error executing login request %s: %v", env.Login.Request, err)
				}
				return executeHTTPRequest(reqDef, fileName)
			}
		}
	}

	// Generate JTD schema if inferSchema is true
	if response.StatusCode == 200 && strings.TrimSpace(respBody) != "" {
		var doc interface{}
		if err := json.Unmarshal([]byte(respBody), &doc); err == nil {
			schema := jtdinfer.InferStrings([]string{respBody}, jtdinfer.WithoutHints()).IntoSchema()
			schemaPath := filepath.Join(SchemasDir, fileName+".jtd.json")
			schemaBytes, err := json.MarshalIndent(schema, "", "  ")
			if err == nil {
				os.MkdirAll(filepath.Dir(schemaPath), 0755)
				os.WriteFile(schemaPath, schemaBytes, 0644)
			}
		}
	}

	s.Stop()
	return response, nil
}

// HandleRequest executes an HTTP request and returns a Response struct.
func HandleRequest(fileName string) (Response, error) {
	reqDef, err := ReadRequestDefinition(fileName)
	if err != nil {
		return Response{}, err
	}

	resp, err := executeHTTPRequest(reqDef, fileName)
	if err != nil {
		return Response{}, fmt.Errorf("error executing request: %v", err)
	}

	if err := processResponse(reqDef, resp); err != nil {
		return Response{}, err
	}

	return resp, nil
}
