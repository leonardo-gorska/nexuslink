CREATE TABLE IF NOT EXISTS links (
    id           BIGSERIAL       PRIMARY KEY,
    hash         VARCHAR(10)     NOT NULL,
    original_url TEXT            NOT NULL,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ,                             -- NULL = nunca expira
    click_count  BIGINT          NOT NULL DEFAULT 0,      -- Counter denormalizado (leitura rápida)
    is_active    BOOLEAN         NOT NULL DEFAULT TRUE,   -- Soft delete

    CONSTRAINT chk_hash_length CHECK (char_length(hash) BETWEEN 5 AND 10),
    CONSTRAINT chk_url_not_empty CHECK (char_length(original_url) > 0)
);

-- Lookup primário (a query mais executada do sistema inteiro)
CREATE UNIQUE INDEX idx_links_hash_active ON links (hash) WHERE is_active = TRUE;

-- Limpeza de links expirados (cron job ou goroutine periódica)
CREATE INDEX idx_links_expires_at ON links (expires_at)
    WHERE expires_at IS NOT NULL AND is_active = TRUE;

-- Listagem admin ordenada
CREATE INDEX idx_links_created_at ON links (created_at DESC);
