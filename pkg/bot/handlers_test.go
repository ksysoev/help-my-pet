package bot

import (
	"context"
	"fmt"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_handleMessage(t *testing.T) {
	tests := []struct {
		aiErr        error
		aiResponse   *core.PetAdviceResponse
		name         string
		message      string
		langCode     string
		userID       int64
		expectError  bool
		isStart      bool
		expectedText string
	}{
		{
			name:         "successful response with keyboard",
			message:      "What food is good for cats?",
			aiResponse:   core.NewPetAdviceResponse("Cats need a balanced diet...", []string{"Yes", "No"}),
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "Cats need a balanced diet...",
		},
		{
			name:         "successful response without keyboard",
			message:      "What food is good for cats?",
			aiResponse:   core.NewPetAdviceResponse("Cats need a balanced diet...", []string{}),
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "ru",
			expectedText: "Cats need a balanced diet...",
		},
		{
			name:         "empty message",
			message:      "",
			aiResponse:   core.NewPetAdviceResponse("", []string{}),
			aiErr:        nil,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "",
		},
		{
			name:         "ai error",
			message:      "What food is good for cats?",
			aiResponse:   core.NewPetAdviceResponse("", []string{}),
			aiErr:        fmt.Errorf("ai error"),
			expectError:  true,
			userID:       123,
			langCode:     "es",
			expectedText: "",
		},
		{
			name:         "rate limit error",
			message:      "What food is good for cats?",
			aiResponse:   core.NewPetAdviceResponse("", []string{}),
			aiErr:        core.ErrRateLimit,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "You have reached the maximum number of requests per hour. Please try again later.",
		},
		{
			name:         "global limit error",
			message:      "What food is good for cats?",
			aiResponse:   core.NewPetAdviceResponse("", []string{}),
			aiErr:        core.ErrGlobalLimit,
			expectError:  false,
			userID:       123,
			langCode:     "en",
			expectedText: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
		},
		{
			name:         "rate limit error - ru",
			message:      "What food is good for cats?",
			aiResponse:   core.NewPetAdviceResponse("", []string{}),
			aiErr:        core.ErrRateLimit,
			expectError:  false,
			userID:       123,
			langCode:     "ru",
			expectedText: "Вы достигли максимального количества запросов в час. Пожалуйста, попробуйте позже.",
		},
		{
			name:         "global limit error - ru",
			message:      "What food is good for cats?",
			aiResponse:   core.NewPetAdviceResponse("", []string{}),
			aiErr:        core.ErrGlobalLimit,
			expectError:  false,
			userID:       123,
			langCode:     "ru",
			expectedText: "Мы достигли дневного лимита запросов. Пожалуйста, возвращайтесь завтра, когда наш бюджет обновится.",
		},
		{
			name:         "start command",
			message:      "/start",
			aiResponse:   core.NewPetAdviceResponse("Welcome to Help My Pet Bot!", []string{}),
			aiErr:        nil,
			expectError:  false,
			isStart:      true,
			userID:       123,
			langCode:     "de",
			expectedText: "Willkommen bei Help My Pet Bot!",
		},
		{
			name:         "message without From field",
			message:      "What food is good for cats?",
			aiResponse:   nil,
			aiErr:        nil,
			expectError:  true,
			userID:       0,
			langCode:     "en",
			expectedText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAI := NewMockAIProvider(t)
			mockBot := NewMockBotAPI(t)

			messages := &i18n.Config{
				Languages: map[string]i18n.Messages{
					"en": {
						Error:       "Sorry, I encountered an error while processing your request. Please try again later.",
						Start:       "Welcome to Help My Pet Bot!",
						RateLimit:   "You have reached the maximum number of requests per hour. Please try again later.",
						GlobalLimit: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
					},
					"ru": {
						Error:       "Извините, произошла ошибка при обработке вашего запроса. Пожалуйста, попробуйте позже.",
						Start:       "Добро пожаловать в Help My Pet Bot!",
						RateLimit:   "Вы достигли максимального количества запросов в час. Пожалуйста, попробуйте позже.",
						GlobalLimit: "Мы достигли дневного лимита запросов. Пожалуйста, возвращайтесь завтра, когда наш бюджет обновится.",
					},
					"es": {
						Error:       "Lo siento, encontré un error al procesar tu solicitud. Por favor, inténtalo más tarde.",
						Start:       "¡Bienvenido a Help My Pet Bot!",
						RateLimit:   "Has alcanzado el número máximo de solicitudes por hora. Por favor, inténtalo más tarde.",
						GlobalLimit: "Hemos alcanzado nuestro límite diario de solicitudes. Por favor, vuelve mañana cuando nuestro presupuesto se haya renovado.",
					},
					"fr": {
						Error:       "Désolé, j'ai rencontré une erreur lors du traitement de votre demande. Veuillez réessayer plus tard.",
						Start:       "Bienvenue sur Help My Pet Bot !",
						RateLimit:   "Vous avez atteint le nombre maximum de demandes par heure. Veuillez réessayer plus tard.",
						GlobalLimit: "Nous avons atteint notre limite quotidienne de demandes. Veuillez revenir demain lorsque notre budget sera renouvelé.",
					},
					"de": {
						Error:       "Entschuldigung, bei der Verarbeitung Ihrer Anfrage ist ein Fehler aufgetreten. Bitte versuchen Sie es später erneut.",
						Start:       "Willkommen bei Help My Pet Bot!",
						RateLimit:   "Sie haben die maximale Anzahl an Anfragen pro Stunde erreicht. Bitte versuchen Sie es später erneut.",
						GlobalLimit: "Wir haben unser tägliches Anfragelimit erreicht. Bitte kommen Sie morgen wieder, wenn unser Budget erneuert wurde.",
					},
				},
			}

			svc := &ServiceImpl{
				Bot:      mockBot,
				AISvc:    mockAI,
				Messages: messages,
			}

			msg := &tgbotapi.Message{
				Text: tt.message,
				Chat: &tgbotapi.Chat{
					ID: 123,
				},
				MessageID: 456,
			}

			// Set From field only if userID is not 0
			if tt.userID != 0 {
				msg.From = &tgbotapi.User{
					ID:           tt.userID,
					LanguageCode: tt.langCode,
				}
			}

			if !tt.isStart && tt.message != "" && msg.From != nil {
				expectedRequest := &core.PetAdviceRequest{
					UserID:  "123",
					ChatID:  "123",
					Message: tt.message,
				}
				mockAI.EXPECT().
					GetPetAdvice(context.Background(), expectedRequest).
					Return(tt.aiResponse, tt.aiErr)
			}

			msgConfig, err := svc.handleMessage(context.Background(), msg)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.message == "" {
				assert.Equal(t, tgbotapi.MessageConfig{}, msgConfig)
				return
			}

			assert.Equal(t, tt.expectedText, msgConfig.Text)
			assert.Equal(t, int64(123), msgConfig.ChatID)

			if tt.aiResponse != nil && len(tt.aiResponse.Answers) > 0 {
				keyboard, ok := msgConfig.ReplyMarkup.(tgbotapi.ReplyKeyboardMarkup)
				assert.True(t, ok)
				assert.Equal(t, len(tt.aiResponse.Answers), len(keyboard.Keyboard))
				for i, row := range keyboard.Keyboard {
					assert.Equal(t, tt.aiResponse.Answers[i], row[0].Text)
				}
			}
		})
	}
}

