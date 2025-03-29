package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	jtdinfer "github.com/bombsimon/jtd-infer-go"
	"github.com/briandowns/spinner"
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

	// Set default login status to 401 if not specified
	if req.Login != nil && len(req.Login.TriggeredBy) == 0 {
		req.Login.TriggeredBy = []int{401}
	}

	return req, nil
}

// executeHTTPRequest performs the actual HTTP request and returns the response.
func executeHTTPRequest(reqDef RequestDefinition, fileName string, showLog bool) (Response, error) {
	env, err := LoadEnv()
	if err != nil {
		return Response{}, fmt.Errorf("error loading env: %v", err)
	}

	// Start spinner
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Prefix = fileName + " "
	s.Start()

	finalURL, err := replacePlaceholders(reqDef.URL, env.Vars)
	if err != nil {
		s.Stop()
		return Response{}, err
	}

	if !strings.HasPrefix(finalURL, "http://") && !strings.HasPrefix(finalURL, "https://") {
		return Response{}, fmt.Errorf("invalid URL: %s (must resolve to http:// or https://)", finalURL)
	}

	for key, value := range reqDef.Headers {
		reqDef.Headers[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in header %s: %v", key, err)
		}
	}

	for key, value := range reqDef.Body.Json {
		if strVal, ok := value.(string); ok {
			reqDef.Body.Json[key], err = replacePlaceholders(strVal, env.Vars)
			if err != nil {
				return Response{}, fmt.Errorf("error replacing placeholders in JSON body %s: %v", key, err)
			}
		}
	}
	for key, value := range reqDef.Body.FormUrlEncoded {
		reqDef.Body.FormUrlEncoded[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in form-urlencoded %s: %v", key, err)
		}
	}
	for key, value := range reqDef.Body.Form.Fields {
		reqDef.Body.Form.Fields[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in form fields %s: %v", key, err)
		}
	}
	for key, filePath := range reqDef.Body.Form.Files {
		reqDef.Body.Form.Files[key], err = replacePlaceholders(filePath, env.Vars)
		if err != nil {
			return Response{}, fmt.Errorf("error replacing placeholders in form files %s: %v", key, err)
		}
	}

	reqDef.Body.Text, err = replacePlaceholders(reqDef.Body.Text, env.Vars)
	if err != nil {
		return Response{}, fmt.Errorf("error replacing placeholders in text body: %v", err)
	}

	var body io.Reader
	contentType := reqDef.Headers["Content-Type"]
	var reqBody string

	switch contentType {
	case "application/json", "":
		if len(reqDef.Body.Json) > 0 {
			bodyBytes, err := json.Marshal(reqDef.Body.Json)
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
		if reqDef.Body.Text != "" {
			reqBody = reqDef.Body.Text
			body = strings.NewReader(reqBody)
		}
	default:
		if len(reqDef.Body.Json) > 0 || len(reqDef.Body.FormUrlEncoded) > 0 || len(reqDef.Body.Form.Fields) > 0 || len(reqDef.Body.Form.Files) > 0 || reqDef.Body.Text != "" {
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

	// Stop spinner and print final status
	s.Stop()
	if showLog || resp.StatusCode >= 300 {
		fmt.Printf("%s %d\n", fileName, response.StatusCode)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 && strings.TrimSpace(respBody) != "" {
		schema := jtdinfer.InferStrings([]string{respBody}, jtdinfer.WithoutHints()).IntoSchema()
		schemaPath := filepath.Join("lpost/schemas", fileName+".jtd.json")
		schemaBytes, err := json.MarshalIndent(schema, "", "  ")
		if err == nil {
			os.MkdirAll(filepath.Dir(schemaPath), 0755)
			os.WriteFile(schemaPath, schemaBytes, 0644)
		}
	}

	return response, nil
}

// HandleRequest executes an HTTP request and returns a Response struct.
func HandleRequest(fileName string) (Response, error) {
	reqDef, err := ReadRequestDefinition(fileName)
	if err != nil {
		return Response{}, err
	}

	// Execute pre-flight request if specified
	if reqDef.PreFlight != "" {
		fmt.Printf("request: %s\n", reqDef.PreFlight)
		_, err := HandleRequest(reqDef.PreFlight)
		if err != nil {
			return Response{}, fmt.Errorf("error executing pre-flight request %s: %v", reqDef.PreFlight, err)
		}
	}

	// Try the main request
	resp, err := executeHTTPRequest(reqDef, fileName, false)
	if err != nil {
		return Response{}, fmt.Errorf("error executing request: %v", err)
	}

	// Check for specified status codes and handle login if specified
	if reqDef.Login != nil {
		for _, status := range reqDef.Login.TriggeredBy {
			if resp.StatusCode == status {
				_, err := HandleRequest(reqDef.Login.Request)
				if err != nil {
					return Response{}, fmt.Errorf("error executing login request %s: %v", reqDef.Login.Request, err)
				}

				resp, err = executeHTTPRequest(reqDef, fileName, true)
				if err != nil {
					return Response{}, fmt.Errorf("error retrying request after login: %v", err)
				}
				break
			}
		}
	}

	// Process set-env-var if present
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
					return Response{}, fmt.Errorf("error setting env var %s: %v", varName, err)
				}
			}
		}
	}

	// Process Set-Cookie headers and save to config.yaml
	if setCookies, ok := resp.RespHeaders["Set-Cookie"]; ok && len(setCookies) > 0 {
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

	// Execute post-flight request if specified
	if reqDef.PostFlight != "" {
		_, err := HandleRequest(reqDef.PostFlight)
		if err != nil {
			return Response{}, fmt.Errorf("error executing post-flight request %s: %v", reqDef.PostFlight, err)
		}
	}

	return resp, nil
}
