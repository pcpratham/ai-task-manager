# AI Task Manager

An intelligent task management application powered by AI. Create, organize, and manage tasks with the help of Google Gemini AI for smart task generation, breakdown, priority suggestions, and time estimates.

## Features

- **User Authentication**: Secure registration and login with JWT tokens
- **Task Management**: Full CRUD operations for tasks with categories, priorities, and due dates
- **AI-Powered Features**:
  - **Natural Language Task Generation**: Describe what you need and AI creates organized tasks
  - **Task Breakdown**: Automatically break down complex tasks into subtasks
  - **Priority Suggestions**: Get AI recommendations for task priorities
  - **Time Estimation**: AI-powered time estimates for tasks
- **Filtering & Search**: Filter tasks by status, priority, category, and search terms
- **Responsive Design**: Works great on desktop and mobile

## Tech Stack

- **Backend**: Go 1.22+ with Gin framework
- **Frontend**: Next.js 14 with App Router, TypeScript, Tailwind CSS, shadcn/ui
- **Database**: MySQL 8.0 (via Docker for local dev)
- **AI**: Google Gemini API (free tier available!)

## Prerequisites

- Go 1.22+
- Node.js 18+
- Docker & Docker Compose
- Google Gemini API Key (free)

## Get Your Free Gemini API Key

1. Go to https://aistudio.google.com/apikey
2. Sign in with your Google account
3. Click **"Create API Key"**
4. Copy the key

**Free limits**: 60 requests/minute, 1500 requests/day - more than enough!

## Quick Start

### 1. Clone the repository

```bash
git clone <repo-url>
cd ai-task-manager
```

### 2. Start MySQL with Docker

```bash
docker compose up -d
```

Wait for MySQL to be ready:
```bash
docker compose logs mysql
```

### 3. Set up the backend

```bash
cd backend

# Copy environment file
cp .env.example .env

# Edit .env and add your Gemini API key
# GEMINI_API_KEY=your-key-here

# Run the backend
go run ./cmd/server
```

The backend will start on http://localhost:8080

### 4. Set up the frontend

```bash
cd frontend

# Copy environment file
cp .env.example .env.local

# Install dependencies
npm install

# Run the frontend
npm run dev
```

The frontend will start on http://localhost:3000

### 5. Open the app

Visit http://localhost:3000 in your browser, register an account, and start managing tasks!

## Environment Variables

### Backend (.env)

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_DSN` | MySQL connection string | `taskuser:taskpassword@tcp(localhost:3306)/taskmanager?parseTime=true` |
| `JWT_SECRET` | Secret for JWT signing | Required |
| `GEMINI_API_KEY` | Google Gemini API key | Required for AI features |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` |

### Frontend (.env.local)

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Backend API URL | `http://localhost:8080` |

## Project Structure

```
ai-task-manager/
├── docker-compose.yml       # MySQL container setup
├── README.md
│
├── backend/
│   ├── cmd/server/main.go   # Entry point
│   ├── internal/
│   │   ├── config/          # Environment configuration
│   │   ├── handler/         # HTTP handlers
│   │   ├── middleware/      # Auth & CORS middleware
│   │   ├── model/           # Data models
│   │   ├── repository/      # Database operations
│   │   └── service/         # Business logic & AI
│   ├── migrations/          # SQL migrations
│   ├── Dockerfile
│   └── go.mod
│
└── frontend/
    ├── src/
    │   ├── app/             # Next.js pages
    │   ├── components/      # React components
    │   ├── context/         # Auth context
    │   └── lib/             # API client & types
    └── package.json
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `GET /api/auth/me` - Get current user

### Tasks
- `GET /api/tasks` - List tasks (with filters)
- `POST /api/tasks` - Create task
- `GET /api/tasks/:id` - Get task with subtasks
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task
- `GET /api/tasks/categories` - Get all categories

### AI Features
- `POST /api/ai/generate` - Generate tasks from natural language
- `POST /api/ai/breakdown/:id` - Break down task into subtasks
- `POST /api/ai/suggest-priority` - Get priority suggestion
- `POST /api/ai/estimate-time` - Get time estimate

## Docker Commands

```bash
# Start MySQL
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs mysql

# Stop MySQL
docker compose down

# Reset database (delete all data)
docker compose down -v && docker compose up -d

# Connect to MySQL CLI
docker exec -it taskmanager-mysql mysql -u taskuser -ptaskpassword taskmanager
```

## Database Connection (GUI Tools)

Connect with DBeaver, TablePlus, or any MySQL client:

| Field | Value |
|-------|-------|
| Host | `localhost` |
| Port | `3306` |
| Database | `taskmanager` |
| Username | `taskuser` |
| Password | `taskpassword` |

## Deployment

### Railway (Backend + MySQL)

1. Create a Railway project
2. Add MySQL service
3. Deploy backend from GitHub (auto-detects Dockerfile)
4. Set environment variables:
   - `DB_DSN` from MySQL service
   - `JWT_SECRET` (generate secure random string)
   - `GEMINI_API_KEY`
   - `ALLOWED_ORIGINS` (your Vercel URL)

### Vercel (Frontend)

1. Import repository to Vercel
2. Set root directory to `frontend`
3. Set environment variable:
   - `NEXT_PUBLIC_API_URL` (your Railway URL)

## License

MIT
