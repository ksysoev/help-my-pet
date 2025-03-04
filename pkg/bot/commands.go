package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ksysoev/help-my-pet/pkg/core/message"
	"github.com/ksysoev/help-my-pet/pkg/i18n"
)

func (s *ServiceImpl) HandleCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	switch msg.Command() {
	case "start":
		return s.handleStart(ctx, msg)
	case "terms":
		return s.handleTermsAndConditions(ctx, msg)
	case "editprofile":
		req, err := message.NewUserMessage(
			fmt.Sprintf("%d", msg.From.ID),
			fmt.Sprintf("%d", msg.Chat.ID),
			msg.Text,
		)
		if err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to create user update: %w", err)
		}

		resp, err := s.AISvc.ProcessEditProfile(ctx, req)
		if err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to process edit profile request: %w", err)
		}

		return tgbotapi.NewMessage(msg.Chat.ID, resp.Message), nil
	case "cancel":
		if err := s.AISvc.CancelQuestionnaire(ctx, fmt.Sprintf("%d", msg.Chat.ID)); err != nil {
			return tgbotapi.MessageConfig{}, fmt.Errorf("failed to reset conversation: %w", err)
		}

		resp := tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf("Questionary is cancelled"))

		resp.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
			RemoveKeyboard: true,
			Selective:      false,
		}

		return resp, nil
	case "help":
		return handleHelp(ctx, msg)
	default:
		return tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf("Unknown command")), nil
	}
}

func (s *ServiceImpl) handleStart(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	startMsg := i18n.GetLocale(ctx).Sprintf(`Welcome to Help My Pet Bot! 🐾

I'm your personal pet care assistant, ready to help you take better care of your furry friend. I can analyze photos of your pet and assist you with:

- Pet health and behavior questions - feel free to share photos of any concerning symptoms or behaviors
- Diet and nutrition advice, including photos of food or treats you're considering
- Training tips and techniques - you can send pictures of specific situations you need help with
- General pet care guidance, with photo analysis of your pet's environment or supplies

Simply type your question or send photos along with your concerns about your pet, and I'll provide helpful, informative answers based on reliable veterinary knowledge. You can send multiple photos if needed to better show the situation.

Remember, while I can offer guidance and analyze photos, for serious medical conditions, always consult with a veterinarian.

To get started, just ask me any question about your pet - with or without photos!`)

	return tgbotapi.NewMessage(msg.Chat.ID, startMsg), nil
}

func (s *ServiceImpl) handleTermsAndConditions(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	m := tgbotapi.NewMessage(msg.Chat.ID, i18n.GetLocale(ctx).Sprintf(`<b>Terms and Conditions</b>
<i>Last updated: 30.01.2025</i>

Thank you for using our veterinary advice chatbot (“the Service”). By accessing or using this Service, you agree to be bound by the following terms and conditions (“Terms”). If you do not agree to these Terms, please discontinue use immediately.

<b>1. Nature of the Service</b>
1.1 The Service provides general information, guidance, and suggestions for pet care, including (but not limited to) diet, behavior, and training.
1.2 The Service is not a substitute for professional veterinary diagnosis, treatment, or care. Always seek the advice of a licensed veterinarian for any questions regarding your pet’s health.

<b>2. No Veterinary-Client-Patient Relationship</b>
2.1 Using the Service or engaging with our AI assistant does not create a veterinarian-client-patient relationship.
2.2 Any advice or guidance provided by the Service is based on limited information and should only be considered general information.

<b>3. Limitation of Liability</b>
3.1 You acknowledge and agree that use of the Service is at your own risk.
3.2 Under no circumstances shall the owners, developers, or licensors of the Service be liable for any direct, indirect, incidental, special, or consequential damages arising out of or in connection with your access to or use of the Service.
3.3 You understand that decisions regarding your pet’s care and any resulting outcomes are your sole responsibility. If you have any doubt about the well-being of your pet or its health, you should immediately consult a licensed veterinarian.

<b>4. No Warranty</b>
4.1 The Service is provided on an “as is” and “as available” basis without warranties of any kind, whether express or implied.
4.2 We do not warrant that the Service will be uninterrupted, error-free, secure, or free from viruses.

<b>5. User Responsibilities</b>
5.1 You are responsible for providing accurate and complete information about your pet when seeking advice.
5.2 You must ensure that all questions, descriptions, and data you provide do not violate any third-party rights or local laws.

<b>6. International Use</b>
6.1 The Service is intended for global use. You are responsible for compliance with all applicable local laws and regulations in your jurisdiction.
6.2 We do not guarantee that the Service or any of its content is appropriate or permissible in any specific country or region.

<b>7. Modifications</b>
7.1 We reserve the right to modify or replace these Terms at any time.
7.2 If we make material changes, we will post the updated Terms and indicate the date of the latest revision at the top of this document.

<b>8. Governing Law and Dispute Resolution</b>
8.1 These Terms shall be governed by and construed in accordance with the laws applicable in the jurisdiction of the Service provider’s principal place of business, without regard to conflict-of-law principles.
8.2 Any dispute arising from or relating to these Terms shall be resolved through amicable negotiation and, if necessary, by binding arbitration or litigation in the applicable courts.

<b>9. Acceptance of Terms</b>
9.1 By continuing to access or use the Service, you acknowledge that you have read, understood, and agree to be bound by these Terms.
9.2 If you do not agree, you must cease using the Service immediately.

If you have any questions or concerns regarding these Terms, or if you need further clarification, please contact at <i>k.sysoev@me.com</i>.`))
	m.ParseMode = "HTML"

	return m, nil
}

func handleHelp(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	helpMsg := i18n.GetLocale(ctx).Sprintf(`<b>Help My Pet Bot Commands</b>:
/start - Start the conversation with the bot
/terms - View the Terms and Conditions of the service
/editprofile - Update your pet's profile information, such as name, age, breed, etc. This information helps the bot provide more accurate advice.
/cancel - Cancel the current questionnaire, if any is in progress (e.g., when you want to start over or change your question)
/help - View this help update`)

	tgMsg := tgbotapi.NewMessage(msg.Chat.ID, helpMsg)
	tgMsg.ParseMode = "HTML"

	return tgMsg, nil
}
