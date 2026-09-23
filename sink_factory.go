// ULOG (Universal Logger Observability Gateway)
// Copyright (C) 2026 Mikhail Dadaev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package ulog

import (
	"encoding/json"
	"fmt"
	"time"
)

// Публичные структуры
type DiscordData struct {
	AvatarURL string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
	TTS       bool   `json:"tts,omitempty"`
	UserName  string `json:"username,omitempty"`
}
type SinkDiscord = SinkHttp
type KafkaData struct {
	Headers   map[string]string `json:"headers,omitempty"`
	Key       string            `json:"key,omitempty"`
	Partition *int32            `json:"partition,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Topic     string            `json:"topic,omitempty"`
	Value     json.RawMessage   `json:"value"`
}
type SinkKafka = SinkHttp
type LokiData struct {
	ResourceLogs []LokiResourceLogs `json:"resourceLogs"`
}
type LokiResourceLogs struct {
	Resource  OTLPResource    `json:"resource"`
	ScopeLogs []LokiScopeLogs `json:"scopeLogs"`
}
type LokiScopeLogs struct {
	Scope      OTLPScope       `json:"scope"`
	LogRecords []LokiLogRecord `json:"logRecords"`
}
type LokiLogRecord struct {
	TimeUnixNano string          `json:"timeUnixNano"`
	SeverityText string          `json:"severityText,omitempty"`
	Body         OTLPBody        `json:"body"`
	Attributes   []OTLPAttribute `json:"attributes,omitempty"`
}
type SinkLoki = SinkHttp
type PrometheusData struct {
	ResourceMetrics []PrometheusResourceMetrics `json:"resourceMetrics"`
}
type PrometheusResourceMetrics struct {
	Resource     OTLPResource             `json:"resource"`
	ScopeMetrics []PrometheusScopeMetrics `json:"scopeMetrics"`
}
type PrometheusScopeMetrics struct {
	Scope   OTLPScope          `json:"scope"`
	Metrics []PrometheusMetric `json:"metrics"`
}
type PrometheusMetric struct {
	Name  string          `json:"name"`
	Gauge PrometheusGauge `json:"gauge"`
}
type PrometheusGauge struct {
	DataPoints []PrometheusDataPoint `json:"dataPoints"`
}
type PrometheusDataPoint struct {
	TimeUnixNano string          `json:"timeUnixNano"`
	AsDouble     float64         `json:"asDouble"`
	Attributes   []OTLPAttribute `json:"attributes,omitempty"`
}
type SinkPrometheus = SinkHttp
type SlackData struct {
	Channel   string `json:"channel,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	Text      string `json:"text"`
	UserName  string `json:"username,omitempty"`
}
type SinkSlack = SinkHttp
type TelegramData struct {
	ChatID              string `json:"chat_id"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
	Text                string `json:"text"`
	ParseMode           string `json:"parse_mode,omitempty"`
}
type SinkTelegram = SinkHttp
type TempoData struct {
	ResourceSpans []TempoResourceSpans `json:"resourceSpans"`
}
type TempoResourceSpans struct {
	Resource   OTLPResource     `json:"resource"`
	ScopeSpans []TempoScopeSpan `json:"scopeSpans"`
}
type TempoScopeSpan struct {
	Scope OTLPScope   `json:"scope"`
	Spans []TempoSpan `json:"spans"`
}
type TempoSpan struct {
	TraceID           string          `json:"traceId"`
	SpanID            string          `json:"spanId"`
	Name              string          `json:"name"`
	Kind              int             `json:"kind"`
	StartTimeUnixNano string          `json:"startTimeUnixNano"`
	EndTimeUnixNano   string          `json:"endTimeUnixNano"`
	Attributes        []OTLPAttribute `json:"attributes,omitempty"`
}
type SinkTempo = SinkHttp
type WechatData struct {
	Content             string   `json:"content"`
	MsgType             string   `json:"msgtype"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}
type SinkWechat = SinkHttp

