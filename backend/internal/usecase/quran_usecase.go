package usecase

import (
	"context"
	"fmt"

	"github.com/abdullahPrasetio/waphafiz/pkg/quranapi"
)

type QuranUseCase interface {
	GetSurahList(ctx context.Context) ([]quranapi.SurahMeta, error)
	GetSurahDetail(ctx context.Context, number int) (*quranapi.SurahDetail, error)
	GetAyahAudio(ctx context.Context, surah, ayah int) (*quranapi.AudioAyah, error)
}

type quranUseCase struct {
	client *quranapi.Client
}

func NewQuranUseCase(client *quranapi.Client) QuranUseCase {
	return &quranUseCase{client: client}
}

func (u *quranUseCase) GetSurahList(ctx context.Context) ([]quranapi.SurahMeta, error) {
	surahs, err := u.client.GetSurahList(ctx)
	if err != nil {
		return nil, fmt.Errorf("get surah list: %w", err)
	}
	return surahs, nil
}

func (u *quranUseCase) GetSurahDetail(ctx context.Context, number int) (*quranapi.SurahDetail, error) {
	if number < 1 || number > 114 {
		return nil, ErrNotFound
	}
	detail, err := u.client.GetSurahDetail(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("get surah detail: %w", err)
	}
	return detail, nil
}

func (u *quranUseCase) GetAyahAudio(ctx context.Context, surah, ayah int) (*quranapi.AudioAyah, error) {
	if surah < 1 || surah > 114 || ayah < 1 {
		return nil, ErrNotFound
	}
	audio, err := u.client.GetAyahAudio(ctx, surah, ayah)
	if err != nil {
		return nil, fmt.Errorf("get ayah audio: %w", err)
	}
	return audio, nil
}
