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
	TTS       *bool  `json:"tts,omitempty"`
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
	LogRecords []LokiLogRecord `json:"logRecords"`
	Scope      OTLPScope       `json:"scope"`
}
type LokiLogRecord struct {
	Attributes           []OTLPAttribute `json:"attributes,omitempty"`
	Body                 OTLPBody        `json:"body"`
	Flags                uint32          `json:"flags,omitempty"`
	ObservedTimeUnixNano string          `json:"observedTimeUnixNano,omitempty"`
	SeverityNumber       int             `json:"severityNumber,omitempty"`
	SeverityText         string          `json:"severityText,omitempty"`
	SpanID               string          `json:"spanId,omitempty"`
	TimeUnixNano         string          `json:"timeUnixNano"`
	TraceID              string          `json:"traceId,omitempty"`
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
	Metrics []PrometheusMetric `json:"metrics"`
	Scope   OTLPScope          `json:"scope"`
}
type PrometheusMetric struct {
	Gauge     *PrometheusGauge     `json:"gauge,omitempty"`
	Histogram *PrometheusHistogram `json:"histogram,omitempty"`
	Name      string               `json:"name"`
	Sum       *PrometheusSum       `json:"sum,omitempty"`
}
type PrometheusGauge struct {
	DataPoints []PrometheusDataPoint `json:"dataPoints"`
}
type PrometheusHistogram struct {
	AggregationTemporality int                        `json:"aggregationTemporality"`
	DataPoints             []PrometheusHistogramPoint `json:"dataPoints"`
}
type PrometheusHistogramPoint struct {
	Attributes     []OTLPAttribute `json:"attributes,omitempty"`
	BucketCounts   []uint64        `json:"bucketCounts"`
	Count          uint64          `json:"count"`
	ExplicitBounds []float64       `json:"explicitBounds"`
	Sum            *float64        `json:"sum,omitempty"`
	TimeUnixNano   string          `json:"timeUnixNano"`
}
type PrometheusSum struct {
	AggregationTemporality int                   `json:"aggregationTemporality"`
	DataPoints             []PrometheusDataPoint `json:"dataPoints"`
	IsMonotonic            bool                  `json:"isMonotonic"`
}
type PrometheusDataPoint struct {
	AsDouble     float64         `json:"asDouble"`
	Attributes   []OTLPAttribute `json:"attributes,omitempty"`
	TimeUnixNano string          `json:"timeUnixNano"`
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
	Attributes        []OTLPAttribute `json:"attributes,omitempty"`
	EndTimeUnixNano   string          `json:"endTimeUnixNano"`
	Flags             uint32          `json:"flags,omitempty"`
	Kind              TypeKind        `json:"kind"`
	Links             []TempoLink     `json:"links,omitempty"`
	Name              string          `json:"name"`
	ParentSpanID      string          `json:"parentSpanId,omitempty"`
	SpanID            string          `json:"spanId"`
	StartTimeUnixNano string          `json:"startTimeUnixNano"`
	Status            TempoStatus     `json:"status,omitempty"`
	TraceID           string          `json:"traceId"`
	TraceState        string          `json:"traceState,omitempty"`
}
type TempoLink struct {
	Attributes []OTLPAttribute `json:"attributes,omitempty"`
	SpanID     string          `json:"spanId"`
	TraceID    string          `json:"traceId"`
	TraceState string          `json:"traceState,omitempty"`
}
type TempoStatus struct {
	Code    TypeStatus `json:"code,omitempty"`
	Message string     `json:"message,omitempty"`
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
			message, _, _, _, err := getDataLog(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid log data: %w", err)
			}
			tts := false
			discordData := DiscordData{
				AvatarURL: avatarURL,
				Content:   message,
				TTS:       &tts,
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
			valueData["_level"] = getLevelText(attributes.typeLevel)
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
			message, traceID, spanID, flags, err := getDataLog(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid log data: %w", err)
			}
			otlp := getOpenTelemetryAttributes(fields, "message", "trace_id", "span_id", "flags")
			now := time.Now().UnixNano()
			lokiData := LokiData{
				ResourceLogs: []LokiResourceLogs{
					{
						Resource: OTLPResource{
							Attributes: otlp.resource,
						},
						ScopeLogs: []LokiScopeLogs{
							{
								LogRecords: []LokiLogRecord{
									{
										Attributes:           otlp.record,
										Body:                 OTLPBody{StringValue: &message},
										Flags:                flags,
										ObservedTimeUnixNano: fmt.Sprintf("%d", now),
										SeverityNumber:       getLevelNumber(attributes.typeLevel),
										SeverityText:         getLevelText(attributes.typeLevel),
										SpanID:               spanID,
										TraceID:              traceID,
										TimeUnixNano:         fmt.Sprintf("%d", now),
									},
								},
								Scope: OTLPScope{
									Name:    "ulog",
									Version: Version,
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
			otlp := getOpenTelemetryAttributes(fields, "name", "value", "type", "count", "sum", "bucket_counts", "explicit_bounds")
			format, err := getOpenTelemetryType(fields)
			if err != nil {
				return nil, err
			}
			var metric PrometheusMetric
			switch format {
			case "counter":
				metric = PrometheusMetric{
					Name: name,
					Sum: &PrometheusSum{
						AggregationTemporality: 2,
						DataPoints: []PrometheusDataPoint{{
							AsDouble:     value,
							Attributes:   otlp.record,
							TimeUnixNano: fmt.Sprintf("%d", time.Now().UnixNano()),
						}},
						IsMonotonic: true,
					},
				}
			case "gauge":
				metric = PrometheusMetric{
					Gauge: &PrometheusGauge{
						DataPoints: []PrometheusDataPoint{{
							AsDouble:     value,
							Attributes:   otlp.record,
							TimeUnixNano: fmt.Sprintf("%d", time.Now().UnixNano()),
						}},
					},
					Name: name,
				}
			case "histogram":
				count, sum, buckets, bounds := getHistogram(fields)
				metric = PrometheusMetric{
					Histogram: &PrometheusHistogram{
						AggregationTemporality: 2,
						DataPoints: []PrometheusHistogramPoint{
							{
								Attributes:     otlp.record,
								BucketCounts:   buckets,
								Count:          count,
								ExplicitBounds: bounds,
								Sum:            &sum,
								TimeUnixNano:   fmt.Sprintf("%d", time.Now().UnixNano()),
							},
						},
					},
					Name: name,
				}
			}
			prometheusData := PrometheusData{
				ResourceMetrics: []PrometheusResourceMetrics{
					{
						Resource: OTLPResource{
							Attributes: otlp.resource,
						},
						ScopeMetrics: []PrometheusScopeMetrics{
							{
								Metrics: []PrometheusMetric{metric},
								Scope: OTLPScope{
									Name:    "ulog",
									Version: Version,
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
			message, _, _, _, err := getDataLog(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid log data: %w", err)
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
			message, _, _, _, err := getDataLog(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid log data: %w", err)
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
			name, traceID, spanID, parentSpanID, traceState, statusCode, statusMessage, links, flags, duration, err := getDataTrace(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid trace data: %w", err)
			}
			otlp := getOpenTelemetryAttributes(fields, "name", "kind", "links", "trace_id", "span_id", "parent_span_id", "duration", "status")
			now := time.Now()
			startNano := now.UnixNano()
			endNano := startNano + duration*1_000_000
			tempoData := TempoData{
				ResourceSpans: []TempoResourceSpans{
					{
						Resource: OTLPResource{
							Attributes: otlp.resource,
						},
						ScopeSpans: []TempoScopeSpan{
							{
								Scope: OTLPScope{
									Name:    "ulog",
									Version: Version,
								},
								Spans: []TempoSpan{
									{
										Attributes:        otlp.record,
										EndTimeUnixNano:   fmt.Sprintf("%d", endNano),
										Flags:             flags,
										Kind:              getKind(fields),
										Links:             links,
										Name:              name,
										ParentSpanID:      parentSpanID,
										SpanID:            spanID,
										StartTimeUnixNano: fmt.Sprintf("%d", startNano),
										Status: TempoStatus{
											Code:    statusCode,
											Message: statusMessage,
										},
										TraceID:    traceID,
										TraceState: traceState,
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
			message, _, _, _, err := getDataLog(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid log data: %w", err)
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
