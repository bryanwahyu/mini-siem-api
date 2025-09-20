#!/usr/bin/env bash
set -Eeuo pipefail

BIN=${BIN:-server-analyst}
OUT=${OUT:-/tmp/sa_samples.log}
HOST=${HOST:-$(hostname -s)}

cat > "$OUT" <<'EOF'
192.0.2.10 - - [10/Sep/2024:12:00:00 +0000] "GET /?q=union select 1,2 from users HTTP/1.1" 200 123 "-" "Mozilla/5.0"
198.51.100.23 - - [10/Sep/2024:12:00:01 +0000] "GET /login HTTP/1.1" 401 0 "-" "curl/8.0"
203.0.113.55 - - [10/Sep/2024:12:00:02 +0000] "GET /index.php?param=<script>alert(1)</script> HTTP/1.1" 200 321 "-" "Mozilla/5.0"
198.51.100.23 - - [10/Sep/2024:12:00:03 +0000] "GET /wp-admin HTTP/1.1" 404 0 "-" "wpscan"
198.51.100.23 - - [10/Sep/2024:12:00:04 +0000] "GET /?q=../../etc/passwd HTTP/1.1" 200 0 "-" "Mozilla/5.0"
198.51.100.50 - - [10/Sep/2024:12:00:05 +0000] "GET /?ref=slot88-gacor-maxwin HTTP/1.1" 200 0 "http://slot88.example" "SpamBot/1.0"
EOF

# Flood lines
for i in $(seq 1 150); do echo "198.51.100.23 - - [10/Sep/2024:12:00:10 +0000] \"GET / HTTP/1.1\" 200 0 \"-\" \"ab/2.3\"" >> "$OUT"; done

echo "Replaying samples..."
$BIN replay -file "$OUT" -source sample -host "$HOST"

echo "Exporting daily events and detections..."
$BIN export -host "$HOST" -prefix prod -type events
$BIN export -prefix prod -type detections

echo "Done."