// OpenTelemetry
type OTLPAttribute struct {
	Key   string        `json:"key"`
	Value OTLPAttrValue `json:"value"`
}
type OTLPAttrValue struct {
	ArrayValue  []OTLPAttrValue `json:"arrayValue,omitempty"`
	BoolValue   bool            `json:"boolValue,omitempty"`
	DoubleValue float64         `json:"doubleValue,omitempty"`
	IntValue    string          `json:"intValue,omitempty"`
	StringValue string          `json:"stringValue,omitempty"`
}
type OTLPBody struct {
	BoolValue   *bool    `json:"boolValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
	IntValue    *string  `json:"intValue,omitempty"`
	StringValue *string  `json:"stringValue,omitempty"`
}
type OTLPResource struct {
	Attributes []OTLPAttribute `json:"attributes"`
}
type OTLPScope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// Публичные конструкторы
func NewSinkDiscord(endPoint, userName, avatarURL string, params ...httpParams) *SinkDiscord {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			message := getDataLog(fields)
			if message == "" {
				message = "empty message"
			}
			discordData := DiscordData{
				AvatarURL: avatarURL,
				Content:   message,
				TTS:       false,
				UserName:  userName,
			}
			return json.Marshal(discordData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewSinkKafka(endPoint string, params ...httpParams) *SinkKafka {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpBatch(100, 5*time.Second),
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelInfo),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			valueData := getKafkaAttributes(fields)
			valueData["_level"] = getLevel(attributes.typeLevel)
			valueData["_type"] = getData(attributes.typeData)
			valueData["_timestamp"] = time.Now().Format(time.RFC3339Nano)
			valueJSON, err := json.Marshal(valueData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value: %w", err)
			}
			key := getKafkaKey(fields)
			records := struct {
				Records []KafkaData `json:"records"`
			}{
				Records: []KafkaData{
					{
						Key:       key,
						Value:     valueJSON,
						Timestamp: time.Now(),
					},
				},
			}
			return json.Marshal(records)
		}),
		WithHttpHeader("Content-Type", "application/vnd.kafka.json.v2+json"),
		WithHttpHeader("Accept", "application/vnd.kafka.v2+json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewSinkLoki(endPoint string, params ...httpParams) *SinkLoki {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelInfo),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			message := getDataLog(fields)
			if message == "" {
				message = "empty message"
			}
			attrs := getOpenTelemetryAttributes(fields, "message")
			now := time.Now().UnixNano()
			lokiData := LokiData{
				ResourceLogs: []LokiResourceLogs{
					{
						Resource: OTLPResource{
							Attributes: []OTLPAttribute{
								{
									Key:   "service.name",
									Value: OTLPAttrValue{StringValue: "ulog"},
								},
							},
						},
						ScopeLogs: []LokiScopeLogs{
							{
								Scope: OTLPScope{
									Name:    "ulog",
									Version: Version,
								},
								LogRecords: []LokiLogRecord{
									{
										TimeUnixNano: fmt.Sprintf("%d", now),
										SeverityText: getLevel(attributes.typeLevel),
										Body:         OTLPBody{StringValue: &message},
										Attributes:   attrs,
									},
								},
							},
						},
					},
				},
			}
			return json.Marshal(lokiData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
	}, params...)...)
}
func NewSinkPrometheus(endPoint string, params ...httpParams) *SinkPrometheus {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpFilterData(DataMetric),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			name, value := getDataMetric(fields)
			attrs := getOpenTelemetryAttributes(fields, "name", "value")
			prometheusData := PrometheusData{
				ResourceMetrics: []PrometheusResourceMetrics{
					{
						Resource: OTLPResource{
							Attributes: []OTLPAttribute{
								{
									Key:   "service.name",
									Value: OTLPAttrValue{StringValue: "ulog"},
								},
							},
						},
						ScopeMetrics: []PrometheusScopeMetrics{
							{
								Scope: OTLPScope{
									Name:    "ulog",
									Version: Version,
								},
								Metrics: []PrometheusMetric{
									{
										Name: name,
										Gauge: PrometheusGauge{
											DataPoints: []PrometheusDataPoint{
												{
													TimeUnixNano: fmt.Sprintf("%d", time.Now().UnixNano()),
													AsDouble:     value,
													Attributes:   attrs,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			return json.Marshal(prometheusData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
	}, params...)...)
}
func NewSinkSlack(endPoint, userName, iconEmoji, iconURL, channel string, params ...httpParams) *SinkSlack {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			message := getDataLog(fields)
			if message == "" {
				message = "empty message"
			}
			slackData := SlackData{
				Channel:   channel,
				IconEmoji: iconEmoji,
				IconURL:   iconURL,
				Text:      message,
				UserName:  userName,
			}
			return json.Marshal(slackData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewSinkTelegram(endPoint, chatID string, params ...httpParams) *SinkTelegram {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			message := getDataLog(fields)
			if message == "" {
				message = "empty message"
			}
			telegramData := TelegramData{
				ChatID:    chatID,
				Text:      message,
				ParseMode: "HTML",
			}
			return json.Marshal(telegramData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewSinkTempo(endPoint string, params ...httpParams) *SinkTempo {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpFilterData(DataTrace),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			name, traceID, spanID, duration, err := getDataTrace(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid trace data: %w", err)
			}
			attrs := getOpenTelemetryAttributes(fields, "name", "trace_id", "span_id", "duration")
			now := time.Now()
			startNano := now.UnixNano()
			endNano := startNano + duration*1_000_000
			tempoData := TempoData{
				ResourceSpans: []TempoResourceSpans{
					{
						Resource: OTLPResource{
							Attributes: []OTLPAttribute{
								{
									Key:   "service.name",
									Value: OTLPAttrValue{StringValue: "ulog"},
								},
							},
						},
						ScopeSpans: []TempoScopeSpan{
							{
								Scope: OTLPScope{
									Name:    "ulog",
									Version: Version,
								},
								Spans: []TempoSpan{
									{
										TraceID:           traceID,
										SpanID:            spanID,
										Name:              name,
										Kind:              1,
										StartTimeUnixNano: fmt.Sprintf("%d", startNano),
										EndTimeUnixNano:   fmt.Sprintf("%d", endNano),
										Attributes:        attrs,
									},
								},
							},
						},
					},
				},
			}
			return json.Marshal(tempoData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
	}, params...)...)
}
func NewSinkWechat(endPoint string, params ...httpParams) *SinkWechat {
	return NewSinkHttp(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			message := getDataLog(fields)
			if message == "" {
				message = "empty message"
			}
			wechatData := WechatData{
				Content: message,
				MsgType: "markdown",
			}
			return json.Marshal(wechatData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
