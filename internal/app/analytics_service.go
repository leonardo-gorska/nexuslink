package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
	"github.com/leonardo-gorska/nexuslink/internal/port/output"
	"github.com/mssola/useragent"
	"github.com/oschwald/geoip2-golang"
)

type AnalyticsService struct {
	store     output.AnalyticsStore
	geoReader *geoip2.Reader
}

func NewAnalyticsService(store output.AnalyticsStore, geoReader *geoip2.Reader) *AnalyticsService {
	return &AnalyticsService{
		store:     store,
		geoReader: geoReader,
	}
}

func (s *AnalyticsService) ProcessClickBatch(ctx context.Context, batch []entity.ClickEvent) error {
	if len(batch) == 0 {
		return nil
	}

	dailyMap := make(map[string]*entity.DailyAnalytics)

	for i := range batch {
		ev := &batch[i]

		// Parse User-Agent
		ua := useragent.New(ev.UserAgent)
		browser, _ := ua.Browser()
		ev.Browser = browser
		ev.OS = ua.OS()
		if ua.Mobile() {
			ev.Device = "mobile"
		} else if ua.Bot() {
			ev.Device = "bot"
		} else {
			ev.Device = "desktop"
		}

		// Parse GeoIP
		if s.geoReader != nil && ev.IP != "" {
			ip := net.ParseIP(ev.IP)
			if ip != nil {
				record, err := s.geoReader.Country(ip)
				if err == nil && record.Country.IsoCode != "" {
					ev.Country = record.Country.IsoCode
				}
			}
		}

		// Aggregate for DailyAnalytics
		date := ev.ClickedAt.Truncate(24 * time.Hour)
		key := fmt.Sprintf("%s-%s-%s-%s-%s", ev.LinkHash, date.Format(time.DateOnly), ev.Country, ev.Device, ev.Browser)

		if da, ok := dailyMap[key]; ok {
			da.ClickCount++
			// Using 1 for simplicity of this implementation scope instead of tracking accurate IP distincts per batch
		} else {
			dailyMap[key] = &entity.DailyAnalytics{
				LinkHash:   ev.LinkHash,
				Date:       date,
				Country:    ev.Country,
				Device:     ev.Device,
				Browser:    ev.Browser,
				ClickCount: 1,
				UniqueIPs:  1, // Assumption per batch element
			}
		}
	}

	var dailyStats []entity.DailyAnalytics
	for _, v := range dailyMap {
		dailyStats = append(dailyStats, *v)
	}

	// 1. Insert Raw Events
	if err := s.store.BatchInsertEvents(ctx, batch); err != nil {
		slog.Error("failed to batch insert events", slog.Any("error", err))
		return err
	}

	// 2. Upsert Daily Analytics
	if err := s.store.UpsertDailyStats(ctx, dailyStats); err != nil {
		slog.Error("failed to upsert daily stats", slog.Any("error", err))
		return err
	}

	return nil
}

func (s *AnalyticsService) GetLinkAnalytics(ctx context.Context, hash string, from, to time.Time) (*entity.AnalyticsReport, error) {
	return s.store.Query(ctx, hash, from, to)
}
