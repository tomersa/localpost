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
	"slices"
	"strings"
	"time"

	jtdinfer "github.com/bombsimon/jtd-infer-go"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"
)

func formatJSON(body string, contentType string) string {
	if body == "" || contentType == "" || !strings.Contains(contentType, "application/json") {
		return body
	}
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(body), "", "  ")
	if err != nil {
		return body
	}
	lines := strings.Split(prettyJSON.String(), "\n")
	for i, line := range lines {
		lines[i] = "    " + line
	}
	return strings.Join(lines, "\n")
}

func readRequestDefinition(filePath string) (RequestDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return RequestDefinition{}, fmt.Errorf("request file %s not found", filePath)
	}

	var req RequestDefinition
	if err := yaml.Unmarshal(data, &req); err != nil {
		return RequestDefinition{}, fmt.Errorf("error parsing %s: %v", filePath, err)
	}

	// Set Method strictly from filename
	req.Method = strings.TrimSuffix(filepath.Base(filePath), ".yaml")

	// Set URL from directory path and BASE_URL from config
	if req.URL == "" {
		env, err := LoadEnv()
		if err != nil {
			return RequestDefinition{}, fmt.Errorf("error loading env: %v", err)
		}
		baseURL, ok := env.Vars["BASE_URL"]
		if !ok {
			return RequestDefinition{}, fmt.Errorf("BASE_URL not found. Please set it in config.yaml")
		}
		relPath, err := filepath.Rel(RequestsDir, filePath)
		if err != nil {
			return RequestDefinition{}, fmt.Errorf("error getting relative path for %s: %v", filePath, err)
		}
		urlPath := strings.TrimSuffix(relPath, filepath.Ext(relPath))
		if urlPath == "." {
			urlPath = ""
		} else {
			urlPath = strings.ReplaceAll(urlPath, string(os.PathSeparator), "/")
		}
		req.URL = baseURL + "/" + urlPath
	}

	return req, nil
}

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

	if setCookies, ok := resp.RespHeaders["Set-Cookie"]; ok && len(setCookies) > 0 {
		for _, cookie := range setCookies {
			parts := strings.SplitN(cookie, ";", 2)
			if len(parts) > 0 {
				kv := strings.SplitN(parts[0], "=", 2)
				if len(kv) == 2 {
					name := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					if err := SetCookie(name, value); err != nil {
						return fmt.Errorf("error setting cookie %s: %v", name, err)
					}
				}
			}
		}
	}

	return nil
}

func inferSchema(resp Response, filePath string) {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && strings.TrimSpace(resp.RespBody) != "" {
		var doc interface{}
		if err := json.Unmarshal([]byte(resp.RespBody), &doc); err == nil {
			schema := jtdinfer.InferStrings([]string{resp.RespBody}, jtdinfer.WithoutHints()).IntoSchema()
			dirPath := filepath.Dir(filePath)
			method := strings.TrimSuffix(filepath.Base(filePath), ".yaml")
			schemaPath := filepath.Join(dirPath, method+".jtd.json")
			schemaBytes, err := json.MarshalIndent(schema, "", "  ")
			if err == nil {
				os.MkdirAll(dirPath, 0755)
				os.WriteFile(schemaPath, schemaBytes, 0644)
			}
		}
	}
}

func executeHTTPRequest(reqDef RequestDefinition, filePath string, inferSchema bool, isRetry bool) (Response, error) {
	env, err := LoadEnv()
	if err != nil {
		return Response{}, fmt.Errorf("error loading env: %v", err)
	}
	cookies, err := LoadCookies()
	if err != nil {
		return Response{}, fmt.Errorf("error loading cookies: %v", err)
	}

	finalURL, err := replacePlaceholders(reqDef.URL, env.Vars)
	if err != nil {
		return Response{}, fmt.Errorf("error replacing placeholders in URL: %v", err)
	}

	if !strings.HasPrefix(finalURL, "http://") && !strings.HasPrefix(finalURL, "https://") {
		return Response{}, fmt.Errorf("invalid URL after placeholder replacement: %s", finalURL)
	}

	for key, value := range reqDef.Headers {
		reqDef.Headers[key], err = replacePlaceholders(value, env.Vars)
		if err != nil {
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
						return Response{}, fmt.Errorf("error replacing placeholders in JSON body %s: %v", key, err)
					}
				}
			}
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

	client := &http.Client{
		Timeout: time.Duration(env.Timeout) * time.Second,
	}
	httpReq, err := http.NewRequest(reqDef.Method, finalURL, body)
	if err != nil {
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

	resp, err := client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
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
	}

	if env.Login != nil && !isRetry {
		if slices.Contains(env.Login.TriggeredBy, response.StatusCode) {
			loginReqDef, err := readRequestDefinition(filepath.Join(RequestsDir, env.Login.Request))
			if err != nil {
				return executeHTTPRequest(loginReqDef, filePath, inferSchema, true)
			}
		}
	}

	return response, nil
}

