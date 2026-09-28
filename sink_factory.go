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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
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
			data, err := getUniversalAlertData(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid alert data: %w", err)
			}
			tts := false
			discordData := DiscordData{
				AvatarURL: avatarURL,
				Content:   data,
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
			otlp := formatOpenTelemetryAttributes(fields, skipKeysLoki...)
			data, err := getLokiData(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid log data: %w", err)
			}
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
										Body:                 OTLPBody{StringValue: &data.message},
										Flags:                data.flags,
										ObservedTimeUnixNano: fmt.Sprintf("%d", now),
										SeverityNumber:       getLevelNumber(attributes.typeLevel),
										SeverityText:         getLevelText(attributes.typeLevel),
										SpanID:               data.spanID,
										TraceID:              data.traceID,
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
			otlp := formatOpenTelemetryAttributes(fields, skipKeysPrometheus...)
			data, err := getPrometheusData(fields)
			if err != nil {
				return nil, err
			}
			var metric PrometheusMetric
			switch data.format {
			case "counter":
				metric = PrometheusMetric{
					Name: data.name,
					Sum: &PrometheusSum{
						AggregationTemporality: 2,
						DataPoints: []PrometheusDataPoint{{
							AsDouble:     data.value,
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
							AsDouble:     data.value,
							Attributes:   otlp.record,
							TimeUnixNano: fmt.Sprintf("%d", time.Now().UnixNano()),
						}},
					},
					Name: data.name,
				}
			case "histogram":
				metric = PrometheusMetric{
					Histogram: &PrometheusHistogram{
						AggregationTemporality: 2,
						DataPoints: []PrometheusHistogramPoint{{
							Attributes:     otlp.record,
							BucketCounts:   data.buckets,
							Count:          data.count,
							ExplicitBounds: data.bounds,
							Sum:            &data.sum,
							TimeUnixNano:   fmt.Sprintf("%d", time.Now().UnixNano()),
						}},
					},
					Name: data.name,
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
			data, err := getUniversalAlertData(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid alert data: %w", err)
			}
			slackData := SlackData{
				Channel:   channel,
				IconEmoji: iconEmoji,
				IconURL:   iconURL,
				Text:      data,
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
			data, err := getUniversalAlertData(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid alert data: %w", err)
			}
			telegramData := TelegramData{
				ChatID:    chatID,
				Text:      data,
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
			otlp := formatOpenTelemetryAttributes(fields, skipKeysTempo...)
			data, err := getTempoData(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid trace data: %w", err)
			}
			now := time.Now()
			startNano := now.UnixNano()
			endNano := startNano + data.duration*1_000_000
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
										Flags:             data.flags,
										Kind:              data.kind,
										Links:             data.links,
										Name:              data.name,
										ParentSpanID:      data.parentSpanID,
										SpanID:            data.spanID,
										StartTimeUnixNano: fmt.Sprintf("%d", startNano),
										Status: TempoStatus{
											Code:    data.statusCode,
											Message: data.statusMessage,
										},
										TraceID:    data.traceID,
										TraceState: data.traceState,
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
			data, err := getUniversalAlertData(fields)
			if err != nil {
				return nil, fmt.Errorf("invalid alert data: %w", err)
			}
			wechatData := WechatData{
				Content: data,
				MsgType: "markdown",
			}
			return json.Marshal(wechatData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}

// Приватные переменные
var (
	skipKeysLoki = []string{
		"message", "trace_id", "span_id", "flags",
	}
	skipKeysPrometheus = []string{
		"name", "value", "type", "count", "sum", "bucket_counts", "explicit_bounds",
	}
	skipKeysTempo = []string{
		"name", "kind", "links", "trace_id", "span_id", "parent_span_id", "duration", "status", "trace_state", "flags",
	}
)

// Приватные структуры
type lokiData struct {
	flags   uint32
	message string
	spanID  string
	traceID string
}
type prometheusData struct {
	buckets []uint64
	bounds  []float64
	count   uint64
	format  string
	name    string
	sum     float64
	value   float64
}
type tempoData struct {
	duration      int64
	flags         uint32
	kind          TypeKind
	links         []TempoLink
	name          string
	parentSpanID  string
	spanID        string
	statusCode    TypeStatus
	statusMessage string
	traceID       string
	traceState    string
}

// Приватные функции
func appendOpenTelemetryAttributeRecord(attrs *[]OTLPAttribute, f Field) {
	switch f.typeValue {
	case FieldString:
		v := f.valueString
		switch f.nameKey {
		case "trace_id":
			normalized, err := normalizeTraceID(v)
			if err != nil {
				fmt.Fprintf(DefaultWriterErr, "ulog: skipping invalid trace_id: %v\n", err)
				return
			}
			v = normalized
		case "span_id":
			normalized, err := normalizeSpanID(v)
			if err != nil {
				fmt.Fprintf(DefaultWriterErr, "ulog: skipping invalid span_id: %v\n", err)
				return
			}
			v = normalized
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{StringValue: v},
		})
	case FieldInt:
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{IntValue: fmt.Sprintf("%d", f.valueInt)},
		})
	case FieldInt64:
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{IntValue: fmt.Sprintf("%d", f.valueInt64)},
		})
	case FieldFloat64:
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{DoubleValue: f.valueFloat64},
		})
	case FieldBool:
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{BoolValue: f.valueBool},
		})
	case FieldDuration:
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{StringValue: f.valueDuration.String()},
		})
	case FieldTime:
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{StringValue: f.valueTime.Format(time.RFC3339Nano)},
		})
	case FieldStrings:
		arr := make([]OTLPAttrValue, len(f.valueStrings))
		for i, v := range f.valueStrings {
			arr[i] = OTLPAttrValue{StringValue: v}
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{ArrayValue: arr},
		})
	case FieldInts:
		arr := make([]OTLPAttrValue, len(f.valueInts))
		for i, v := range f.valueInts {
			arr[i] = OTLPAttrValue{IntValue: fmt.Sprintf("%d", v)}
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{ArrayValue: arr},
		})
	case FieldInts64:
		arr := make([]OTLPAttrValue, len(f.valueInts64))
		for i, v := range f.valueInts64 {
			arr[i] = OTLPAttrValue{IntValue: fmt.Sprintf("%d", v)}
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{ArrayValue: arr},
		})
	case FieldFloats64:
		arr := make([]OTLPAttrValue, len(f.valueFloats64))
		for i, v := range f.valueFloats64 {
			arr[i] = OTLPAttrValue{DoubleValue: v}
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{ArrayValue: arr},
		})
	case FieldBools:
		arr := make([]OTLPAttrValue, len(f.valueBools))
		for i, v := range f.valueBools {
			arr[i] = OTLPAttrValue{BoolValue: v}
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{ArrayValue: arr},
		})
	case FieldDurations:
		arr := make([]OTLPAttrValue, len(f.valueDurations))
		for i, v := range f.valueDurations {
			arr[i] = OTLPAttrValue{StringValue: v.String()}
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{ArrayValue: arr},
		})
	case FieldTimes:
		arr := make([]OTLPAttrValue, len(f.valueTimes))
		for i, v := range f.valueTimes {
			arr[i] = OTLPAttrValue{StringValue: v.Format(time.RFC3339Nano)}
		}
		*attrs = append(*attrs, OTLPAttribute{
			Key:   f.nameKey,
			Value: OTLPAttrValue{ArrayValue: arr},
		})
	}
}
func formatOpenTelemetryAttributes(fields []Field, skipKeys ...string) otlpAttributes {
	result := otlpAttributes{
		resource: make([]OTLPAttribute, 0, 5),
		record:   make([]OTLPAttribute, 0, len(fields)),
	}
	hasService := false
	for _, f := range fields {
		if f.typeValue == FieldString {
			switch f.nameKey {
			case "environment":
				result.resource = append(result.resource, OTLPAttribute{
					Key:   "deployment.environment.name",
					Value: OTLPAttrValue{StringValue: f.valueString},
				})
				continue
			case "instance_id":
				result.resource = append(result.resource, OTLPAttribute{
					Key:   "service.instance.id",
					Value: OTLPAttrValue{StringValue: f.valueString},
				})
				continue
			case "service":
				result.resource = append(result.resource, OTLPAttribute{
					Key:   "service.name",
					Value: OTLPAttrValue{StringValue: f.valueString},
				})
				hasService = true
				continue
			case "namespace":
				result.resource = append(result.resource, OTLPAttribute{
					Key:   "service.namespace",
					Value: OTLPAttrValue{StringValue: f.valueString},
				})
				continue
			case "version":
				result.resource = append(result.resource, OTLPAttribute{
					Key:   "service.version",
					Value: OTLPAttrValue{StringValue: f.valueString},
				})
				continue
			}
		}
		skip := false
		for _, k := range skipKeys {
			if f.nameKey == k {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		appendOpenTelemetryAttributeRecord(&result.record, f)
	}
	if !hasService {
		result.resource = append(result.resource, OTLPAttribute{
			Key:   "service.name",
			Value: OTLPAttrValue{StringValue: "ulog"},
		})
	}
	return result
}
func getKafkaAttributes(fields []Field) map[string]any {
	valueData := make(map[string]any, len(fields))
	for _, field := range fields {
		v := getUniversalFieldValue(field)
		if field.typeValue == FieldString {
			switch field.nameKey {
			case "trace_id":
				if normalized, err := normalizeTraceID(field.valueString); err == nil {
					v = normalized
				}
			case "span_id":
				if normalized, err := normalizeSpanID(field.valueString); err == nil {
					v = normalized
				}
			}
		}
		valueData[field.nameKey] = v
	}
	return valueData
}
func getKafkaKey(fields []Field) string {
	priorities := []string{"trace_id", "node_id", "user_id", "request_id"}
	for _, k := range priorities {
		for _, field := range fields {
			if field.nameKey != k || field.typeValue != FieldString {
				continue
			}
			if k == "trace_id" {
				if normalized, err := normalizeTraceID(field.valueString); err == nil {
					return normalized
				}
				continue
			}
			return field.valueString
		}
	}
	return ""
}
func getLokiData(fields []Field) (lokiData, error) {
	var (
		result lokiData
		err    error
	)
	for _, f := range fields {
		switch f.nameKey {
		case "message":
			if f.typeValue == FieldString {
				result.message = f.valueString
			}
		case "trace_id":
			if f.typeValue == FieldString {
				result.traceID, err = normalizeTraceID(f.valueString)
				if err != nil {
					return lokiData{}, fmt.Errorf("invalid trace_id: %w", err)
				}
			}
		case "span_id":
			if f.typeValue == FieldString {
				result.spanID, err = normalizeSpanID(f.valueString)
				if err != nil {
					return lokiData{}, fmt.Errorf("invalid span_id: %w", err)
				}
			}
		case "flags":
			switch f.typeValue {
			case FieldInt:
				result.flags = uint32(f.valueInt)
			case FieldInt64:
				result.flags = uint32(f.valueInt64)
			case FieldBool:
				if f.valueBool {
					result.flags = TraceFlagsSampled
				}
			}
		}
	}
	if result.message == "" {
		result.message = "empty message"
	}
	return result, nil
}
func getPrometheusData(fields []Field) (prometheusData, error) {
	var result prometheusData
	for _, field := range fields {
		switch field.nameKey {
		case "name":
			if field.typeValue == FieldString {
				result.name = field.valueString
			}
		case "value":
			switch field.typeValue {
			case FieldFloat64:
				result.value = field.valueFloat64
			case FieldInt64:
				result.value = float64(field.valueInt64)
			case FieldInt:
				result.value = float64(field.valueInt)
			case FieldDuration:
				result.value = float64(field.valueDuration.Milliseconds())
			default:
				fmt.Fprintf(DefaultWriterErr, "ulog: unsupported metric value type: %d\n", field.typeValue)
			}
		case "type":
			if field.typeValue == FieldString {
				switch field.valueString {
				case "counter", "gauge", "histogram":
					result.format = field.valueString
				}
			}
		case "count":
			switch field.typeValue {
			case FieldInt64:
				result.count = uint64(field.valueInt64)
			case FieldInt:
				result.count = uint64(field.valueInt)
			}
		case "sum":
			if field.typeValue == FieldFloat64 {
				result.sum = field.valueFloat64
			}
		case "bucket_counts":
			switch field.typeValue {
			case FieldInts64:
				result.buckets = make([]uint64, len(field.valueInts64))
				for i, v := range field.valueInts64 {
					result.buckets[i] = uint64(v)
				}
			case FieldInts:
				result.buckets = make([]uint64, len(field.valueInts))
				for i, v := range field.valueInts {
					result.buckets[i] = uint64(v)
				}
			}
		case "explicit_bounds":
			if field.typeValue == FieldFloats64 {
				result.bounds = field.valueFloats64
			}
		}
	}
	if result.name == "" {
		result.name = "unnamed-metric"
	}
	if result.format == "" {
		return prometheusData{}, fmt.Errorf("ulog: metric type is required (use String(\"type\", \"counter\"|\"gauge\"|\"histogram\"))")
	}
	return result, nil
}
func getTempoData(fields []Field) (tempoData, error) {
	var (
		rawTraceID      string
		rawSpanID       string
		rawParentSpanID string
		rawName         string
		rawStatus       string
		rawDur          int64
		hasDur          bool
		result          tempoData
	)
	for _, f := range fields {
		switch f.nameKey {
		case "trace_id":
			if f.typeValue == FieldString {
				rawTraceID = f.valueString
			}
		case "span_id":
			if f.typeValue == FieldString {
				rawSpanID = f.valueString
			}
		case "trace_state":
			if f.typeValue == FieldString {
				result.traceState = f.valueString
			}
		case "flags":
			switch f.typeValue {
			case FieldInt:
				result.flags = uint32(f.valueInt)
			case FieldInt64:
				result.flags = uint32(f.valueInt64)
			case FieldBool:
				if f.valueBool {
					result.flags = TraceFlagsSampled
				}
			}
		case "parent_span_id":
			if f.typeValue == FieldString {
				rawParentSpanID = f.valueString
			}
		case "name":
			if f.typeValue == FieldString {
				rawName = f.valueString
			}
		case "kind":
			if f.typeValue == FieldInt {
				result.kind = TypeKind(f.valueInt)
			}
		case "status":
			if f.typeValue == FieldString {
				rawStatus = f.valueString
			}
		case "links":
			if f.typeValue == FieldStrings {
				for _, raw := range f.valueStrings {
					link, parseErr := getTempoDataLink(raw)
					if parseErr != nil {
						return tempoData{}, fmt.Errorf("invalid link: %w", parseErr)
					}
					result.links = append(result.links, link)
				}
			}
		case "duration":
			var ms int64
			switch f.typeValue {
			case FieldInt:
				ms = int64(f.valueInt)
			case FieldInt64:
				ms = f.valueInt64
			case FieldDuration:
				ms = f.valueDuration.Milliseconds()
			case FieldString:
				d, parseErr := time.ParseDuration(f.valueString)
				if parseErr != nil {
					return tempoData{}, fmt.Errorf("invalid duration string: %w", parseErr)
				}
				ms = d.Milliseconds()
			}
			rawDur = ms
			hasDur = true
		}
	}
	if rawTraceID == "" {
		return tempoData{}, fmt.Errorf("trace_id is required")
	}
	traceID, err := normalizeTraceID(rawTraceID)
	if err != nil {
		return tempoData{}, err
	}
	result.traceID = traceID
	if rawSpanID == "" {
		return tempoData{}, fmt.Errorf("span_id is required")
	}
	spanID, err := normalizeSpanID(rawSpanID)
	if err != nil {
		return tempoData{}, err
	}
	result.spanID = spanID
	if rawParentSpanID != "" {
		parentSpanID, err := normalizeSpanID(rawParentSpanID)
		if err != nil {
			return tempoData{}, fmt.Errorf("invalid parent_span_id: %w", err)
		}
		result.parentSpanID = parentSpanID
	}
	switch strings.ToLower(rawStatus) {
	case "ok", "success":
		result.statusCode = StatusOK
	case "error", "failed":
		result.statusCode = StatusError
		result.statusMessage = rawStatus
	default:
		result.statusCode = StatusUnset
	}
	result.name = rawName
	if result.name == "" {
		result.name = "unnamed-trace"
	}
	switch {
	case hasDur && rawDur > 0:
		result.duration = rawDur
	case hasDur && rawDur <= 0:
		return tempoData{}, fmt.Errorf("duration must be positive, got %d", rawDur)
	default:
		result.duration = 1
	}
	return result, nil
}
func getTempoDataLink(raw string) (TempoLink, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 {
		return TempoLink{}, fmt.Errorf("expected 'trace_id:span_id[:trace_state]', got %q", raw)
	}
	traceID, err := normalizeTraceID(parts[0])
	if err != nil {
		return TempoLink{}, fmt.Errorf("invalid link trace_id: %w", err)
	}
	spanID, err := normalizeSpanID(parts[1])
	if err != nil {
		return TempoLink{}, fmt.Errorf("invalid link span_id: %w", err)
	}
	link := TempoLink{
		TraceID: traceID,
		SpanID:  spanID,
	}
	if len(parts) == 3 {
		link.TraceState = parts[2]
	}
	return link, nil
}
func getUniversalAlertData(fields []Field) (string, error) {
	for _, f := range fields {
		if f.nameKey == "message" && f.typeValue == FieldString {
			if f.valueString == "" {
				return "empty message", nil
			}
			return f.valueString, nil
		}
	}
	return "empty message", nil
}
func getUniversalFieldValue(field Field) any {
	if extractor, ok := fieldExtractor[field.typeValue]; ok {
		return extractor(field)
	}
	return nil
}
func normalizeTraceID(value string) (string, error) {
	v := strings.ToLower(strings.ReplaceAll(value, "-", ""))
	if len(v) != 32 {
		return "", fmt.Errorf("invalid trace_id length: got %d, want 32 (input: %q)", len(v), value)
	}
	if _, err := hex.DecodeString(v); err != nil {
		return "", fmt.Errorf("invalid trace_id hex: %w (input: %q)", err, value)
	}
	return v, nil
}
func normalizeSpanID(value string) (string, error) {
	v := strings.ToLower(strings.ReplaceAll(value, "-", ""))
	if len(v) != 16 {
		return "", fmt.Errorf("invalid span_id length: got %d, want 16 (input: %q)", len(v), value)
	}
	if _, err := hex.DecodeString(v); err != nil {
		return "", fmt.Errorf("invalid span_id hex: %w (input: %q)", err, value)
	}
	return v, nil
}
