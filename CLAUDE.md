# CLAUDE.md

## Build and Run Commands

### Backend (Go / Gin)
- **Run development server**: `go run main.go` (run from the `backend` directory)
- **Build application**: `go build -o server main.go` (run from the `backend` directory)
- **Run tests**: `go test ./...` (run from the `backend` directory)

### Frontend (Vue 3 / Vite)
- **Install dependencies**: `npm install` (run from the `frontend` directory)
- **Run development server**: `npm run dev` (run from the `frontend` directory)
- **Build production bundle**: `npm run build` (run from the `frontend` directory)

## Agent Skills

### Issue Tracker
Issues and PRDs for this repo live as GitHub issues. See `docs/agents/issue-tracker.md` ([issue-tracker.md](file:///home/oyon/sehuacho/veterinaria/docs/agents/issue-tracker.md)).

### Triage Labels
We use the standard label vocabulary (needs-triage, needs-info, ready-for-agent, ready-for-human, wontfix) for classifying issues. See `docs/agents/triage-labels.md` ([triage-labels.md](file:///home/oyon/sehuacho/veterinaria/docs/agents/triage-labels.md)).

### Domain Docs
This project uses a single-context domain layout. See `docs/agents/domain.md` ([domain.md](file:///home/oyon/sehuacho/veterinaria/docs/agents/domain.md)).