func TestService_Run_SuccessfulMessageHandling(t *testing.T) {
	mockAI := NewMockAIProvider(t)
	mockBot := NewMockBotAPI(t)
	messages := &i18n.Config{
		Languages: map[string]i18n.Messages{
			"en": {
				Error:       "Sorry, I encountered an error while processing your request. Please try again later.",
				Start:       "Welcome to Help My Pet Bot!",
				RateLimit:   "You have reached the maximum number of requests per hour. Please try again later.",
				GlobalLimit: "We have reached our daily request limit. Please come back tomorrow when our budget is refreshed.",
			},
		},
	}

	svc := &ServiceImpl{
		Bot:      mockBot,
		AISvc:    mockAI,
		Messages: messages,
	}

	updates := make(chan tgbotapi.Update)
	mockBot.EXPECT().
		GetUpdatesChan(tgbotapi.UpdateConfig{Offset: 0, Timeout: 30}).
		Return(updates)

	// Expect typing action
	mockBot.EXPECT().
		Send(tgbotapi.NewChatAction(int64(123), tgbotapi.ChatTyping)).
		Return(tgbotapi.Message{}, nil)

	// Expect AI request
	mockAI.EXPECT().
		GetPetAdvice(mock.Anything, &core.PetAdviceRequest{
			UserID:  "123",
			ChatID:  "123",
			Message: "test message",
		}).
		Return(core.NewPetAdviceResponse("test response", []string{}), nil)

	// Expect message send
	mockBot.EXPECT().
		Send(mock.MatchedBy(func(c tgbotapi.Chattable) bool {
			msg, ok := c.(tgbotapi.MessageConfig)
			return ok && msg.Text == "test response"
		})).
		Return(tgbotapi.Message{}, nil)

	mockBot.EXPECT().
		StopReceivingUpdates().
		Return()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)

	go func() {
		errCh <- svc.Run(ctx)
	}()

	updates <- tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "test message",
			Chat: &tgbotapi.Chat{
				ID: 123,
			},
			MessageID: 456,
			From: &tgbotapi.User{
				ID:           123,
				LanguageCode: "en",
			},
		},
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-errCh
	assert.NoError(t, err)
}
