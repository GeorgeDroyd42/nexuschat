docker network create auth-network 2>/dev/null || true

docker rm -f auth_postgres
sudo docker run --name auth_postgres \
  --network auth-network \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=authdb \
  -p 5432:5432 \
  -v "$(pwd)/data:/var/lib/postgresql/data" \
  -d postgres:latest
sudo docker run --name auth_redis \
  --network auth-network \
  -p 6379:6379 \
  -d redis:alpine  
