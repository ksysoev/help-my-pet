package bot

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/bot/media"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

const (
	maxAllowedPhotos = 2
	maxFileSize      = 1024 * 1024
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

	if mediaGroup.Text == "" {
		return tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf("Please, provide your question in text format along with photo(s)")), nil
	}

	if len(mediaGroup.PhotoIDs) == 0 {
		// should be impossible to reach this point in real life
		return tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf("Please, provide at least one image")), nil
	}

	if len(mediaGroup.PhotoIDs) > maxAllowedPhotos {
		return tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf("Please, provide no more than %d image(s)", maxAllowedPhotos)), nil
	}

	photoData := make([]*message.Image, 0, len(mediaGroup.PhotoIDs))

	for _, photoID := range mediaGroup.PhotoIDs {
		data, err := s.downloadPhoto(ctx, photoID)
		if err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to download photo: %w", err)
		}

		slog.Debug("Media group ID", slog.String("text", mediaGroup.Text), slog.Any("photo", *data))

		photoData = append(photoData, data)
	}

	usrMsg, err := message.NewUserMessage(
		fmt.Sprintf("%d", msg.From.ID),
		fmt.Sprintf("%d", msg.Chat.ID),
		mediaGroup.Text,
	)

	usrMsg.Images = photoData

	if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("failed to create user message: %w", err)
	}

	response, err := s.AISvc.ProcessMessage(ctx, usrMsg)

	if err != nil {
		return tgbotapi.MessageConfig{}, fmt.Errorf("failed to process user message: %w", err)
	}

	// Create msg with buttons if available
	resp := tgbotapi.NewMessage(msg.Chat.ID, response.Message)

	// Handle keyboard markup based on answers
	if len(response.Answers) > 0 {
		keyboard := make([][]tgbotapi.KeyboardButton, len(response.Answers))
		for i, answer := range response.Answers {
			keyboard[i] = []tgbotapi.KeyboardButton{
				{Text: answer},
			}
		}
		resp.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:        keyboard,
			OneTimeKeyboard: true,
			ResizeKeyboard:  true,
		}
	} else {
		resp.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
			Selective:      false,
		}
	}

	return resp, nil
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
		if photo.FileSize > maxFileSize {
			continue
		}

		if photo.FileSize > bestPhoto.FileSize {
			bestPhoto = photo
		}
	}

	return bestPhoto.FileID
}

// downloadPhoto retrieves an image file from the bot API and downloads it using its file ID.
// It fetches the file metadata, constructs the download URL, and performs the HTTP request to obtain the photo data.
// ctx is the context for managing request duration and cancellations.
// fileID is the unique identifier for the photo file to download.
// Returns a pointer to an image containing the MIME type and base64-encoded data, or an error if any step fails.
func (s *ServiceImpl) downloadPhoto(_ context.Context, fileID string) (*message.Image, error) {
	file, err := s.Bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, err
	}

	url := file.Link(s.token)

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to download photo: %w", err)
	}

	if resp.Body == nil {
		return nil, fmt.Errorf("failed to download photo: response body is nil")
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download photo: unexpected status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to download photo: %w", err)
	}

	base64data := base64.StdEncoding.EncodeToString(data)

	img := &message.Image{
		MIME: "image/jpeg",
		Data: base64data,
	}

	return img, nil
}
