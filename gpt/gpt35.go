package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qingconglaixueit/wechatbot/config"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const BASEURL = "https://api.openai.com/v1/chat/"

const (
	RoleUser      string = "user"
	RoleAssistant string = "assistant"
	RoleSystem    string = "system"
)

// ChatGPT35ResponseBody 请求体
type ChatGPT35ResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChoiceItem           `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChoiceItem struct {
	Message      Messages `json:"message"`
	Index        int      `json:"index"`
	Logprobs     int      `json:"logprobs"`
	FinishReason string   `json:"finish_reason"`
}

// ChatGPTRequestBody 响应体
type ChatGPTRequestBody struct {
	Model            string      `json:"model"`
	Messages         []*Messages `json:"messages"`
	TopP             int         `json:"top_p"`
	FrequencyPenalty int         `json:"frequency_penalty"`
	PresencePenalty  int         `json:"presence_penalty"`
}

type Messages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionsForGpt35 curl https://api.openai.com/v1/chat/completions \
// -H 'Content-Type: application/json' \
// -H 'Authorization: Bearer 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' \
// -d '{
// "model": "gpt-3.5-turbo",
// "messages": [{"role": "user", "content": "Hello!"}]
// }'
// @link https://platform.openai.com/docs/api-reference/chat/create
func CompletionsForGpt35(msg string) (string, error) {
	cfg := config.LoadConfig()
	requestBody := ChatGPTRequestBody{
		Model: cfg.Model,
		Messages: []*Messages{
			{
				Role:    "user",
				Content: msg,
			},
		},
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	logger.Info(fmt.Sprintf("request gpt json string : %v", string(requestData)))
	req, err := http.NewRequest("POST", BASEURL+"completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := config.LoadConfig().ApiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return "", errors.New(fmt.Sprintf("请求GTP出错了，gpt api status code not equals 200,code is %d ,details:  %v ", response.StatusCode, string(body)))
	}
	//var resp *ChatGPTResponseBody
	//err = json.NewDecoder(response.Body).Decode(&resp)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	logger.Info(fmt.Sprintf("response gpt json string : %v", string(body)))

	gptResponseBody := &ChatGPT35ResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Message.Content
	}
	logger.Info(fmt.Sprintf("gpt response text: %s ", reply))
	return reply, nil
}
