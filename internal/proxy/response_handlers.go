package proxy

import (
	"bufio"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

type openAIResponse struct {
	Usage usageInfo `json:"usage"`
}

type geminiResponse struct {
	UsageMetadata usageInfo `json:"usageMetadata"`
}

func (ps *ProxyServer) handleStreamingResponse(c *gin.Context, resp *http.Response) (*usageInfo, string) {
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
	var responseBodyBuilder strings.Builder
	const maxBodySize = 262144 // 256KB truncate limit
	
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Record the response body
		if responseBodyBuilder.Len() < maxBodySize {
			responseBodyBuilder.WriteString(line)
			responseBodyBuilder.WriteString("\n")
		}

		if _, err := c.Writer.Write([]byte(line + "\n")); err != nil {
			logUpstreamError("writing stream to client", err)
			return nil, responseBodyBuilder.String()
		}
		flusher.Flush()

		// Parse usage if present in the data line
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data != "[DONE]" {
				if strings.Contains(data, "\"usage\"") {
					var uResp openAIResponse
					if err := json.Unmarshal([]byte(data), &uResp); err == nil {
						uResp.Usage.Normalize()
						if uResp.Usage.TotalTokens > 0 {
							finalUsage = &uResp.Usage
						} else {
							// Log if usage found but tokens=0, might be a parsing issue
							logrus.Debugf("Usage block detected in stream but parsed as 0 tokens. Data: %s", data)
						}
					}
				} else if strings.Contains(data, "\"usageMetadata\"") {
					var gResp geminiResponse
					if err := json.Unmarshal([]byte(data), &gResp); err == nil {
						gResp.UsageMetadata.Normalize()
						// Special case for Gemini: it might send usage in segments
						if gResp.UsageMetadata.TotalTokens > 0 {
							if finalUsage == nil {
								finalUsage = &gResp.UsageMetadata
							} else {
								// Accumulate or pick the largest (Gemini usually sends cumulative at the end)
								if gResp.UsageMetadata.TotalTokens > finalUsage.TotalTokens {
									finalUsage = &gResp.UsageMetadata
								}
							}
						} else {
							logrus.Debugf("usageMetadata block detected in stream but parsed as 0 tokens. Data: %s", data)
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logUpstreamError("reading stream from upstream", err)
	}

	return finalUsage, responseBodyBuilder.String()
}

func (ps *ProxyServer) handleNormalResponse(c *gin.Context, resp *http.Response) (*usageInfo, string) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logUpstreamError("reading normal response body", err)
		return nil, ""
	}

	// Try to parse usage
	var openAIResp openAIResponse
	if err := json.Unmarshal(bodyBytes, &openAIResp); err == nil && strings.Contains(string(bodyBytes), "\"usage\"") {
		if _, err := c.Writer.Write(bodyBytes); err != nil {
			logUpstreamError("writing normal response to client", err)
		}
		openAIResp.Usage.Normalize()
		if openAIResp.Usage.TotalTokens > 0 {
			return &openAIResp.Usage, string(bodyBytes)
		}
		logrus.Debugf("Usage block detected in normal response but parsed as 0 tokens. Body: %s", string(bodyBytes))
	} else {
		var gResp geminiResponse
		if err := json.Unmarshal(bodyBytes, &gResp); err == nil && strings.Contains(string(bodyBytes), "\"usageMetadata\"") {
			if _, err := c.Writer.Write(bodyBytes); err != nil {
				logUpstreamError("writing normal response to client", err)
			}
			gResp.UsageMetadata.Normalize()
			if gResp.UsageMetadata.TotalTokens > 0 {
				return &gResp.UsageMetadata, string(bodyBytes)
			}
			logrus.Debugf("usageMetadata block detected in normal response but parsed as 0 tokens. Body: %s", string(bodyBytes))
		}
	}
	// Fallback if not standard OpenAI JSON
	if _, err := c.Writer.Write(bodyBytes); err != nil {
		logUpstreamError("writing normal response to client", err)
	}

	return nil, string(bodyBytes)
}
