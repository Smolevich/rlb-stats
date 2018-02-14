package aggregate

import (
	"time"

	"github.com/umputun/rlb-stats/app/candle"
)

// Do aggregate candles in input, 0 maxPoints is unlimited
func Do(candles []candle.Candle, aggInterval time.Duration) (result []candle.Candle) {

	aggInterval = aggInterval.Truncate(time.Minute)
	var firstDate, lastDate = time.Now(), time.Time{}
	for _, c := range candles {
		if c.StartMinute.Before(firstDate) {
			firstDate = c.StartMinute
		}
		if c.StartMinute.After(lastDate) {
			lastDate = c.StartMinute
		}
	}

	for aggTime := firstDate; aggTime.Before(lastDate.Add(aggInterval)); aggTime = aggTime.Add(aggInterval) {
		minuteCandle := candle.NewCandle()
		minuteCandle.StartMinute = aggTime
		for _, c := range candles {
			if c.StartMinute == aggTime || c.StartMinute.After(aggTime) && c.StartMinute.Before(aggTime.Add(aggInterval)) {
				c = updateAndDiscardTime(minuteCandle, c)
			}
		}
		if len(minuteCandle.Nodes) != 0 {
			result = append(result, minuteCandle)
		}
	}
	return result
}

func updateAndDiscardTime(source candle.Candle, appendix candle.Candle) candle.Candle {
	for n := range appendix.Nodes {
		m, ok := source.Nodes[n]
		if !ok {
			m = candle.NewInfo()
		}
		// to calculate mean time multiply source and appendix by their volume and divide everything by total volume
		m.MeanAnswerTime = (m.MeanAnswerTime*time.Duration(m.Volume) + appendix.Nodes[n].MeanAnswerTime*time.Duration(appendix.Nodes[n].Volume)) /
			time.Duration(m.Volume+appendix.Nodes[n].Volume)
		if m.MinAnswerTime > appendix.Nodes[n].MinAnswerTime {
			m.MinAnswerTime = appendix.Nodes[n].MinAnswerTime
		}
		if m.MaxAnswerTime < appendix.Nodes[n].MaxAnswerTime {
			m.MaxAnswerTime = appendix.Nodes[n].MaxAnswerTime
		}
		for file := range appendix.Nodes[n].Files {
			m.Files[file] += appendix.Nodes[n].Files[file]
		}
		m.Volume += appendix.Nodes[n].Volume
		source.Nodes[n] = m
	}
	return source
}
