# Nexus-Dashboard

A real-time data monitoring and visualization platform built with Go, Kafka, TimescaleDB, and React. Nexus-Dashboard enables high-performance ingestion and analysis of time-series data with an intuitive web-based interface.

## 🎯 Features

- **Real-time Data Ingestion**: Kafka-based event streaming for high-throughput data collection
- **Time-Series Database**: TimescaleDB (PostgreSQL extension) for efficient time-series storage
- **Scalable Architecture**: Decoupled API and ingestor services
- **Modern Dashboard**: React-based frontend with Tailwind CSS styling
- **Docker Compose**: Complete infrastructure setup in a single command

## 🏗️ Architecture

### Core Components

```
Nexus-Dashboard/
├── backend/
│   ├── cmd/
│   │   ├── api/           # API server service
│   │   └── ingestor/      # Data ingestion service
│   ├── internal/          # Business logic and handlers
│   ├── pkg/              # Reusable packages
│   ├── migrations/       # Database schema migrations
│   └── go.mod           # Go dependencies
├── frontend/
│   ├── src/             # React components and pages
│   └── package.json    # Node.js dependencies
└── docker-compose.yml  # Infrastructure services
```

### Infrastructure Stack

- **Database**: TimescaleDB 2 (PostgreSQL 16 with time-series extensions)
- **Message Broker**: Apache Kafka 7.6
- **Coordination**: Zookeeper 7.6

## 🚀 Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.18+ (for backend development)
- Node.js 16+ (for frontend development)
- Make

### Quick Start

1. **Start Infrastructure**
   ```bash
   make infra-up
   ```
   This launches:
   - TimescaleDB on port 5432
   - Kafka on port 9092
   - Zookeeper

2. **Build & Run Backend Services**
   ```bash
   # Terminal 1 - API Server
   make api

   # Terminal 2 - Data Ingestor
   make ingestor
   ```

3. **Start Frontend Development Server**
   ```bash
   make frontend-dev
   ```
   The dashboard will be available at `http://localhost:5173`

### Complete Setup

To set everything up at once:
```bash
make run
```

Then start the API and ingestor in separate terminals as shown above.

## 📋 Available Commands

```makefile
# Infrastructure Management
make infra-up          # Start Docker containers
make infra-down        # Stop Docker containers

# Backend Services
make build-api         # Build API binary
make build-ingestor    # Build ingestor binary
make api              # Build and run API server
make ingestor         # Build and run ingestor service

# Frontend
make frontend-dev     # Start Vite development server
make frontend-build   # Build for production

# Complete Setup
make run              # Build all services (start infrastructure + binaries)
```

## 🔧 Configuration

### Backend Configuration

Configure backend services via environment variables:

```bash
cp backend/.env.example backend/.env
# Edit backend/.env with your settings
```

### Database Connection

The default configuration:
- **Host**: localhost (timescaledb)
- **Port**: 5432
- **Database**: nexus
- **User**: nexus
- **Password**: nexus_secret

## 📊 Database Schema

Database migrations are automatically applied on container startup via:
- `backend/migrations/001_init.sql`

Tables are initialized with TimescaleDB hypertables for optimized time-series queries.

## 🎨 Frontend Stack

- **Framework**: React with Vite
- **Styling**: Tailwind CSS
- **Build Tool**: Vite

### Frontend Development

```bash
cd frontend
npm install
npm run dev      # Start development server
npm run build    # Build for production
```

## 📦 Backend Stack

- **Language**: Go
- **Database Driver**: PostgreSQL/TimescaleDB
- **Message Queue**: Kafka consumer

### Backend Services

**API Server** (`cmd/api`)
- Serves HTTP endpoints
- Handles client requests
- Queries time-series data

**Ingestor** (`cmd/ingestor`)
- Consumes from Kafka topics
- Writes data to TimescaleDB
- Processes real-time events

## 🐳 Docker Services

All services are defined in `docker-compose.yml`:

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| timescaledb | timescale/timescaledb:2-pg16 | 5432 | Time-series database |
| kafka | confluentinc/cp-kafka:7.6 | 9092 | Event streaming |
| zookeeper | confluentinc/cp-zookeeper:7.6 | 2181 | Kafka coordination |

### Health Checks

All services include health checks to ensure readiness:
- Database: PostgreSQL ready check
- Kafka: Topic list verification
- Zookeeper: Socket connectivity test

## 📝 Development Workflow

1. **Start Infrastructure**
   ```bash
   make infra-up
   ```

2. **Make Code Changes**
   - Backend: Edit files in `backend/`
   - Frontend: Edit files in `frontend/src/`

3. **Rebuild & Run**
   ```bash
   make api          # Terminal 1
   make ingestor     # Terminal 2
   make frontend-dev # Terminal 3
   ```

4. **Stop Everything**
   ```bash
   make infra-down
   ```

## 🛠️ Troubleshooting

### Container Issues

```bash
# View logs
docker compose logs timescaledb
docker compose logs kafka
docker compose logs zookeeper

# Restart services
docker compose restart
```

### Database Connection Errors

Ensure TimescaleDB is healthy:
```bash
docker compose ps
# Look for "healthy" status
```

### Port Conflicts

If ports are in use, modify `docker-compose.yml`:
```yaml
ports:
  - '5433:5432'  # Map to different host port
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

For issues and questions, please open a GitHub issue in this repository.

---

**Built with ❤️ for real-time data monitoring**
