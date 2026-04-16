CREATE TABLE IF NOT EXISTS click_events (
    id           BIGSERIAL       NOT NULL,
    link_hash    VARCHAR(10)     NOT NULL,                -- FK lógica (sem FK física por performance)
    ip_address   INET            NOT NULL,                -- IPv4/IPv6 nativo do PostgreSQL
    user_agent   TEXT,
    referer      TEXT,
    country      VARCHAR(2),                               -- ISO 3166-1 alpha-2 (Worker resolve via GeoIP)
    device_type  VARCHAR(10),                               -- desktop | mobile | tablet | bot
    browser      VARCHAR(50),
    os           VARCHAR(50),
    clicked_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, clicked_at)                            -- PK composta para suportar particionamento
) PARTITION BY RANGE (clicked_at);

-- Partição do mês corrente e do próximo (Worker pode criar novas via SQL)
CREATE TABLE click_events_2026_04 PARTITION OF click_events
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE click_events_2026_05 PARTITION OF click_events
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE click_events_2026_06 PARTITION OF click_events
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

-- Agregações por link (a query mais comum do analytics)
CREATE INDEX idx_click_events_hash_time ON click_events (link_hash, clicked_at DESC);

-- Queries geográficas
CREATE INDEX idx_click_events_country ON click_events (country, clicked_at DESC)
    WHERE country IS NOT NULL;
