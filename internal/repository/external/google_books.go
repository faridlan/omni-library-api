package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/faridlan/omni-library-api/internal/domain"
)

type googleBooksResponse struct {
	Items []struct {
		VolumeInfo struct {
			Title         string   `json:"title"`
			Authors       []string `json:"authors"`
			PublishedDate string   `json:"publishedDate"`
			Description   string   `json:"description"`
			PageCount     int      `json:"pageCount"`
			ImageLinks    struct {
				Thumbnail string `json:"thumbnail"`
			} `json:"imageLinks"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

type googleBooksFetcher struct {
	httpClient *http.Client
	apiKey     string
}

func NewGoogleBooksFetcher(apiKey string) domain.BookMetadataFetcher {
	return &googleBooksFetcher{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
	}
}

func (f *googleBooksFetcher) FetchByISBN(ctx context.Context, isbn string) (*domain.Book, error) {
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=isbn:%s&key=%s", isbn, f.apiKey)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("google books api limit exceeded (429 Too Many Requests)")
		}
		return nil, fmt.Errorf("google books api error: %s", resp.Status)
	}

	var apiResp googleBooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Items) == 0 {
		return nil, nil
	}

	item := apiResp.Items[0].VolumeInfo

	pubDate, _ := time.Parse("2006-01-02", item.PublishedDate)
	if item.PublishedDate != "" && pubDate.IsZero() {
		pubDate, _ = time.Parse("2006", item.PublishedDate)
	}

	return &domain.Book{
		ISBN:          isbn,
		Title:         item.Title,
		Authors:       item.Authors,
		PublishedDate: pubDate,
		Description:   item.Description,
		PageCount:     item.PageCount,
		CoverURL:      item.ImageLinks.Thumbnail,
	}, nil
}
