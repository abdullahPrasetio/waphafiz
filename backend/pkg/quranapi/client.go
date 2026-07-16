package quranapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	baseURL    = "https://api.alquran.cloud/v1"
	cacheTTL   = 24 * time.Hour
	httpTimeout = 10 * time.Second
)

type Client struct {
	http  *http.Client
	redis *redis.Client
}

func New(redis *redis.Client) *Client {
	return &Client{
		http:  &http.Client{Timeout: httpTimeout},
		redis: redis,
	}
}

// SurahMeta is metadata for a surah from the list endpoint.
type SurahMeta struct {
	Number           int    `json:"number"`
	Name             string `json:"name"`
	EnglishName      string `json:"englishName"`
	EnglishNameTrans string `json:"englishNameTranslation"`
	NumberOfAyahs    int    `json:"numberOfAyahs"`
	RevelationType   string `json:"revelationType"`
}

// Ayah is a single verse from a surah.
type Ayah struct {
	Number     int    `json:"number"`
	Text       string `json:"text"`
	NumberInSurah int `json:"numberInSurah"`
	Juz        int    `json:"juz"`
	Page       int    `json:"page"`
}

// SurahDetail holds full surah info including all ayahs.
type SurahDetail struct {
	SurahMeta
	Ayahs []Ayah `json:"ayahs"`
}

// AudioAyah holds audio data for a single ayah.
type AudioAyah struct {
	Number    int    `json:"number"`
	Audio     string `json:"audio"`
	AudioFile string `json:"audioFile"`
}

func (c *Client) GetSurahList(ctx context.Context) ([]SurahMeta, error) {
	cacheKey := "quran:surah:list"
	if cached, err := c.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var surahs []SurahMeta
		if json.Unmarshal(cached, &surahs) == nil {
			return surahs, nil
		}
	}

	var resp struct {
		Data []SurahMeta `json:"data"`
	}
	if err := c.get(ctx, "/surah", &resp); err != nil {
		return nil, err
	}

	if b, err := json.Marshal(resp.Data); err == nil {
		c.redis.Set(ctx, cacheKey, b, cacheTTL) //nolint:errcheck
	}
	return resp.Data, nil
}

func (c *Client) GetSurahDetail(ctx context.Context, number int) (*SurahDetail, error) {
	cacheKey := fmt.Sprintf("quran:surah:%d", number)
	if cached, err := c.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var detail SurahDetail
		if json.Unmarshal(cached, &detail) == nil {
			return &detail, nil
		}
	}

	var resp struct {
		Data SurahDetail `json:"data"`
	}
	if err := c.get(ctx, fmt.Sprintf("/surah/%d", number), &resp); err != nil {
		return nil, err
	}

	if b, err := json.Marshal(resp.Data); err == nil {
		c.redis.Set(ctx, cacheKey, b, cacheTTL) //nolint:errcheck
	}
	return &resp.Data, nil
}

func (c *Client) GetAyahAudio(ctx context.Context, surah, ayah int) (*AudioAyah, error) {
	edition := "ar.alafasy"
	ref := fmt.Sprintf("%d:%d", surah, ayah)
	cacheKey := fmt.Sprintf("quran:ayah:%s:%s", ref, edition)

	if cached, err := c.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		var audio AudioAyah
		if json.Unmarshal(cached, &audio) == nil {
			return &audio, nil
		}
	}

	var resp struct {
		Data AudioAyah `json:"data"`
	}
	if err := c.get(ctx, fmt.Sprintf("/ayah/%s/%s", ref, edition), &resp); err != nil {
		return nil, err
	}

	if b, err := json.Marshal(resp.Data); err == nil {
		c.redis.Set(ctx, cacheKey, b, cacheTTL) //nolint:errcheck
	}
	return &resp.Data, nil
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	res, err := c.http.Do(req)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("quranapi request failed")
		return fmt.Errorf("quran api unavailable: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("quran api returned %d for %s", res.StatusCode, path)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
