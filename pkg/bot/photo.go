package bot

import (
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/media"
)

// handlePhoto processes a photo sent by the user and responds with a confirmation message.
// It reduces the photo group using the photoReducer function and logs the media group ID.
// Accepts ctx, the request context, and msg, the incoming Telegram message containing the photo.
// Returns a confirmation Telegram message response or an error if photo reduction fails.
func (s *ServiceImpl) handlePhoto(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	mediaGroup, err := s.photoReducer(ctx, msg)
	if err != nil {
		return tgbotapi.MessageConfig{}, nil
	}

	slog.Debug("Received photo", slog.Any("media_group_id", mediaGroup))

	return tgbotapi.NewMessage(msg.Chat.ID, "Photo received"), nil
}

// photoReducer processes a photo message, aggregates it within a media group, and finalizes the group after a delay.
// It selects the best photo from the provided message using specific criteria, adds it to a media group with a caption,
// and completes the group after waiting for additional photos or upon context cancellation.
// Accepts ctx, the request context for lifecycle management, and msg, the Telegram message containing the photo data.
// Returns a finalized media.Group containing aggregated photo IDs and captions, or an error if the context is cancelled or times out.
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

// getBestPhotoID selects the photo with the largest file size from a slice of PhotoSize structs.
// It returns an empty string if the input slice is empty.
// photos is the list of PhotoSize objects to evaluate.
// Returns the FileID of the largest photo or an empty string if no photos are provided.
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
