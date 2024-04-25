package apis

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kamrul1157024/byoai-gemini/apis/middlewares"
	"github.com/kamrul1157024/byoai-gemini/internal/apperror"
	"github.com/kamrul1157024/byoai-gemini/internal/services"
)

func StreamResponse(c *gin.Context, ch <-chan string) {
	c.Stream(func(w io.Writer) bool {
		output, ok := <-ch
		if !ok {
			return false
		}
		c.Writer.Write([]byte(output))
		return true
	})
}

func ParseBody[B any](c *gin.Context, body *B) *B {
	err := c.BindJSON(&body)
	apperror.CheckAndLog(err, "Failed to parse JSON")
	return body
}

func generateTextUsingGemini(c *gin.Context) {
	textGenerationRequestBody := ParseBody(c, &services.TextGenerationParams{})
	ch := services.GetResponseChanForGenerativeAI(textGenerationRequestBody)
	StreamResponse(c, ch)
}

func generateContextFulChatUsingGemini(c *gin.Context) {
	chatRequestBody := ParseBody(c, &services.ChatParams{})
	ch := services.GetResponseChanForChat(chatRequestBody)
	StreamResponse(c, ch)
}

func correctTextUsingGemini(c *gin.Context) {
	correctiveAIRequestBody := ParseBody(c, &services.TextCorrectionParams{})
	ch := services.GetResponseChanForCorrectiveAI(correctiveAIRequestBody)
	StreamResponse(c, ch)
}

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func AddRoutesForGeminiAI(r *gin.RouterGroup) {
	r.Use(middlewares.StreamingHeader())
	r.POST("/generative/text", generateTextUsingGemini)
	r.POST("/corrective/text", correctTextUsingGemini)
	r.POST("/chat", generateContextFulChatUsingGemini)
}
