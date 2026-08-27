package kavenegar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	domainsms "github.com/rahil-gallery/rahil-gallery-server/internal/domain/sms"
)

const (
	baseURL       = "https://api.kavenegar.com/v1"
	maxSendPerReq = 200 // Kavenegar send limit per call
)

type Client struct {
	apiKey string
	http   *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: strings.TrimSpace(apiKey),
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

var _ domainsms.Provider = (*Client)(nil)

type apiResponse struct {
	Return struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	} `json:"return"`
	Entries json.RawMessage `json:"entries"`
}

type sendEntry struct {
	MessageID int64  `json:"messageid"`
	Status    int    `json:"status"`
	Receptor  string `json:"receptor"`
}

// SendBulk uses sms/send.json (same message, many receptors).
// Caller should chunk; this method also refuses >200.
func (c *Client) SendBulk(ctx context.Context, sender string, receptors []string, message string) (domainsms.BulkResult, error) {
	if len(receptors) == 0 {
		return domainsms.BulkResult{}, nil
	}
	if len(receptors) > maxSendPerReq {
		return domainsms.BulkResult{}, fmt.Errorf("%w: max %d receptors per request", domainsms.ErrInvalidReceptor, maxSendPerReq)
	}
	form := url.Values{}
	form.Set("receptor", strings.Join(receptors, ","))
	form.Set("sender", sender)
	form.Set("message", message)
	body, err := c.postForm(ctx, "sms/send.json", form)
	if err != nil {
		return domainsms.BulkResult{}, err
	}
	var entries []sendEntry
	_ = json.Unmarshal(body.Entries, &entries)
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		ids = append(ids, fmt.Sprintf("%d", e.MessageID))
	}
	return domainsms.BulkResult{
		Accepted:   len(entries),
		MessageIDs: ids,
		RawStatus:  fmt.Sprintf("%d:%s", body.Return.Status, body.Return.Message),
	}, nil
}

// SendLookup uses verify/lookup.json (birthday templates).
// Pass tokens["token"] and/or tokens["token10"] depending on your template.
func (c *Client) SendLookup(ctx context.Context, receptor, template string, tokens map[string]string) error {
	form := url.Values{}
	form.Set("receptor", receptor)
	form.Set("template", template)
	if v := strings.TrimSpace(tokens["token"]); v != "" {
		form.Set("token", v)
	}
	if v := strings.TrimSpace(tokens["token10"]); v != "" {
		form.Set("token10", v)
	}
	// Kavenegar requires at least token; if only token10 is set, also set token to a safe fallback.
	if form.Get("token") == "" {
		if t10 := form.Get("token10"); t10 != "" {
			// token cannot contain spaces — strip for the required field
			form.Set("token", strings.ReplaceAll(t10, " ", ""))
		} else {
			form.Set("token", "مشتری")
		}
	}
	_, err := c.postForm(ctx, "verify/lookup.json", form)
	return err
}

func (c *Client) postForm(ctx context.Context, methodPath string, form url.Values) (*apiResponse, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", baseURL, c.apiKey, methodPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainsms.ErrProviderFailed, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %v", domainsms.ErrProviderFailed, err)
	}
	var parsed apiResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", domainsms.ErrProviderFailed, err)
	}
	if parsed.Return.Status != 200 {
		return nil, fmt.Errorf("%w: status=%d message=%s", domainsms.ErrProviderFailed, parsed.Return.Status, parsed.Return.Message)
	}
	return &parsed, nil
}
