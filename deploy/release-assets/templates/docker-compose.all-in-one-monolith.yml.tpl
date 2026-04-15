services:
  omnishard:
    image: __IMAGE_NAMESPACE__/omnishard-all-in-one-monolith:__OMNISHARD_TAG__
    restart: unless-stopped
    environment:
      NEXT_PUBLIC_API_URL: "http://localhost:8080"
      API_INTERNAL_URL: "http://127.0.0.1:8080"
      GATEWAY_URL: "http://127.0.0.1:8080"
      NEXT_PUBLIC_BETTER_AUTH_URL: "http://localhost:3000"
      Omnishard_DB_PATH: "/app/data/Omnishard.db"
      Omnishard_SHARDMAP_DB_PATH: "/app/data/Omnishard-shardmap.db"
    volumes:
      - omnishard_data:/app/data
    ports:
      - "3000:3000"
      - "8080:8080"

volumes:
  omnishard_data: