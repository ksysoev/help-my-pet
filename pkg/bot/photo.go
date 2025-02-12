package bot

import (
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/media"
)

func (s *ServiceImpl) handlePhoto(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	mediaGroup, err := s.photoReducer(ctx, msg)
	if err != nil {
		return tgbotapi.MessageConfig{}, nil
	}

	slog.Debug("Received photo", slog.Any("media_group_id", mediaGroup))

	return tgbotapi.NewMessage(msg.Chat.ID, "Photo received"), nil
}

func (s *ServiceImpl) photoReducer(ctx context.Context, msg *tgbotapi.Message) (*media.Group, error) {
	photoID := getBestPhotoID(msg.Photo)

	s.collector.AddMediaGroup(msg.MediaGroupID, msg.Caption, photoID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(250 * time.Millisecond):
		return s.collector.FinishMediaGroup(msg.MediaGroupID), nil
	}
}

func getBestPhotoID(photos []tgbotapi.PhotoSize) string {
	if len(photos) == 0 {
		return ""
	}

	bestPhoto := photos[0]
	for _, photo := range photos {
		if photo.FileSize > bestPhoto.FileSize {
			bestPhoto = photo
		}
	}

	return bestPhoto.FileID
}
