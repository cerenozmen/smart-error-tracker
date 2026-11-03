# Smart Error Tracker (SET)

Uygulamalardan gelen hataları HTTP API üzerinden toplayan, aynı hataları tek kayıt altında gruplayan ve Cohere ile yapay zeka destekli analiz raporu üreten bir hata takip sistemi.

Go (Fiber), PostgreSQL ve Docker ile geliştirilmiştir.

## Özellikler

- **Hata toplama:** Uygulamalar hatalarını proje API anahtarıyla `POST /api/v1/errors` adresine gönderir.
- **Hata gruplama:** Hata mesajı ve stack trace'ten MD5 parmak izi üretilir. Aynı parmak izine sahip hatalar tek bir `issue` altında toplanır, her tekrar ise `errors` tablosuna ayrı kaydedilir.
- **Yapay zeka analizi:** Hata bir Goroutine içinde Cohere'e gönderilir, böylece API yanıtı analizi beklemez. Hatanın nedenini ve çözüm önerisini içeren Türkçe rapor veritabanına kaydedilir.
- **Model yedekleme:** Bir model yanıt vermezse sırayla `command-r7b`, `command-light` ve `command-nightly` denenir.
- **Web paneli:** Kayıtlı hatalar, stack trace'ler ve analiz raporları `/dashboard` adresinden görüntülenir.

## Nasıl Çalışır?

```mermaid
flowchart LR
    A[İstemci uygulama] -->|POST /api/v1/errors| B[Fiber API]
    B -->|API anahtarı kontrolü| D[(PostgreSQL)]
    B -->|Issue ve hata kaydı| D
    B -->|201 Created| A
    B -.->|Goroutine| C[Cohere API]
    C -.->|Analiz raporu| D
    E[Web paneli] -->|GET /dashboard| B
```

1. API, isteği gönderen projeyi `api_key` ile doğrular.
2. `message + stack_trace` değerinin MD5 özeti parmak izi olarak hesaplanır.
3. `issues` tablosunda aynı proje ve parmak izine sahip kayıt varsa `last_seen_at` güncellenir, yoksa yeni kayıt açılır.
4. Hatanın kendisi `errors` tablosuna eklenir ve istemciye `201` döndürülür.
5. Arka planda Cohere'den alınan rapor, ilgili kaydın `ai_analysis` alanına yazılır.

## Teknolojiler

| Katman | Teknoloji |
| --- | --- |
| Backend | Go 1.25, Fiber v2 |
| Veritabanı | PostgreSQL 15, pgx v5 |
| Yapay zeka | Cohere API (cohere-go v2) |
| Arayüz | Go `html/template`, Tailwind CSS, Alpine.js |
| Container | Docker |

## Kurulum

