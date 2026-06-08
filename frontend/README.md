# Logarift Frontend

React + Vite frontend for local friction logging and dashboarding.

The default logging UI is intentionally small:

```text
1. When?
2. Friction level: green / yellow / orange / red
3. Rich notes with links and screenshots
```

The frontend posts these fields to:

```text
POST /api/v1/friction-events/quick
```

The backend then uses deterministic local enrichment rules to infer workflow stage, friction layer, friction type, time loss, resume time, interruptions, and tags. When asynchronous LLM enrichment is queued, the frontend opens `GET /api/v1/enrichment-jobs/{id}/events` as a Server-Sent Events stream and falls back to polling if EventSource is unavailable.

Run directly:

```bash
cd frontend
npm install
npm run dev
```

The frontend expects the backend at `http://localhost:8080` by default.
Override with:

```bash
VITE_LOGARIFT_API_BASE_URL=http://localhost:8080 npm run dev
```

## Current UI Structure

The frontend is split into two main tabs:

- **Log**: quick friction composer and recent logs only.
- **Dashboard**: metrics, score cards, and breakdowns.

Optional goals and sessions are hidden behind the **Optional context** modal so they do not slow down the default logging flow.

## Rich Notes

The notes field supports:

- formatted text
- links
- image upload through the Screenshot button
- pasted screenshots from clipboard
- image drag and drop

Screenshots are uploaded to the backend and inserted into notes as local image URLs.

## Tooltips

Metrics and complex fields include tooltip affordances. This is required for fields such as avg cognitive load, avg inference confidence, CLA, FCI, and SDC.