func HandleRequest(filePath string, verbose, toInferSchema bool) (Response, error) {
	reqDef, err := readRequestDefinition(filePath)
	if err != nil {
		return Response{}, err
	}

	// Setup progress writer
	pw := progress.NewWriter()
	pw.SetAutoStop(false)
	pw.SetTrackerLength(25)
	pw.SetMessageLength(60)
	pw.SetStyle(progress.StyleDefault)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Visibility.Percentage = false
	pw.Style().Visibility.Value = false
	pw.Style().Visibility.TrackerOverall = false
	pw.Style().Visibility.Time = true

	tracker := &progress.Tracker{
		Message: fmt.Sprintf("%s idle", reqDef.URL),
		Total:   0,
	}
	pw.AppendTracker(tracker)

	go pw.Render()

	resp, err := executeHTTPRequest(reqDef, filePath, toInferSchema, false)
	if err != nil {
		pw.Stop()
		tracker.MarkAsErrored()
		return Response{}, fmt.Errorf("error executing request: %v", err)
	}

	if err := processResponse(reqDef, resp); err != nil {
		pw.Stop()
		tracker.MarkAsErrored()
		return Response{}, err
	}

	var statusColor text.Color
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		statusColor = text.FgGreen
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		statusColor = text.FgYellow
	case resp.StatusCode >= 500:
		statusColor = text.FgRed
	default:
		statusColor = text.FgWhite
	}

	tracker.Total = 100
	tracker.UpdateMessage(statusColor.Sprintf("%s %d", resp.ReqURL, resp.StatusCode))
	tracker.MarkAsDone()

	for pw.IsRenderInProgress() {
		if pw.LengthActive() == 0 {
			pw.Stop()
		}
		time.Sleep(time.Millisecond * 10)
	}

	if toInferSchema {
		inferSchema(resp, filePath)
	}

	reqContentType := ""
	respContentType := ""
	if len(resp.ReqHeaders["Content-Type"]) > 0 {
		reqContentType = strings.ToLower(strings.Split(resp.ReqHeaders["Content-Type"], ";")[0])
	}
	if len(resp.RespHeaders["Content-Type"]) > 0 {
		respContentType = strings.ToLower(strings.Split(resp.RespHeaders["Content-Type"][0], ";")[0])
	}
	reqBodyDisplay := formatJSON(resp.ReqBody, reqContentType)
	respBodyDisplay := formatJSON(resp.RespBody, respContentType)

	fmt.Println("----------------------------------------")
	if verbose {
		fmt.Println(color.CyanString("Request:"))
		fmt.Println(color.HiBlueString("  Headers:"))
		if resp.ReqHeaders == nil || len(resp.ReqHeaders) == 0 {
			fmt.Println(color.HiYellowString("    <Empty>"))
		}
		for k, v := range resp.ReqHeaders {
			fmt.Printf("    %s: %s\n", k, v)
		}
		fmt.Println(color.HiBlueString("  Body:"))
		if resp.ReqBody != "" {
			fmt.Printf("%s\n", reqBodyDisplay)
		} else {
			fmt.Println(color.HiYellowString("    <Empty>"))
		}
		fmt.Println(color.CyanString("Response:"))
		fmt.Println(color.HiBlueString("  Headers:"))
		if resp.RespHeaders == nil || len(resp.RespHeaders) == 0 {
			fmt.Println(color.HiYellowString("    <Empty>"))
		}
		for k, v := range resp.RespHeaders {
			for _, val := range v {
				fmt.Printf("    %s: %s\n", k, val)
			}
		}
		fmt.Println(color.HiBlueString("  Body:"))
	}
	if resp.RespBody != "" {
		fmt.Printf("%s\n", respBodyDisplay)
	} else {
		fmt.Println(color.HiYellowString("    <Empty>"))
	}

	return resp, nil
}
