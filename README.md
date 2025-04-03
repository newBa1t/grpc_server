Для проверки приложения оставил local.env тут

# General application configuration
LOG_LEVEL=info

# gRPC server
GRPC_PORT=:50052
GRPC_WRITE_TIMEOUT=15s

# PostgreSQL configuration
DB_HOST=localhost
DB_PORT=5431
DB_NAME=grpc_servis
DB_USER=admin
DB_PASSWORD=admin
DB_SSL_MODE=disable
DB_POOL_MAX_CONNS=10
DB_POOL_MAX_CONN_LIFETIME=300s
DB_POOL_MAX_CONN_IDLE_TIME=15