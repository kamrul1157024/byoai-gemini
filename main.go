package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TextGenerationRequestBody struct {
	WordLimit   int32   `json:"wordLimit"`
	Prompt      string  `json:"prompt"`
	Sentiment   *string `json:"sentiment"`
	Tone        *string `json:"tone"`
	GenerateFor *string `json:"generateFor"`
	UseCase     string  `json:"useCase"`
	Description *string `json:"description,omitEmpty"`
}

type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type GeminiRequestBody struct {
	Contents []Content `json:"contents"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type GeminiStreamResponse struct {
	Candidates []Candidate `json:"candidates"`
}

var INITIAL_PROMPTS_FOR_TEXT_GENERATION = map[string]string{
	"title":       "Suggest a title about ${topic}",
	"caption":     "Generate a caption about ${topic}",
	"description": "Write a description about ${topic}",
	"idea":        "Write creative idea related keywords on the ${topic}",
	"intro":       "Write a introduction paragraph on the following topic: ${topic}",
	"outline":     "Suggest an outline where each bullet point is separated by new line about the topic: ${topic}",
	"post":        "Write a post on the topic : ${topic}",
}

func getInitialPromptWithTopic(
	useCase string,
	topic string,
) string {
	return strings.Replace(INITIAL_PROMPTS_FOR_TEXT_GENERATION[useCase], "${topic}", topic, -1)
}

func getFormattedPromptForGenerativeAI(textGenerationRequestBody TextGenerationRequestBody) string {
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

func getGeminiRequest(textGenerationRequestBody TextGenerationRequestBody) *GeminiRequestBody {
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

func getTextFromGeminiResponse(geminiStreamResponse GeminiStreamResponse) string {
	texts := []string{}

	for _, candidate := range geminiStreamResponse.Candidates {
		for _, part := range candidate.Content.Parts {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, " ")
}

func generateTextUsingGemini(c *gin.Context) {
	textGenerationRequestBody := TextGenerationRequestBody{}
	err := c.BindJSON(&textGenerationRequestBody)
	if err == nil {
		fmt.Println(err)
	}
	reqBody := getGeminiRequest(textGenerationRequestBody)
	reqJson, _ := json.Marshal(reqBody)
	resp, err := http.Post(fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:streamGenerateContent?alt=sse&key=%s", "AIzaSyAb09y9YVf_sr1oMNz7CiJ3lueQwFRXNXI"), "application/json", bytes.NewBuffer([]byte(reqJson)))
	if err != nil {
		panic("Failed While Calling API")
	}

	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Transfer-Encoding", "chunked")

	c.Stream(func(w io.Writer) bool {
		for {
			lineBytes, err := reader.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			line := string(lineBytes)
			if !strings.Contains(line, "data:") {
				continue
			}
			jsonStr := line[6:]
			geminiStreamResponse := GeminiStreamResponse{}
			err = json.Unmarshal([]byte(jsonStr), &geminiStreamResponse)

			bufferText := getTextFromGeminiResponse(geminiStreamResponse)
			w.Write([]byte(bufferText))
			if err != nil {
				print(err)
				return true
			}
		}

		return false
	})
}

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func main() {
	router := gin.Default()
	router.POST("/generative/text", generateTextUsingGemini)
	router.GET("/_status", getStatus)

	router.Run("localhost:8000")
}