**Gereksinimler:** Docker ve bir [Cohere API anahtarı](https://dashboard.cohere.com/api-keys).

Aşağıdaki komutlar Bash içindir. PowerShell kullanıyorsanız satır sonlarındaki `\` karakterini `` ` `` ile değiştirin.

### 1. Projeyi indirin

```bash
git clone <repo-url>
cd smart-error-tracker
```

### 2. Docker ağını ve veritabanını başlatın

```bash
docker network create set_net

docker run -d \
  --name set_postgres \
  --network set_net \
  -p 5434:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=123456 \
  -e POSTGRES_DB=smart_error_tracker \
  -v set_pgdata:/var/lib/postgresql/data \
  postgres:15-alpine
```

> `123456` yalnızca yerel geliştirme için örnek bir paroladır.

### 3. Veritabanı şemasını oluşturun

Veritabanı konsolunu açın:

```bash
docker exec -it set_postgres psql -U postgres -d smart_error_tracker
```

Aşağıdaki SQL'i yapıştırın, ardından `\q` ile çıkın:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_key UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

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

CREATE TABLE IF NOT EXISTS errors (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER REFERENCES issues(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    stack_trace TEXT,
    environment VARCHAR(100),
    payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_issues_project_fingerprint ON issues(project_id, fingerprint);
CREATE INDEX IF NOT EXISTS idx_errors_issue_id ON errors(issue_id);
```

### 4. Uygulamayı derleyip çalıştırın

```bash
docker build -t smart-error-tracker -f dockerfile .

docker run -d \
  --name set_app \
  --network set_net \
  -p 8080:8080 \
  -e DB_CONN_STR="postgres://postgres:123456@set_postgres:5432/smart_error_tracker?sslmode=disable" \
  -e COHERE_API_KEY="COHERE_API_ANAHTARINIZ" \
  smart-error-tracker
```

Sunucunun çalıştığını doğrulayın:

```bash
curl http://localhost:8080/ping
```

### Docker olmadan çalıştırma

Go 1.25 veya üzeri gerekir. 2. ve 3. adımlardaki veritabanı çalışırken proje dizininde şu komutları çalıştırın:

```bash
export COHERE_API_KEY="COHERE_API_ANAHTARINIZ"
go run .
```

## Ortam Değişkenleri

| Değişken | Zorunlu | Açıklama |
| --- | --- | --- |
| `COHERE_API_KEY` | Analiz için evet | Tanımlı değilse hatalar yine kaydedilir, ancak analiz raporu üretilmez. |
| `DB_CONN_STR` | Hayır | PostgreSQL bağlantı adresi. Varsayılan: `postgres://postgres:123456@127.0.0.1:5434/smart_error_tracker?sslmode=disable` |

## API Kullanımı

| Metot | Adres | Açıklama |
| --- | --- | --- |
| `GET` | `/ping` | Sunucu durumunu kontrol eder |
| `POST` | `/api/v1/projects` | Yeni proje ve API anahtarı oluşturur |
| `POST` | `/api/v1/errors` | Hata kaydeder ve analizi başlatır |
| `GET` | `/dashboard` | Web panelini açar |

`POST` isteklerinde `Content-Type: application/json` başlığı zorunludur.

### Proje oluşturma

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "Demo Projesi"}'
```

Yanıt (`201 Created`):

```json
{
  "api_key": "fc8dc95a-c7a9-4899-a24b-89acb052a6f1",
  "message": "Proje başarıyla oluşturuldu! 🚀",
  "name": "Demo Projesi"
}
```

### Hata gönderme

Yanıttaki `api_key` değerini kullanın:

```bash
curl -X POST http://localhost:8080/api/v1/errors \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "PROJE_API_ANAHTARI",
    "message": "database connection timeout",
    "stack_trace": "main.go:55 -> database.ConnectDB() -> timeout after 5s",
    "environment": "production",
    "payload": {"user_id": 42}
  }'
```

`environment` ve `payload` alanları isteğe bağlıdır.

Yanıt (`201 Created`):

```json
{
  "fingerprint": "6695c7ef9ad1140a9ce53655ffc8043c",
  "issue_id": 1,
  "message": "Hata başarıyla kaydedildi! "
}
```

| Durum kodu | Anlamı |
| --- | --- |
| `400` | İstek gövdesi geçersiz, `Content-Type` başlığı eksik veya proje adı boş |
| `401` | API anahtarı geçersiz |
| `500` | Veritabanı işlemi başarısız |

### Sonuçları izleme

[http://localhost:8080/dashboard](http://localhost:8080/dashboard) adresini açın. Analiz arka planda üretildiği için raporu görmek üzere sayfayı yenilemeniz gerekebilir.

## Proje Yapısı

```text
.
├── main.go                     # Sunucu, rotalar ve web paneli
├── database/db.go              # PostgreSQL bağlantı havuzu
├── handlers/
│   ├── error_handler.go        # Hata kaydı, gruplama, analiz başlatma
│   └── project.go              # Proje oluşturma
├── services/
│   ├── ai_service.go           # Cohere entegrasyonu
│   └── project_service.go      # Proje veritabanı işlemleri
├── models/models.go            # Veri modelleri
├── templates/dashboard.html    # Web paneli arayüzü
├── dockerfile                  # Çok aşamalı Docker derlemesi
└── vendor/                     # Go bağımlılıkları
```

## Bilinen Sınırlamalar

- **Kimlik doğrulama yok:** `/dashboard` ve `/api/v1/projects` herkese açıktır ve panel tüm projelerin hatalarını gösterir. Bu nedenle uygulama bu haliyle internete açılmamalıdır.
- **Tekrarlanan analiz:** Aynı hatanın her tekrarı yeni bir Cohere çağrısı yapar ve önceki raporun üzerine yazar.
- **Hassas gruplama:** Parmak izi metnin tamamından üretildiği için satır numarası gibi küçük bir değişiklik yeni bir `issue` açar.
- **Analiz hataları panele yansımaz:** Analiz başarısız olursa hata yalnızca sunucu günlüğüne yazılır ve panelde "hazırlanıyor" mesajı görünmeye devam eder.
- **Dile bağlı öneri:** Rapordaki örnek kod, hatanın geldiği dilden bağımsız olarak Go ile istenir.
- **Sınırlı panel:** Sayfa kendini otomatik yenilemez, Markdown raporları düz metin olarak gösterilir, hata tekrar sayısı görüntülenmez. Veritabanı ve yapay zeka durum kartları sabit değerler gösterir.
- **Durum yönetimi yok:** `status` alanı kaydedilir, ancak bir hatayı çözüldü olarak işaretleyecek bir uç nokta yoktur.
- **Manuel şema:** Veritabanı şeması elle uygulanır, migration aracı kullanılmaz.

## Geliştirme Planı

- [ ] Panel ve proje oluşturma için kimlik doğrulama
- [ ] Analizin her `issue` için bir kez çalışması
- [ ] Panelde hata tekrar sayısı, filtreleme ve durum yönetimi
- [ ] Stack trace normalizasyonu ile daha iyi gruplama
- [ ] Docker Compose ve otomatik veritabanı migration'ları
- [ ] Birim ve entegrasyon testleri
