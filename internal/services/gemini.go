package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kamrul1157024/byoai-gemini/internal/apperror"
	"github.com/kamrul1157024/byoai-gemini/internal/config"
)

var INITIAL_PROMPTS_FOR_TEXT_GENERATION = map[string]string{
	"title":       "Suggest a title about ${topic}",
	"caption":     "Generate a caption about ${topic}",
	"description": "Write a description about ${topic}",
	"idea":        "Write creative idea related keywords on the ${topic}",
	"intro":       "Write a introduction paragraph on the following topic: ${topic}",
	"outline":     "Suggest an outline where each bullet point is separated by new line about the topic: ${topic}",
	"post":        "Write a post on the topic : ${topic}",
}

var PROMPTS_FOR_CONTENT_EDIT = map[string]string{
	"shorten":                           "Rewrite the following content but keep it short, simple",
	"elaborate":                         "Rewrite the following content to make it descriptive and easy to understand",
	"refine":                            "Rephrase but keeping the meaning intact of the following",
	"fixSpellingAndGrammaticalMistakes": "Fix spelling and grammatical mistakes of the following content",
	"informalTone":                      "Rephrase the following content in informal tone",
	"formalTone":                        "Rephrase the following content in formal and official tone",
}

type TextGenerationParams struct {
	WordLimit   int32   `json:"wordLimit"`
	Prompt      string  `json:"prompt"`
	Sentiment   *string `json:"sentiment"`
	Tone        *string `json:"tone"`
	GenerateFor *string `json:"generateFor"`
	UseCase     string  `json:"useCase"`
	Description *string `json:"description,omitEmpty"`
}

type TextCorrectionParams struct {
	Input          string `json:"input"`
	FineTuneOption string `json:"fineTuneOption"`
}

type ConversationItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatParams struct {
	Conversation []ConversationItem `json:"conversation"`
}

type Part struct {
	Text string `json:"text"`
}

type ChatContent struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type GeminiRequestBody struct {
	Contents []Content `json:"contents"`
}

