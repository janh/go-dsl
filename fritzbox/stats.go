// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"3e8.eu/go/dsl/internal/htmlutil"
	"3e8.eu/go/dsl/models"
)

var regexpFilterCharacters = regexp.MustCompile(`\P{L}+`)

type statsItemValue struct {
	Upstream   string `json:"us"`
	Downstream string `json:"ds"`
}

type statsItem struct {
	Title string            `json:"title"`
	Val   []json.RawMessage `json:"val"`
}

type statsJSON struct {
	Data struct {
		NegotiatedValues []statsItem `json:"negotiatedValues"`
		ErrorCounters    []statsItem `json:"errorCounters"`
	} `json:"data"`
}

func parseStats(status *models.Status, d *rawDataStats) {
	var valuesConnection map[string][2]string

	if !d.Legacy {
		var data statsJSON
		json.Unmarshal([]byte(d.Data), &data)

		valuesConnection = parseStatsTableValues(data.Data.NegotiatedValues)
	} else {
		valuesConnection = parseStatsTableValuesLegacy(d.Data)

		parseStatsErrorsLegacy(status, d.Data)
	}

	interpretStatsConnection(status, valuesConnection)
}

func interpretStatsConnection(status *models.Status, values map[string][2]string) {
	status.DownstreamActualRate.IntValue, status.UpstreamActualRate.IntValue =
		interpretStatsIntValues(values, "aktuelledatenrate")
	status.DownstreamAttainableRate.IntValue, status.UpstreamAttainableRate.IntValue =
		interpretStatsIntValues(values, "leitungskapazität")

	status.DownstreamMinimumErrorFreeThroughput.IntValue, status.UpstreamMinimumErrorFreeThroughput.IntValue =
		interpretStatsIntValues(values, "mineffektivedatenrate")

	status.DownstreamBitswap.Enabled, status.UpstreamBitswap.Enabled =
		interpretStatsBoolValues(values, "trägertauschbitswap")
	status.DownstreamSeamlessRateAdaptation.Enabled, status.UpstreamSeamlessRateAdaptation.Enabled =
		interpretStatsBoolValues(values, "nahtloseratenadaption")

	status.DownstreamInterleavingDelay, status.UpstreamInterleavingDelay =
		interpretStatsInterleavingDelays(values, "latenz")
	status.DownstreamImpulseNoiseProtection.FloatValue, status.UpstreamImpulseNoiseProtection.FloatValue =
		interpretStatsFloatValues(values, "impulsstörungsschutzinp")
	status.DownstreamRetransmissionEnabled, status.UpstreamRetransmissionEnabled =
		interpretStatsBoolValues(values, "ginp")

	status.DownstreamVectoringState, status.UpstreamVectoringState =
		interpretStatsVectoringValues(values, "gvector")

	status.DownstreamAttenuation.FloatValue, status.UpstreamAttenuation.FloatValue =
		interpretStatsFloatValues(values, "leitungsdämpfung")
	status.DownstreamSNRMargin.FloatValue, status.UpstreamSNRMargin.FloatValue =
		interpretStatsFloatValues(values, "störabstandsmarge")
}

func parseStatsErrorsLegacy(status *models.Status, dslStats string) {
	doc, err := html.Parse(strings.NewReader(dslStats))
	if err != nil {
		return
	}

	table := htmlutil.FindLastNode(doc, htmlutil.MatcherTagName("table"))
	if table == nil {
		return
	}

	rows := htmlutil.FindAllNodes(table, htmlutil.MatcherTagName("tr"))

	if len(rows) == 4 {

		// versions < 7.19

		columnsNear := htmlutil.FindAllNodes(rows[2], htmlutil.MatcherTagName("td"))
		if len(columnsNear) == 5 {
			status.DownstreamESCount = interpretStatsIntValue(htmlutil.GetText(columnsNear[1]))
			status.DownstreamSESCount = interpretStatsIntValue(htmlutil.GetText(columnsNear[2]))
		}

		columnsFar := htmlutil.FindAllNodes(rows[3], htmlutil.MatcherTagName("td"))
		if len(columnsFar) == 5 {
			status.UpstreamESCount = interpretStatsIntValue(htmlutil.GetText(columnsFar[1]))
			status.UpstreamSESCount = interpretStatsIntValue(htmlutil.GetText(columnsFar[2]))
		}

	} else {

		// versions >= 7.19

		columnsErroredSeconds := htmlutil.FindAllNodes(rows[2], htmlutil.MatcherTagName("td"))
		if len(columnsErroredSeconds) == 3 {
			status.DownstreamESCount = interpretStatsIntValue(htmlutil.GetText(columnsErroredSeconds[1]))
			status.UpstreamESCount = interpretStatsIntValue(htmlutil.GetText(columnsErroredSeconds[2]))
		}

		columnsSeverelyErroredSeconds := htmlutil.FindAllNodes(rows[3], htmlutil.MatcherTagName("td"))
		if len(columnsErroredSeconds) == 3 {
			status.DownstreamSESCount = interpretStatsIntValue(htmlutil.GetText(columnsSeverelyErroredSeconds[1]))
			status.UpstreamSESCount = interpretStatsIntValue(htmlutil.GetText(columnsSeverelyErroredSeconds[2]))
		}

	}
}

