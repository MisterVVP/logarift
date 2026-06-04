import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

const API_BASE = import.meta.env.VITE_LOGARIFT_API_BASE_URL || 'http://localhost:8080';

const workflowStages = ['planning', 'local_development', 'build', 'test', 'code_review', 'merge', 'deployment', 'operation', 'debugging', 'documentation', 'coordination', 'learning'];
const frictionLayers = ['technical', 'temporal', 'cognitive', 'social_process', 'motivational', 'environmental'];
const frictionTypes = ['slow_feedback', 'failed_feedback', 'unclear_error', 'missing_documentation', 'ambiguous_ownership', 'interruption', 'waiting_for_review', 'waiting_for_ci', 'context_switch', 'rework', 'tooling_failure', 'environment_setup', 'coordination_overhead', 'decision_blocked'];
const goalStatuses = ['active', 'completed', 'deferred', 'abandoned'];

function nowLocalInput() {
  const date = new Date();
  date.setMinutes(date.getMinutes() - date.getTimezoneOffset());
  return date.toISOString().slice(0, 16);
}

function toApiTime(local) {
  return new Date(local).toISOString();
}

async function api(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });
  if (response.status === 204) return null;
  const text = await response.text();
  const data = text ? JSON.parse(text) : null;
  if (!response.ok) {
    const message = data?.error?.fields?.map((f) => `${f.field}: ${f.message}`).join('; ') || data?.error?.message || `HTTP ${response.status}`;
    throw new Error(message);
  }
  return data;
}

function Select({ label, value, onChange, values }) {
  return (
    <label>
      <span>{label}</span>
      <select value={value} onChange={(e) => onChange(e.target.value)}>
        {values.map((v) => <option key={v} value={v}>{v}</option>)}
      </select>
    </label>
  );
}

function NumberInput({ label, value, min, max, onChange }) {
  return (
    <label>
      <span>{label}</span>
      <input type="number" min={min} max={max} value={value} onChange={(e) => onChange(Number(e.target.value))} />
    </label>
  );
}

function EventForm({ goals, sessions, onCreated }) {
  const [form, setForm] = useState({
    timestamp_start: nowLocalInput(),
    workflow_stage: 'test',
    friction_layer: 'technical',
    friction_type: 'failed_feedback',
    severity_self: 3,
    cognitive_load_self: 3,
    emotion_valence: -1,
    time_lost_minutes: 10,
    resume_time_minutes: 5,
    recovery_minutes: 0,
    interruption_count: 0,
    goal_id: '',
    session_id: '',
    tags: '',
    notes: '',
  });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  function patch(key, value) {
    setForm((old) => ({ ...old, [key]: value }));
  }

  async function submit(e) {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      const payload = {
        ...form,
        timestamp_start: toApiTime(form.timestamp_start),
        tags: form.tags.split(',').map((t) => t.trim()).filter(Boolean),
      };
      if (!payload.goal_id) delete payload.goal_id;
      if (!payload.session_id) delete payload.session_id;
      await api('/api/v1/friction-events', { method: 'POST', body: JSON.stringify(payload) });
      patch('notes', '');
      patch('tags', '');
      onCreated();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="card wide">
      <h2>Log friction</h2>
      <form onSubmit={submit} className="grid-form">
        <label>
          <span>Started</span>
          <input type="datetime-local" value={form.timestamp_start} onChange={(e) => patch('timestamp_start', e.target.value)} required />
        </label>
        <Select label="Workflow" value={form.workflow_stage} onChange={(v) => patch('workflow_stage', v)} values={workflowStages} />
        <Select label="Layer" value={form.friction_layer} onChange={(v) => patch('friction_layer', v)} values={frictionLayers} />
        <Select label="Type" value={form.friction_type} onChange={(v) => patch('friction_type', v)} values={frictionTypes} />
        <NumberInput label="Severity" min="1" max="5" value={form.severity_self} onChange={(v) => patch('severity_self', v)} />
        <NumberInput label="Cognitive load" min="1" max="5" value={form.cognitive_load_self} onChange={(v) => patch('cognitive_load_self', v)} />
        <NumberInput label="Emotion valence" min="-2" max="2" value={form.emotion_valence} onChange={(v) => patch('emotion_valence', v)} />
        <NumberInput label="Time lost, min" min="0" value={form.time_lost_minutes} onChange={(v) => patch('time_lost_minutes', v)} />
        <NumberInput label="Resume time, min" min="0" value={form.resume_time_minutes} onChange={(v) => patch('resume_time_minutes', v)} />
        <NumberInput label="Interruptions" min="0" value={form.interruption_count} onChange={(v) => patch('interruption_count', v)} />
        <label>
          <span>Goal</span>
          <select value={form.goal_id} onChange={(e) => patch('goal_id', e.target.value)}>
            <option value="">None</option>
            {goals.map((goal) => <option key={goal.id} value={goal.id}>{goal.title}</option>)}
          </select>
        </label>
        <label>
          <span>Session</span>
          <select value={form.session_id} onChange={(e) => patch('session_id', e.target.value)}>
            <option value="">None</option>
            {sessions.map((session) => <option key={session.id} value={session.id}>{session.title}</option>)}
          </select>
        </label>
        <label className="full">
          <span>Tags, comma-separated</span>
          <input value={form.tags} onChange={(e) => patch('tags', e.target.value)} placeholder="ci, flaky-test" />
        </label>
        <label className="full">
          <span>Notes</span>
          <textarea value={form.notes} onChange={(e) => patch('notes', e.target.value)} placeholder="What happened?" />
        </label>
        {error && <p className="error full">{error}</p>}
        <button className="primary" disabled={busy}>{busy ? 'Saving…' : 'Save event'}</button>
      </form>
    </section>
  );
}

