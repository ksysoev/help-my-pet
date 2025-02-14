package bot

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
)

func TestDownloadPhoto(t *testing.T) {
	tests := []struct {
		name         string
		fileID       string
		mockFileErr  error
		responseBody string
		responseErr  error
		statusCode   int
		expected     *message.Image
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "successful download",
			fileID:       "photo123",
			statusCode:   http.StatusOK,
			responseBody: "image data",
			expected: &message.Image{
				MIME: "image/jpeg",
				Data: base64.StdEncoding.EncodeToString([]byte("image data")),
			},
		},
		{
			name:        "file retrieval error",
			fileID:      "photo123",
			mockFileErr: errors.New("failed to retrieve file metadata"),
			wantErr:     true,
			errMsg:      "failed to retrieve file metadata",
		},
		{
			name:        "HTTP error during download",
			fileID:      "photo123",
			statusCode:  http.StatusInternalServerError,
			responseErr: errors.New("failed to make request"),
			wantErr:     true,
			errMsg:      "failed to download photo: failed to make request",
		},
		{
			name:       "unexpected status code",
			fileID:     "photo123",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
			errMsg:     "failed to download photo: unexpected status code 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockBot := NewMockBotAPI(t)

			mockHTTP := NewMockhttpClient(t)

			svc := &ServiceImpl{
				Bot:        mockBot,
				token:      "dummyToken",
				httpClient: mockHTTP,
			}

			mockFile := tgbotapi.File{
				FileID:   tt.fileID,
				FilePath: "path/to/file",
			}

			mockBot.EXPECT().GetFile(tgbotapi.FileConfig{FileID: tt.fileID}).Return(mockFile, tt.mockFileErr)

			if tt.statusCode != 0 {
				mockHTTP.EXPECT().Get(mock.Anything).Return(&http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				}, tt.responseErr)
			}

			// Act
			result, err := svc.downloadPhoto(context.Background(), tt.fileID)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

		})
	}
}

func TestGetBestPhotoID(t *testing.T) {
	tests := []struct {
		name     string
		photos   []tgbotapi.PhotoSize
		expected string
	}{
		{
			name:     "no photos",
			photos:   []tgbotapi.PhotoSize{},
			expected: "",
		},
		{
			name: "one photo",
			photos: []tgbotapi.PhotoSize{
				{FileID: "photo1", FileSize: 500},
			},
			expected: "photo1",
		},
		{
			name: "multiple photos with valid sizes",
			photos: []tgbotapi.PhotoSize{
				{FileID: "photo1", FileSize: 200},
				{FileID: "photo2", FileSize: 300},
				{FileID: "photo3", FileSize: 100},
			},
			expected: "photo2",
		},
		{
			name: "photos exceeding max file size",
			photos: []tgbotapi.PhotoSize{
				{FileID: "photo1", FileSize: 200},
				{FileID: "photo2", FileSize: 300},
				{FileID: "photo3", FileSize: 70000000}, // Exceeds maxFileSize
			},
			expected: "photo2",
		},
		{
			name: "all photos exceed max file size, return first",
			photos: []tgbotapi.PhotoSize{
				{FileID: "photo1", FileSize: 70000000},
				{FileID: "photo2", FileSize: 80000000},
			},
			expected: "photo1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBestPhotoID(tt.photos)
			assert.Equal(t, tt.expected, result)
		})
	}
}
