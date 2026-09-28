package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("не найдено")

type ObjectClient struct {
	baseURL    string
	httpClient *http.Client

	mu       sync.Mutex
	qrToID   map[string]uint
	qrLoaded bool
}

func NewObjectClient(baseURL string, httpClient *http.Client) *ObjectClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &ObjectClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		qrToID:     map[string]uint{},
	}
}

func (c *ObjectClient) GetObjectByID(ctx context.Context, id uint) (*ObjectResponse, error) {
	var object ObjectResponse
	if err := c.getJSON(ctx, fmt.Sprintf("%s/api/max/v1/objects/%d", c.baseURL, id), &object); err != nil {
		return nil, err
	}
	return &object, nil
}

var addressSplit = regexp.MustCompile(`(?i)^(.+?)\s+(\d+[a-zа-яё]?(?:/\d+[a-zа-яё]?)?(?:\s*к\s*\d+[a-zа-яё]?)?)$`)

func (c *ObjectClient) Search(ctx context.Context, query string) ([]ObjectResponse, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, 0, ErrNotFound
	}
	parts := addressSplit.FindStringSubmatch(strings.Join(strings.Fields(strings.ReplaceAll(query, ",", " ")), " "))
	if parts == nil {
		return c.searchQuery(ctx, query, 8, 0)
	}
	list, total, err := c.searchQuery(ctx, parts[1], 200, 0)
	if err != nil {
		return nil, 0, err
	}
	if len(list) < total {
		rest, _, err := c.searchQuery(ctx, parts[1], min(300, total-len(list)), len(list))
		if err != nil {
			return nil, 0, err
		}
		list = append(list, rest...)
	}
	house := compactHouse(parts[2])
	matched := make([]ObjectResponse, 0, 1)
	for _, obj := range list {
		if compactHouse(obj.HouseNumber) == house {
			matched = append(matched, obj)
		}
	}
	return matched, len(matched), nil
}

func compactHouse(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
}

func (c *ObjectClient) searchQuery(ctx context.Context, query string, limit, offset int) ([]ObjectResponse, int, error) {
	var page objectList
	endpoint := fmt.Sprintf("%s/api/max/v1/objects?limit=%d&offset=%d&q=%s", c.baseURL, limit, offset, url.QueryEscape(query))
	if err := c.getJSON(ctx, endpoint, &page); err != nil {
		return nil, 0, err
	}
	return page.Objects, page.Total, nil
}

func (c *ObjectClient) QRImage(ctx context.Context, id uint) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/objects/%d/qr", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "image/png")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK || len(body) == 0 {
		return nil, fmt.Errorf("qr %s", resp.Status)
	}
	return body, nil
}

func (c *ObjectClient) FindObject(ctx context.Context, raw string) (*ObjectResponse, error) {
	raw = strings.TrimSpace(raw)
	if rest, ok := strings.CutPrefix(strings.ToLower(raw), "arcticpass:object:"); ok {
		raw = strings.TrimSpace(rest)
	}
	if id, err := strconv.ParseUint(raw, 10, 64); err == nil && id > 0 {
		return c.GetObjectByID(ctx, uint(id))
	}
	if !looksLikeQR(raw) {
		return nil, ErrNotFound
	}
	return c.FindByQR(ctx, raw)
}

func looksLikeQR(raw string) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if len(raw) != 36 {
		return false
	}
	for i, ch := range raw {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
				return false
			}
		}
	}
	return true
}

func (c *ObjectClient) FindByQR(ctx context.Context, code string) (*ObjectResponse, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return nil, ErrNotFound
	}

	id, ok, err := c.qrID(ctx, code, false)
	if err != nil {
		return nil, err
	}
	if !ok {
		id, ok, err = c.qrID(ctx, code, true)
		if err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, ErrNotFound
	}
	return c.GetObjectByID(ctx, id)
}

func (c *ObjectClient) qrID(ctx context.Context, code string, refresh bool) (uint, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if refresh || !c.qrLoaded {
		if err := c.loadQRIndex(ctx); err != nil {
			return 0, false, err
		}
	}
	id, ok := c.qrToID[code]
	return id, ok, nil
}

type objectList struct {
	Objects []ObjectResponse `json:"objects"`
	Total   int              `json:"total"`
}

func (c *ObjectClient) loadQRIndex(ctx context.Context) error {
	index := map[string]uint{}
	offset := 0
	for {
		var page objectList
		url := fmt.Sprintf("%s/api/max/v1/objects?limit=5000&offset=%d", c.baseURL, offset)
		if err := c.getJSON(ctx, url, &page); err != nil {
			return err
		}
		for _, obj := range page.Objects {
			code := strings.ToLower(strings.TrimSpace(obj.QRCode))
			if code != "" {
				index[code] = obj.ID
			}
		}
		offset += len(page.Objects)
		if len(page.Objects) == 0 || offset >= page.Total {
			break
		}
	}
	c.qrToID = index
	c.qrLoaded = true
	return nil
}

func (c *ObjectClient) GetApartment(ctx context.Context, objectID uint) (*ApartmentResponse, error) {
	var apt ApartmentResponse
	err := c.getJSON(ctx, fmt.Sprintf("%s/api/max/v1/objects/%d/apartment", c.baseURL, objectID), &apt)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &apt, nil
}

func (c *ObjectClient) getJSON(ctx context.Context, url string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr ErrorResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("%s", apiErr.Error)
		}
		return fmt.Errorf("ответ API %s", resp.Status)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return err
	}
	return nil
}