type GeminiChatRequestBody struct {
	Contents []ChatContent `json:"contents"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type GeminiStreamResponse struct {
	Candidates []Candidate `json:"candidates"`
}

func getTextFromGeminiResponse(geminiStreamResponse GeminiStreamResponse) string {
	texts := []string{}

	for _, candidate := range geminiStreamResponse.Candidates {
		for _, part := range candidate.Content.Parts {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, " ")
}

func transfromChunk(chunkChannel <-chan []byte) <-chan string {
	textChannel := make(chan string)
	go func() {
		defer close(textChannel)
		for chunk := range chunkChannel {

			line := string(chunk)
			if !strings.Contains(line, "data:") {
				continue
			}
			jsonStr := line[6:]
			geminiStreamResponse := GeminiStreamResponse{}
			err := json.Unmarshal([]byte(jsonStr), &geminiStreamResponse)
			apperror.CheckAndLog(err, nil)
			bufferText := getTextFromGeminiResponse(geminiStreamResponse)
			fmt.Println(bufferText)
			textChannel <- bufferText
		}
	}()
	return textChannel
}

func streamToChannel(reader *bufio.Reader) <-chan []byte {
	chunkChannel := make(chan []byte)
	go func() {
		defer close(chunkChannel)
		for {
			lineBytes, err := reader.ReadBytes('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				apperror.CheckAndLog(err, nil)
				break
			}

			chunkChannel <- lineBytes
		}
	}()
	return chunkChannel
}

func getInitialPromptWithTopic(
	useCase string,
	topic string,
) string {
	return strings.Replace(INITIAL_PROMPTS_FOR_TEXT_GENERATION[useCase], "${topic}", topic, -1)
}

func getFormattedPromptForGenerativeAI(textGenerationRequestBody *TextGenerationParams) string {
	prompt := []string{}
	prompt = append(prompt, getInitialPromptWithTopic(textGenerationRequestBody.UseCase, textGenerationRequestBody.Prompt))
	prompt = append(prompt, ` in ${wordLimit} words`)
	if textGenerationRequestBody.GenerateFor != nil {
		prompt = append(prompt, ` for ${generateFor}`)
	}
	if textGenerationRequestBody.Sentiment != nil {
		prompt = append(prompt, ` in ${sentiment} sentiment`)
	}
	if textGenerationRequestBody.Tone != nil {
		prompt = append(prompt, ` and tone should be ${tone}`)
	}
	if textGenerationRequestBody.Description != nil {
		prompt = append(prompt, `. Take ideas from this description : ${description}`)
	}
	finalPrompt := strings.Join(prompt, " ")
	// logger.info(
	// 	`Generated prompt for ai text generation, prompt : ${finalPrompt}`,
	// )
	return finalPrompt
}

func getFormattedPromptForCorrectiveAI(correctiveAIRequestBody *TextCorrectionParams) string {
	fmt.Println(correctiveAIRequestBody)
	prompt := fmt.Sprintf(`
  You are helpful writting assistant,
  if you can not find proper response just send back the content.
  Do not add addtional information text with the response, only send the text that sent for correction
  %s
  '''
  %s
  '''`,
		PROMPTS_FOR_CONTENT_EDIT[correctiveAIRequestBody.FineTuneOption],
		correctiveAIRequestBody.Input,
	)
	fmt.Println(prompt)
	return prompt
}

func getGeminiPayloadForTextGeneration(textGenerationRequestBody *TextGenerationParams) *GeminiRequestBody {
	prompt := getFormattedPromptForGenerativeAI(textGenerationRequestBody)
	return &GeminiRequestBody{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: prompt,
					},
				},
			},
		},
	}
}

func getGeminiRule(chatParamsRule string) string {
	if chatParamsRule == "assistant" {
		return "model"
	}
	return "user"
}

func getGeminiPayloadForChat(chatParams *ChatParams) *GeminiChatRequestBody {
	contents := []ChatContent{}
	contents = append(contents, ChatContent{
		Role: "user",
		Parts: []Part{
			{
				Text: "You are a helpful writing assistant",
			},
		},
	})
	for _, conversationItem := range chatParams.Conversation {
		if conversationItem.Role == "system" {
			continue
		}
		contents = append(contents, ChatContent{
			Role: getGeminiRule(conversationItem.Role),
			Parts: []Part{
				{
					Text: conversationItem.Content,
				},
			},
		})
	}
	return &GeminiChatRequestBody{
		Contents: contents,
	}
}

func getGeminiPayloadForTextCorrection(textCorrectionParams *TextCorrectionParams) *GeminiChatRequestBody {
	prompt := getFormattedPromptForCorrectiveAI(textCorrectionParams)
	return &GeminiChatRequestBody{
		Contents: []ChatContent{
			{
				Role: "user",
				Parts: []Part{
					{
						Text: prompt,
					},
				},
			},
		},
	}
}

func callApi[V any](payload *V) <-chan string {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:streamGenerateContent?alt=sse&key=%s",
		config.AppConfig.GEMINI_API_KEY,
	)
	reqJson, err := json.Marshal(payload)
	fmt.Println(string(reqJson))
	apperror.CheckAndLog(err, nil)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(reqJson)))
	apperror.CheckAndLog(err, nil)
	reader := bufio.NewReader(resp.Body)
	bytesChan := streamToChannel(reader)
	textChan := transfromChunk(bytesChan)
	return textChan
}

func GetResponseChanForGenerativeAI(textGenerationRequestBody *TextGenerationParams) <-chan string {
	payload := getGeminiPayloadForTextGeneration(textGenerationRequestBody)
	return callApi(payload)
}

func GetResponseChanForChat(chatRequestBody *ChatParams) <-chan string {
	payload := getGeminiPayloadForChat(chatRequestBody)
	return callApi(payload)
}

func GetResponseChanForCorrectiveAI(correctiveAIRequestBody *TextCorrectionParams) <-chan string {
	payload := getGeminiPayloadForTextCorrection(correctiveAIRequestBody)
	return callApi(payload)
}
