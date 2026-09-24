package filestore

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type uploadThingFileReq struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Type string `json:"type"`
}

type uploadThingPrepareReq struct {
	Files              []uploadThingFileReq `json:"files"`
	ContentDisposition string               `json:"contentDisposition"`
}

type uploadThingFileResp struct {
	URL      string            `json:"url"`
	Fields   map[string]string `json:"fields"`
	FileURL  string            `json:"fileUrl"`
	UfsURL   string            `json:"ufsUrl"`
	Key      string            `json:"key"`
	FileName string            `json:"fileName"`
}

func (s *Storage) uploadToUploadThing(ctx context.Context, rawToken, filename, contentType string, data []byte) (*Result, error) {
	apiKey := extractAPIKey(rawToken)
	if apiKey == "" {
		apiKey = rawToken
	}

	prepareReqBody := uploadThingPrepareReq{
		Files: []uploadThingFileReq{
			{
				Name: filename,
				Size: len(data),
				Type: contentType,
			},
		},
		ContentDisposition: "inline",
	}

	reqBytes, err := json.Marshal(prepareReqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal prepare upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.uploadthing.com/v6/uploadFiles", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("new prepare request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-uploadthing-api-key", apiKey)
	req.Header.Set("x-uploadthing-token", rawToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prepare upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("uploadthing prepare returned status %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read prepare response: %w", err)
	}

	var files []uploadThingFileResp
	if err := json.Unmarshal(respBody, &files); err != nil {
		var wrapper struct {
			Data []uploadThingFileResp `json:"data"`
		}
		if wrapErr := json.Unmarshal(respBody, &wrapper); wrapErr == nil && len(wrapper.Data) > 0 {
			files = wrapper.Data
		} else {
			return nil, fmt.Errorf("unmarshal prepare response: %w", err)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no upload targets returned from uploadthing")
	}

	target := files[0]
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, target.URL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create s3 put request: %w", err)
	}
	uploadReq.Header.Set("Content-Type", contentType)

	uploadResp, err := s.httpClient.Do(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("s3 upload failed: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
		return nil, fmt.Errorf("s3 upload returned status: %d", uploadResp.StatusCode)
	}

	finalURL := target.UfsURL
	if finalURL == "" {
		finalURL = target.FileURL
	}
	if finalURL == "" && target.Key != "" {
		finalURL = "https://utfs.io/f/" + target.Key
	}

	return &Result{
		URL:  finalURL,
		Key:  target.Key,
		Name: filename,
		Size: int64(len(data)),
	}, nil
}

func extractAPIKey(token string) string {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return token
	}
	var parsed struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.Unmarshal(decoded, &parsed); err == nil && parsed.APIKey != "" {
		return parsed.APIKey
	}
	return token
}
