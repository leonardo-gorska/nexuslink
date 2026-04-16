CREATE TABLE IF NOT EXISTS daily_analytics (
    id           BIGSERIAL       PRIMARY KEY,
    link_hash    VARCHAR(10)     NOT NULL,
    event_date   DATE            NOT NULL,
    country      VARCHAR(2),                               -- NULL = total global
    device_type  VARCHAR(10),                               -- NULL = total global
    browser      VARCHAR(50),                               -- NULL = total global
    click_count  INTEGER         NOT NULL DEFAULT 0,
    unique_ips   INTEGER         NOT NULL DEFAULT 0,       -- Visitantes únicos estimados
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

);

CREATE UNIQUE INDEX uq_daily_dimensions ON daily_analytics (link_hash, event_date, COALESCE(country, ''), COALESCE(device_type, ''), COALESCE(browser, ''));

-- Dashboard por link
CREATE INDEX idx_daily_link_date ON daily_analytics (link_hash, event_date DESC);

-- Ranking global (top links do dia)
CREATE INDEX idx_daily_top_clicks ON daily_analytics (event_date DESC, click_count DESC)
    WHERE country IS NULL AND device_type IS NULL AND browser IS NULL;