func parseStatsTableValues(items []statsItem) map[string][2]string {
	data := make(map[string][2]string)

	for _, item := range items {
		if len(item.Val) < 1 {
			continue
		}

		var val statsItemValue
		err := json.Unmarshal(item.Val[0], &val)
		if err != nil {
			// the value may also be a single string, we are not interested in those
			continue
		}

		key := strings.ToLower(regexpFilterCharacters.ReplaceAllString(item.Title, ""))
		data[key] = [2]string{val.Downstream, val.Upstream}
	}

	return data
}

func parseStatsTableValuesLegacy(dslStats string) map[string][2]string {
	doc, err := html.Parse(strings.NewReader(dslStats))
	if err != nil {
		return nil
	}

	data := make(map[string][2]string)

	table := htmlutil.FindFirstNode(doc, htmlutil.MatcherTagName("table"))
	if table == nil {
		return nil
	}

	rows := htmlutil.FindAllNodes(table, htmlutil.MatcherTagName("tr"))
	for _, row := range rows {
		columns := htmlutil.FindAllNodes(row, htmlutil.MatcherTagName("td"))
		if len(columns) != 4 {
			continue
		}

		key := htmlutil.GetText(columns[0])
		key = strings.ToLower(regexpFilterCharacters.ReplaceAllString(key, ""))

		if len(key) == 0 {
			continue
		}

		val1 := htmlutil.GetText(columns[2])
		val2 := htmlutil.GetText(columns[3])

		data[key] = [2]string{val1, val2}
	}

	return data
}

func interpretStatsIntValue(val string) (out models.IntValue) {
	if valInt, err := strconv.ParseInt(val, 10, 64); err == nil {
		out.Int = valInt
		out.Valid = true
	}
	return
}

func interpretStatsIntValues(values map[string][2]string, key string) (downstream, upstream models.IntValue) {
	if val, ok := values[key]; ok {
		downstream = interpretStatsIntValue(val[0])
		upstream = interpretStatsIntValue(val[1])
	}
	return
}

func interpretStatsFloatValue(val string) (out models.FloatValue) {
	if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
		out.Float = valFloat
		out.Valid = true
	}
	return
}

func interpretStatsFloatValues(values map[string][2]string, key string) (downstream, upstream models.FloatValue) {
	if val, ok := values[key]; ok {
		downstream = interpretStatsFloatValue(val[0])
		upstream = interpretStatsFloatValue(val[1])
	}
	return
}

func interpetStatsInterleavingDelay(val string) (out models.ValueMilliseconds) {
	if val == "fast" || val == "< 1" {
		out.FloatValue.Valid = true
	} else if strings.HasSuffix(val, "ms") {
		val = strings.TrimSpace(val[0 : len(val)-2])
		if valFloat, err := strconv.ParseFloat(val, 64); err == nil {
			out.FloatValue.Float = valFloat
			out.FloatValue.Valid = true
		}
	}
	return
}

func interpretStatsInterleavingDelays(values map[string][2]string, key string) (downstream, upstream models.ValueMilliseconds) {
	if val, ok := values[key]; ok {
		downstream = interpetStatsInterleavingDelay(val[0])
		upstream = interpetStatsInterleavingDelay(val[1])
	}
	return
}

func interpretStatsBoolValue(val string) (out models.BoolValue) {
	if val == "an" || val == "aus" {
		out.Bool = val == "an"
		out.Valid = true
	}
	return
}

func interpretStatsBoolValues(values map[string][2]string, key string) (downstream, upstream models.BoolValue) {
	if val, ok := values[key]; ok {
		downstream = interpretStatsBoolValue(val[0])
		upstream = interpretStatsBoolValue(val[1])
	}
	return
}

func interpretStatsVectoringValue(val string) (out models.VectoringValue) {
	if val == "aus" {
		out.State = models.VectoringStateOff
		out.Valid = true
	} else if val == "friendly" {
		out.State = models.VectoringStateFriendly
		out.Valid = true
	} else if val == "full" {
		out.State = models.VectoringStateFull
		out.Valid = true
	}
	return
}

func interpretStatsVectoringValues(values map[string][2]string, key string) (downstream, upstream models.VectoringValue) {
	if val, ok := values[key]; ok {
		downstream = interpretStatsVectoringValue(val[0])
		upstream = interpretStatsVectoringValue(val[1])
	}
	return
}
