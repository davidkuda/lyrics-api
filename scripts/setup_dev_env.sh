# run "source ./setup_test_env.sh" to export env var for the session
export DB_ADDR="localhost:5432"
export DB_NAME="lyricsapi"
export DB_USER="lyricsapi"
export DB_PASSWORD="lyricsapi"

# jwt
export JWT_SECRET="verysecret"
export JWT_ISSUER="kuda.ai"
export JWT_AUDIENCE="kuda.ai"
export COOKIE_DOMAIN="localhost"

# CORS
export ALLOWED_CORS_ORIGINS="http://localhost:3000 http://localhost:3001"
