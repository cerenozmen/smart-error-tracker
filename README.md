Smart Error Tracker (SET)
Smart Error Tracker (SET), Go (Fiber) ve PostgreSQL kullanılarak geliştirilmiş, uygulamalardaki hataları gerçek zamanlı olarak yakalayan, gruplayan ve yapay zeka (Gemini AI) desteğiyle analiz eden modern bir hata takip sistemidir.
Özellikler
Gerçek Zamanlı Hata Toplama: Uygulamalarınızdan gelen hataları HTTP API üzerinden anında yakalar.
Hata Gruplama (Fingerprinting): Benzersiz hata mesajı ve Stack Trace karmasını (MD5) kullanarak aynı hataları tek bir çatı (Issue) altında toplar, veritabanında kirliliği önler.
Yapay Zeka Analiz Raporu: Yakalanan hatalar, asenkron olarak (Goroutine)  Cohore servisine gönderilir. Yapay zeka hatanın kaynağını analiz ederek çözüm önerilerini doğrudan veritabanına kaydeder.
Kullanıcı Dostu Arayüz (Dashboard): Kaydedilen tüm hataları, hata sıklığını ve yapay zeka analiz raporlarını pürüzsüz bir web paneli üzerinden sunar.
Docker Uyumlu: Kolay ve hızlı kurulum için veritabanı ve uygulama katmanları Dockerize edilmiştir.
Teknolojiler
Backend: Go (Golang) & Fiber Web Framework
Veritabanı: PostgreSQL
Yapay Zeka: Cohere API
Containerization: Docker 
Kurulum ve Çalıştırma
1. Veritabanının Hazırlanması (PostgreSQL)
Docker üzerinde kalıcı ve güvenli bir PostgreSQL veritabanını başlatmak için terminalden şu komutu çalıştırın:
docker run -d \
  --name set_postgres \
  -p 5434:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=123456 \
  -e POSTGRES_DB=smart_error_tracker \
  -v set_pgdata:/var/lib/postgresql/data \
  postgres:15-alpine


Veritabanı Şemasının Oluşturulması
pgAdmin veya psql kullanarak smart_error_tracker veritabanına bağlanın ve aşağıdaki şema sorgusunu çalıştırın:
-- UUID eklentisini aktifleştirme
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projeler Tablosu
CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_key UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Hata Grupları (Issues) Tablosu
CREATE TABLE IF NOT EXISTS issues (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    fingerprint VARCHAR(64) NOT NULL,
    title TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'unresolved',
    ai_analysis TEXT,
    ai_analyzed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_project_fingerprint UNIQUE (project_id, fingerprint)
);

-- Tekil Hatalar (Errors) Tablosu
CREATE TABLE IF NOT EXISTS errors (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER REFERENCES issues(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    stack_trace TEXT,
    environment VARCHAR(100),
    payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- İndeksler
CREATE INDEX IF NOT EXISTS idx_issues_project_fingerprint ON issues(project_id, fingerprint);
CREATE INDEX IF NOT EXISTS idx_errors_issue_id ON errors(issue_id);


Uygulamanın Docker ile Çalıştırılması
Projenin ana dizininde uygulamanızı derleyip (Image oluşturup) aşağıdaki komutla API anahtarlarınızı tanımlayarak çalıştırın:
docker run -d \
  --name SET \
  -p 8080:8080 \
  -e DB_CONN_STR="postgres://postgres:123456@host.docker.internal:5434/smart_error_tracker?sslmode=disable" \
  -e GEMINI_API_KEY="SENIN_GEMINI_API_KEYIN" \
  smart-error-tracker


API Kullanımı (Postman Adımları)
1. Yeni Bir Proje Oluşturma
Metot: POST
Headers: Key:Content-Type/Value:application/json
URL: http://localhost:8080/api/v1/projects
Gövde (JSON):
{
  "name": "Benim Ilk Go Projem"
}


Yanıt olarak gelen api_key değerini kopyalayın.
2. Yeni Bir Hata Fırlatma
Metot: POST
URL: http://localhost:8080/api/v1/errors
Gövde (JSON):
{
  "api_key": "PROJE_OLUSTURURKEN_ALINAN_API_KEY",
  "message": "database connection timeout",
  "stack_trace": "main.go:55 -> database.ConnectDB() -> timeout after 5s",
  "environment": "production"
}


İzleme
Gönderdiğiniz hataları ve arka planda üretilen yapay zeka analiz raporlarını tarayıcınızdan şu adrese giderek izleyebilirsiniz:
👉 http://localhost:8080/dashboard
