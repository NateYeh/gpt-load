package proxy

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)
type openAIResponse struct {
	Usage usageInfo `json:"usage"`
}

func (ps *ProxyServer) handleStreamingResponse(c *gin.Context, resp *http.Response) *usageInfo {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		logrus.Error("Streaming unsupported by the writer, falling back to normal response")
		return ps.handleNormalResponse(c, resp)
	}

	var finalUsage *usageInfo
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := c.Writer.Write([]byte(line + "\n")); err != nil {
			logUpstreamError("writing stream to client", err)
			return nil
		}
		flusher.Flush()

		// Parse usage if present in the data line
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data != "[DONE]" && strings.Contains(data, "\"usage\"") {
				var uResp openAIResponse
				if err := json.Unmarshal([]byte(data), &uResp); err == nil && uResp.Usage.TotalTokens > 0 {
					finalUsage = &uResp.Usage
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logUpstreamError("reading stream from upstream", err)
	}

	return finalUsage
}

func (ps *ProxyServer) handleNormalResponse(c *gin.Context, resp *http.Response) *usageInfo {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logUpstreamError("reading normal response body", err)
		return nil
	}

	// Try to parse usage
	var openAIResp openAIResponse
	if err := json.Unmarshal(bodyBytes, &openAIResp); err == nil {
		if _, err := c.Writer.Write(bodyBytes); err != nil {
			logUpstreamError("writing normal response to client", err)
		}
		if openAIResp.Usage.TotalTokens > 0 {
			return &openAIResp.Usage
		}
	} else {
		// Fallback if not standard OpenAI JSON
		if _, err := c.Writer.Write(bodyBytes); err != nil {
			logUpstreamError("writing normal response to client", err)
		}
	}

	return nil
}