function GoalSessionPanel({ onChanged }) {
  const [goalTitle, setGoalTitle] = useState('');
  const [goalStatus, setGoalStatus] = useState('active');
  const [sessionTitle, setSessionTitle] = useState('');
  const [error, setError] = useState('');

  async function createGoal(e) {
    e.preventDefault();
    setError('');
    try {
      await api('/api/v1/work-goals', { method: 'POST', body: JSON.stringify({ title: goalTitle, status: goalStatus }) });
      setGoalTitle('');
      onChanged();
    } catch (err) { setError(err.message); }
  }

  async function createSession(e) {
    e.preventDefault();
    setError('');
    try {
      await api('/api/v1/work-sessions', { method: 'POST', body: JSON.stringify({ title: sessionTitle, started_at: new Date().toISOString() }) });
      setSessionTitle('');
      onChanged();
    } catch (err) { setError(err.message); }
  }

  return (
    <section className="card">
      <h2>Goals & sessions</h2>
      <form onSubmit={createGoal} className="mini-form">
        <input value={goalTitle} onChange={(e) => setGoalTitle(e.target.value)} placeholder="New goal" />
        <select value={goalStatus} onChange={(e) => setGoalStatus(e.target.value)}>
          {goalStatuses.map((s) => <option key={s}>{s}</option>)}
        </select>
        <button>Create goal</button>
      </form>
      <form onSubmit={createSession} className="mini-form">
        <input value={sessionTitle} onChange={(e) => setSessionTitle(e.target.value)} placeholder="New session" />
        <button>Create session</button>
      </form>
      {error && <p className="error">{error}</p>}
    </section>
  );
}

function Dashboard({ events, snapshots, onCalculate }) {
  const totals = useMemo(() => {
    const byLayer = {};
    const byType = {};
    let time = 0;
    let cognitive = 0;
    for (const event of events) {
      time += event.time_lost_minutes || 0;
      cognitive += event.cognitive_load_self || 0;
      byLayer[event.friction_layer] = (byLayer[event.friction_layer] || 0) + 1;
      byType[event.friction_type] = (byType[event.friction_type] || 0) + (event.time_lost_minutes || 0);
    }
    return { time, avgCognitive: events.length ? (cognitive / events.length).toFixed(1) : '0.0', byLayer, byType };
  }, [events]);
  const latest = snapshots[0];

  return (
    <section className="card wide">
      <div className="section-heading">
        <h2>Dashboard</h2>
        <button onClick={onCalculate}>Calculate current scores</button>
      </div>
      <div className="cards-row">
        <div className="metric"><strong>{events.length}</strong><span>events</span></div>
        <div className="metric"><strong>{totals.time}</strong><span>minutes lost</span></div>
        <div className="metric"><strong>{totals.avgCognitive}</strong><span>avg cognitive load</span></div>
        <div className="metric"><strong>{latest?.scores?.cla?.toFixed?.(1) ?? '—'}</strong><span>CLA</span></div>
        <div className="metric"><strong>{latest?.scores?.fci?.toFixed?.(1) ?? '—'}</strong><span>FCI</span></div>
        <div className="metric"><strong>{latest?.scores?.sdc?.toFixed?.(2) ?? '—'}</strong><span>SDC</span></div>
      </div>
      <div className="analytics-grid">
        <Breakdown title="Events by layer" data={totals.byLayer} />
        <Breakdown title="Time lost by type" data={totals.byType} suffix=" min" />
      </div>
    </section>
  );
}

