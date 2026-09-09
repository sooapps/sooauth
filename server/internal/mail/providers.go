package mail

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ProviderConfig struct {
	Provider     string // "smtp", "resend", "postmark", "ses"
	FromName     string
	FromEmail    string
	ReplyTo      string
	BrandName    string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPTLSMode  string // "starttls", "implicit", "none"
	APIKey       string
	AWSAccessKey string
	AWSSecretKey string
	AWSRegion    string

	// Optional overrides for testing
	BaseURL    string
	HTTPClient *http.Client
}

func formatFrom(name, email, defaultBrand string) string {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultBrand
	}
	if name != "" && !strings.Contains(email, "<") {
		return fmt.Sprintf("%s <%s>", name, email)
	}
	return email
}

func NewSender(cfg ProviderConfig) (Sender, error) {
	from := formatFrom(cfg.FromName, cfg.FromEmail, cfg.BrandName)
	if from == "" {
		from = "sooauth@localhost"
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "resend":
		if strings.TrimSpace(cfg.APIKey) == "" {
			return nil, fmt.Errorf("resend api_key required")
		}
		return &ResendSender{
			baseURL:    cfg.BaseURL,
			apiKey:     cfg.APIKey,
			from:       from,
			replyTo:    strings.TrimSpace(cfg.ReplyTo),
			httpClient: cfg.HTTPClient,
		}, nil

	case "postmark":
		if strings.TrimSpace(cfg.APIKey) == "" {
			return nil, fmt.Errorf("postmark server_token required")
		}
		return &PostmarkSender{
			baseURL:     cfg.BaseURL,
			serverToken: cfg.APIKey,
			from:        from,
			replyTo:     strings.TrimSpace(cfg.ReplyTo),
			httpClient:  cfg.HTTPClient,
		}, nil

	case "ses":
		if strings.TrimSpace(cfg.AWSAccessKey) == "" || strings.TrimSpace(cfg.AWSSecretKey) == "" {
			return nil, fmt.Errorf("aws ses credentials required")
		}
		region := strings.TrimSpace(cfg.AWSRegion)
		if region == "" {
			region = "us-east-1"
		}
		return &SESSender{
			endpoint:        cfg.BaseURL,
			accessKeyID:     cfg.AWSAccessKey,
			secretAccessKey: cfg.AWSSecretKey,
			region:          region,
			from:            from,
			replyTo:         strings.TrimSpace(cfg.ReplyTo),
			httpClient:      cfg.HTTPClient,
		}, nil

	case "smtp", "":
		portStr := "587"
		if cfg.SMTPPort > 0 {
			portStr = strconv.Itoa(cfg.SMTPPort)
		}
		return NewWithSettings(Settings{
			Host:      cfg.SMTPHost,
			Port:      portStr,
			User:      cfg.SMTPUser,
			Password:  cfg.SMTPPassword,
			From:      from,
			BrandName: cfg.FromName,
			TLSMode:   cfg.SMTPTLSMode,
		}), nil

	default:
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.Provider)
	}
}

type ResendSender struct {
	baseURL    string
	apiKey     string
	from       string
	replyTo    string
	httpClient *http.Client
}

func (r *ResendSender) SendOutbound(msg Outbound) error {
	payload := map[string]any{
		"from":    r.from,
		"to":      []string{msg.To},
		"subject": msg.Subject,
		"text":    msg.Plain,
	}
	if msg.HTML != "" {
		payload["html"] = msg.HTML
	}
	if r.replyTo != "" {
		payload["reply_to"] = r.replyTo
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	baseURL := r.baseURL
	if baseURL == "" {
		baseURL = "https://api.resend.com"
	}
	req, err := http.NewRequest("POST", strings.TrimRight(baseURL, "/")+"/emails", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := r.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend error (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type PostmarkSender struct {
	baseURL     string
	serverToken string
	from        string
	replyTo     string
	httpClient  *http.Client
}

func (p *PostmarkSender) SendOutbound(msg Outbound) error {
	payload := map[string]any{
		"From":     p.from,
		"To":       msg.To,
		"Subject":  msg.Subject,
		"TextBody": msg.Plain,
	}
	if msg.HTML != "" {
		payload["HtmlBody"] = msg.HTML
	}
	if p.replyTo != "" {
		payload["ReplyTo"] = p.replyTo
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	baseURL := p.baseURL
	if baseURL == "" {
		baseURL = "https://api.postmarkapp.com"
	}
	req, err := http.NewRequest("POST", strings.TrimRight(baseURL, "/")+"/email", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("X-Postmark-Server-Token", p.serverToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := p.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("postmark error (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type SESSender struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	region          string
	from            string
	replyTo         string
	httpClient      *http.Client
}

func (s *SESSender) SendOutbound(msg Outbound) error {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	payload := map[string]any{
		"FromEmailAddress": s.from,
		"Destination": map[string]any{
			"ToAddresses": []string{msg.To},
		},
		"Content": map[string]any{
			"Simple": map[string]any{
				"Subject": map[string]any{
					"Data":    msg.Subject,
					"Charset": "UTF-8",
				},
				"Body": map[string]any{
					"Text": map[string]any{
						"Data":    msg.Plain,
						"Charset": "UTF-8",
					},
				},
			},
		},
	}
	if msg.HTML != "" {
		bodyMap := payload["Content"].(map[string]any)["Simple"].(map[string]any)["Body"].(map[string]any)
		bodyMap["Html"] = map[string]any{
			"Data":    msg.HTML,
			"Charset": "UTF-8",
		}
	}
	if s.replyTo != "" {
		payload["ReplyToAddresses"] = []string{s.replyTo}
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := s.endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://email.%s.amazonaws.com", s.region)
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	targetURL := fmt.Sprintf("%s/v2/email/outbound-emails", strings.TrimRight(endpoint, "/"))

	req, err := http.NewRequest("POST", targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write(bodyBytes)
	payloadHash := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", u.Host)
	req.Header.Set("X-Amz-Date", amzDate)

	canonicalHeaders := fmt.Sprintf("content-type:application/json\nhost:%s\nx-amz-date:%s\n", u.Host, amzDate)
	signedHeaders := "content-type;host;x-amz-date"
	canonicalRequest := fmt.Sprintf("POST\n/v2/email/outbound-emails\n\n%s\n%s\n%s", canonicalHeaders, signedHeaders, payloadHash)

	h = sha256.New()
	h.Write([]byte(canonicalRequest))
	canonicalReqHash := hex.EncodeToString(h.Sum(nil))

	credentialScope := fmt.Sprintf("%s/%s/ses/aws4_request", dateStamp, s.region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", amzDate, credentialScope, canonicalReqHash)

	signingKey := getSignatureKey(s.secretAccessKey, dateStamp, s.region, "ses")
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKeyID, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ses error (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignatureKey(key, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+key), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(regionName))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}
