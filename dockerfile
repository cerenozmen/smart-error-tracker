# 1. Aşama: Uygulamayı derleme aşaması 
FROM golang:1.25-alpine AS builder

# Git ve derleme için gerekli araçları 
RUN apk add --no-cache git

WORKDIR /app

# Go modülleri
COPY go.mod go.sum ./

# Güvenli ve hızlı indirme için Go Proxy tanımlıyoruz
ENV GOPROXY=https://proxy.golang.org,direct

# Bağımlılıkları indiriyoruz
RUN go mod download

# Tüm proje dosyalarını kopyalıyoruz
COPY . .

# Uygulamayı optimize edilmiş bir binary olarak derliyoruz
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# 2. Aşama: Çalıştırma aşaması 
FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Derlenen binary dosyasını ve templates klasörünü kopyalıyoruz
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates

# Uygulamanın çalışacağı portu belirtiyoruz
EXPOSE 8080

# Uygulamayı başlatıyoruz
CMD ["./main"]