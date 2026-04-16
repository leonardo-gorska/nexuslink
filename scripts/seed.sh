#!/bin/bash
# scripts/seed.sh
# Cria diversos links aleatórios contra a API NexusLink para base inicial de testes

API_URL="http://localhost:8080/api/v1/links"

echo "🧪 [Seed] Inicializando criação das entidades de teste..."
echo "--------------------------------------------------------"

urls=(
    "https://github.com/leonardo-gorska"
    "https://linkedin.com/in/leonardogorska"
    "https://google.com"
    "https://go.dev"
    "https://pkg.go.dev/github.com/go-chi/chi"
    "https://rabbitmq.com"
    "https://redis.io"
    "https://postgresql.org"
    "https://kubernetes.io"
    "https://aws.amazon.com"
)

for URL in "${urls[@]}"; do
    echo "[+] Criando hash curto para: $URL"
    curl -s -X POST "$API_URL" \
         -H "Content-Type: application/json" \
         -d "{\"url\": \"$URL\", \"expires_in\": \"720h\"}" | grep -oE '"short_url":"[^"]+"' || true
    echo ""
done

echo "✅ [Seed] Finalizado com sucesso. Seed inject complete."