function Breakdown({ title, data, suffix = '' }) {
  const entries = Object.entries(data).sort((a, b) => b[1] - a[1]).slice(0, 8);
  return (
    <div className="breakdown">
      <h3>{title}</h3>
      {entries.length === 0 && <p className="muted">No data yet.</p>}
      {entries.map(([key, value]) => (
        <div key={key} className="bar-row">
          <span>{key}</span>
          <strong>{value}{suffix}</strong>
        </div>
      ))}
    </div>
  );
}

function Timeline({ events }) {
  return (
    <section className="card wide">
      <h2>Recent friction events</h2>
      {events.length === 0 && <p className="muted">No friction events yet.</p>}
      <div className="timeline">
        {events.map((event) => (
          <article key={event.id} className="event-item">
            <div>
              <strong>{event.friction_type}</strong>
              <p>{event.workflow_stage} · {event.friction_layer} · severity {event.severity_self} · load {event.cognitive_load_self}</p>
              {event.notes && <p>{event.notes}</p>}
            </div>
            <time>{new Date(event.timestamp_start).toLocaleString()}</time>
          </article>
        ))}
      </div>
    </section>
  );
}

function App() {
  const [events, setEvents] = useState([]);
  const [goals, setGoals] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [snapshots, setSnapshots] = useState([]);
  const [status, setStatus] = useState('loading');
  const [error, setError] = useState('');

  async function refresh() {
    setError('');
    try {
      const [statusRes, eventRes, goalRes, sessionRes, snapshotRes] = await Promise.all([
        api('/api/v1/status'),
        api('/api/v1/friction-events?limit=100'),
        api('/api/v1/work-goals?limit=100'),
        api('/api/v1/work-sessions?limit=100'),
        api('/api/v1/score-snapshots?limit=20'),
      ]);
      setStatus(statusRes.database?.ready ? 'ready' : 'not ready');
      setEvents(eventRes.events || []);
      setGoals(goalRes.goals || []);
      setSessions(sessionRes.sessions || []);
      setSnapshots(snapshotRes.snapshots || []);
    } catch (err) {
      setStatus('offline');
      setError(err.message);
    }
  }

  async function calculateScores() {
    const end = new Date();
    const start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000);
    try {
      await api('/api/v1/scores/calculate', { method: 'POST', body: JSON.stringify({ period_start: start.toISOString(), period_end: end.toISOString(), score_type: 'rolling_7d' }) });
      await refresh();
    } catch (err) {
      setError(err.message);
    }
  }

  useEffect(() => { refresh(); }, []);

  return (
    <main>
      <header>
        <div>
          <h1>Logarift</h1>
          <p>Local-first Developer Experience friction logging and analysis.</p>
        </div>
        <button onClick={refresh}>Refresh</button>
      </header>
      <p className={`status ${status}`}>Backend status: {status}</p>
      {error && <p className="error">{error}</p>}
      <div className="layout">
        <EventForm goals={goals} sessions={sessions} onCreated={refresh} />
        <GoalSessionPanel onChanged={refresh} />
        <Dashboard events={events} snapshots={snapshots} onCalculate={calculateScores} />
        <Timeline events={events} />
      </div>
    </main>
  );
}

createRoot(document.getElementById('root')).render(<App />);
