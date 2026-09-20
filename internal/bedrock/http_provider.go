package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultHTTPProviderResponseLimit = 4 * 1024 * 1024

// OpenAICompatibleProvider invokes a configured OpenAI-compatible chat-completions endpoint.
// It intentionally depends only on the wire contract, not on any hosted vendor.
type OpenAICompatibleProvider struct {
	Endpoint    string
	Model       string
	APIKey      string
	Timeout     time.Duration
	MaxResponse int64
	Client      *http.Client
}

func (p OpenAICompatibleProvider) Name() string { return "openai-compatible:" + p.Model }

func (p OpenAICompatibleProvider) Execute(ctx context.Context, req ProviderRequest) (ProviderResponse, error) {
	endpoint, err := validateProviderEndpoint(p.Endpoint)
	if err != nil {
		return ProviderResponse{}, err
	}
	if strings.TrimSpace(p.Model) == "" {
		return ProviderResponse{}, errors.New("provider model is required")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("encode provider request: %w", err)
	}
	wireRequest := struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}{Model: p.Model}
	wireRequest.Messages = append(wireRequest.Messages,
		struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{"system", "Return only one JSON object matching {\"summary\":string,\"changes\":[{\"path\":string,\"content\":string}]}. Do not use Markdown fences."},
		struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{"user", string(payload)},
	)
	body, err := json.Marshal(wireRequest)
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("encode HTTP provider request: %w", err)
	}

	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(runCtx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("create HTTP provider request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		if runCtx.Err() != nil {
			if ctx.Err() != nil {
				return ProviderResponse{}, fmt.Errorf("provider canceled: %w", ctx.Err())
			}
			return ProviderResponse{}, fmt.Errorf("provider timed out: %w", runCtx.Err())
		}
		return ProviderResponse{}, fmt.Errorf("provider transport failed: %w", err)
	}
	defer resp.Body.Close()

	limit := p.MaxResponse
	if limit <= 0 {
		limit = defaultHTTPProviderResponseLimit
	}
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("read provider response: %w", err)
	}
	if int64(len(responseBody)) > limit {
		return ProviderResponse{}, fmt.Errorf("provider response exceeded %d bytes", limit)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail := strings.TrimSpace(string(responseBody))
		if len(detail) > 1024 {
			detail = detail[:1024] + "..."
		}
		detail = redactValues(detail, []string{p.APIKey})
		return ProviderResponse{}, fmt.Errorf("provider HTTP %s: %s", resp.Status, detail)
	}

	var wireResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	if err := dec.Decode(&wireResponse); err != nil {
		return ProviderResponse{}, fmt.Errorf("decode HTTP provider response: %w", err)
	}
	var wireExtra any
	if err := dec.Decode(&wireExtra); !errors.Is(err, io.EOF) {
		if err == nil {
			return ProviderResponse{}, errors.New("HTTP provider response contained multiple JSON values")
		}
		return ProviderResponse{}, fmt.Errorf("HTTP provider response contained trailing data: %w", err)
	}
	if len(wireResponse.Choices) == 0 || strings.TrimSpace(wireResponse.Choices[0].Message.Content) == "" {
		return ProviderResponse{}, errors.New("provider returned no assistant content")
	}
	var result ProviderResponse
	resultDecoder := json.NewDecoder(strings.NewReader(wireResponse.Choices[0].Message.Content))
	resultDecoder.DisallowUnknownFields()
	if err := resultDecoder.Decode(&result); err != nil {
		return ProviderResponse{}, fmt.Errorf("decode provider result: %w", err)
	}
	var extra any
	if err := resultDecoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return ProviderResponse{}, errors.New("provider result contained multiple JSON values")
		}
		return ProviderResponse{}, fmt.Errorf("provider result contained trailing data: %w", err)
	}
	return result, nil
}

func validateProviderEndpoint(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("provider endpoint is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid provider endpoint: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("provider endpoint must use http or https")
	}
	if u.Host == "" {
		return nil, errors.New("provider endpoint must include a host")
	}
	if u.User != nil {
		return nil, errors.New("provider endpoint must not contain credentials")
	}
	return u, nil
}
