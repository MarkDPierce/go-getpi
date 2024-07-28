package client

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"go-getpi/config"
	"go-getpi/utils"

	"github.com/PuerkitoBio/goquery"
)

type Client struct {
	Client  *http.Client
	Host    config.Host
	Token   string // Encrypted token
	Options Options
}

type Options struct {
	Whitelist         bool
	RegexWhitelist    bool
	Blacklist         bool
	RegexList         bool
	AdList            bool
	Client            bool
	Group             bool
	AuditLog          bool
	StaticDhcpLeases  bool
	LocalDnsRecords   bool
	LocalCnameRecords bool
	FlushTables       bool
}

func NewHost(baseUrl, password string, path ...string) *config.Host {
	p := utils.DeterminePath(path)
	baseUrl, p = utils.ExtractIncludedPath(baseUrl, p)
	fullURL := utils.CombineBaseURLAndPath(baseUrl, p)

	host := &config.Host{
		BaseURL:  utils.TrimTrailingSlash(baseUrl),
		Password: password,
		Path:     utils.TrimTrailingSlash(p),
		FullURL:  fullURL,
	}
	log.Printf("Full URL: %s", host.FullURL)
	return host
}

func NewClient(h config.Host) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %v", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !h.SslSecure},
	}

	client := &Client{
		Client: &http.Client{
			Jar:       jar,
			Transport: transport,
		},
		Host: h,
	}
	return client, nil
}

func (c *Client) Login(insecure bool) (string, error) {
	loginURL := fmt.Sprintf("%s/index.php?login", c.Host.FullURL)
	log.Printf("Login URL: %s", loginURL)
	data := url.Values{}
	data.Set("pw", c.Host.Password)
	data.Set("persistentlogin", "off")

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating request to login: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Read response body for more details
		log.Printf("Response body: %s", string(body))
		return "", fmt.Errorf("failed to log in: %s\n URL: %s\n LoginURL: %s\n ", resp.Status, c.Host.FullURL, loginURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	token, err := c.parseResponseForToken(string(body))
	if err != nil {
		return "", fmt.Errorf("error parsing response for token: %v", err)
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	encryptedToken, err := utils.Encrypt(token, encryptionKey)
	if err != nil {
		return "", fmt.Errorf("error encrypting token: %v", err)
	}

	c.Token = encryptedToken

	log.Printf("Token retrieved and encrypted: %s", encryptedToken)
	log.Printf("Cookies after login: %v", c.Client.Jar.Cookies(req.URL))

	return encryptedToken, nil
}

func (c *Client) parseResponseForToken(responseBody string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(responseBody))
	if err != nil {
		return "", err
	}

	token := doc.Find("div#token").Text()
	if token == "" {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}

func (c *Client) DownloadBackup() ([]byte, error) {
	backupURL := fmt.Sprintf("%s/scripts/pi-hole/php/teleporter.php", c.Host.FullURL)

	form := url.Values{}
	form.Set("pw", c.Host.Password)
	form.Set("persistentlogin", "off")
	decryptedToken, err := utils.Decrypt(c.Token, os.Getenv("ENCRYPTION_KEY"))
	if err != nil {
		return nil, fmt.Errorf("error decrypting token: %v", err)
	}
	form.Set("token", decryptedToken)

	req, err := http.NewRequest("POST", backupURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "application/gzip" {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download backup: %s\nStatus: %d\nResponse Body: %s", backupURL, resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	err = os.WriteFile("backup.gz", data, 0644)
	if err != nil {
		return nil, fmt.Errorf("error saving backup file: %v", err)
	}

	fmt.Println("✔️ Backup completed!")
	return data, nil
}

func (c *Client) UploadBackup(backup []byte) (bool, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	if err := w.WriteField("action", "in"); err != nil {
		return false, fmt.Errorf("failed to write action field: %v", err)
	}

	decryptedToken, err := utils.Decrypt(c.Token, os.Getenv("ENCRYPTION_KEY"))
	if err != nil {
		return false, fmt.Errorf("error decrypting token: %v", err)
	}
	if err := w.WriteField("token", decryptedToken); err != nil { // Add token as a form field
		return false, fmt.Errorf("failed to write token field: %v", err)
	}

	part, err := w.CreateFormFile("zip_file", "backup.tar.gz")
	if err != nil {
		return false, fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := part.Write(backup); err != nil {
		return false, fmt.Errorf("failed to write file to form: %v", err)
	}

	if err := w.Close(); err != nil {
		return false, fmt.Errorf("failed to close writer: %v", err)
	}

	uploadURL := fmt.Sprintf("%s/scripts/pi-hole/php/teleporter.php", c.Host.FullURL)
	req, err := http.NewRequest("POST", uploadURL, &b)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Log the cookies before making the request
	log.Printf("Cookies before upload request: %v", c.Client.Jar.Cookies(req.URL))

	resp, err := c.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}
	uploadText := string(body)

	if resp.StatusCode != http.StatusOK || !(strings.HasSuffix(uploadText, "OK") || strings.HasSuffix(uploadText, "Done importing")) {
		return false, fmt.Errorf("failed to upload backup: %s\nStatus: %d\nResponse Body: %s", uploadURL, resp.StatusCode, uploadText)
	}

	fmt.Println("✔️ Backup uploaded successfully!")
	fmt.Printf("Result:\n%s\n", uploadText)

	return true, nil
}

func (c *Client) UpdateGravity() (bool, error) {
	updateURL := fmt.Sprintf("%s/scripts/pi-hole/php/gravity.sh.php", c.Host.FullURL)

	req, err := http.NewRequest("GET", updateURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}
	decryptedToken, err := utils.Decrypt(c.Token, os.Getenv("ENCRYPTION_KEY"))
	if err != nil {
		return false, fmt.Errorf("error decrypting token: %v", err)
	}
	req.Header.Set("token", decryptedToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}
	updateText := string(body)

	updateText = strings.ReplaceAll(updateText, "\ndata:", "")
	updateText = strings.TrimSpace(updateText)

	if resp.StatusCode != http.StatusOK || !strings.HasSuffix(updateText, "Pi-hole blocking is enabled") {
		return false, fmt.Errorf(
			"failed updating gravity on \"%s\": Status: %d, Response Body: %s",
			c.Host.FullURL, resp.StatusCode, updateText,
		)
	}

	fmt.Println("✔️ Gravity updated successfully!")
	fmt.Printf("Result:\n%s\n", updateText)

	return true, nil
}
