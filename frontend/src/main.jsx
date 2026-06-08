import React, { useEffect, useMemo, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

const configuredAPIBase = import.meta.env.VITE_LOGARIFT_API_BASE_URL;
const API_BASE = (configuredAPIBase === undefined ? 'http://localhost:8080' : configuredAPIBase).replace(/\/$/, '');
const goalStatuses = ['active', 'completed', 'deferred', 'abandoned'];
const levels = [
  { value: 'green', label: 'Papercut', description: 'Small annoyance', tooltip: 'Use green when the friction was visible but did not break flow.' },
  { value: 'yellow', label: 'Annoying', description: 'Meaningful slowdown', tooltip: 'Use yellow when the issue slowed you down but you kept moving.' },
  { value: 'orange', label: 'Disruptive', description: 'Lost flow or time', tooltip: 'Use orange when the issue broke flow, caused rework, or cost meaningful time.' },
  { value: 'red', label: 'Blocked', description: 'Could not progress', tooltip: 'Use red when the issue blocked progress or caused high frustration.' },
];

function nowLocalInput() {
  const date = new Date();
  date.setMinutes(date.getMinutes() - date.getTimezoneOffset());
  return date.toISOString().slice(0, 16);
}

function toApiTime(local) {
  return new Date(local).toISOString();
}

function defaultHeaders(options) {
  if (options.body instanceof FormData) return options.headers || {};
  return { 'Content-Type': 'application/json', ...(options.headers || {}) };
}

function apiURL(path) {
  return `${API_BASE}${path}`;
}

async function api(path, options = {}) {
  const response = await fetch(apiURL(path), {
    ...options,
    headers: defaultHeaders(options),
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

function stripHtml(html) {
  const div = document.createElement('div');
  div.innerHTML = html || '';
  return div.textContent || div.innerText || '';
}

function sanitizeRichHtml(html) {
  const allowedTags = new Set(['A', 'B', 'BR', 'CODE', 'DIV', 'EM', 'I', 'IMG', 'LI', 'OL', 'P', 'PRE', 'SPAN', 'STRONG', 'UL']);
  const allowedAttrs = {
    A: new Set(['href', 'title', 'target', 'rel']),
    IMG: new Set(['src', 'alt', 'title']),
  };
  const doc = new DOMParser().parseFromString(`<div>${html || ''}</div>`, 'text/html');
  const root = doc.body.firstElementChild;

  function clean(node) {
    for (const child of [...node.childNodes]) {
      if (child.nodeType === Node.ELEMENT_NODE) {
        if (!allowedTags.has(child.tagName)) {
          child.replaceWith(...child.childNodes);
          continue;
        }
        for (const attr of [...child.attributes]) {
          const allowed = allowedAttrs[child.tagName]?.has(attr.name.toLowerCase()) || false;
          if (!allowed) child.removeAttribute(attr.name);
        }
        if (child.tagName === 'A') {
          const href = child.getAttribute('href') || '';
          if (!href.startsWith('http://') && !href.startsWith('https://') && !href.startsWith('/uploads/')) {
            child.removeAttribute('href');
          } else {
            child.setAttribute('target', '_blank');
            child.setAttribute('rel', 'noreferrer');
          }
        }
        if (child.tagName === 'IMG') {
          const src = child.getAttribute('src') || '';
          if (!src.startsWith('http://') && !src.startsWith('https://') && !src.startsWith('/uploads/')) {
            child.remove();
            continue;
          }
        }
        clean(child);
      }
    }
  }

  clean(root);
  return root.innerHTML;
}

function InfoTip({ text }) {
  return <span className="info-tip" title={text} aria-label={text}>?</span>;
}

function FieldLabel({ children, tooltip }) {
  return (
    <span className="field-label">
      {children}
      {tooltip && <InfoTip text={tooltip} />}
    </span>
  );
}

function RichNotesEditor({ value, onChange }) {
  const editorRef = useRef(null);
  const fileRef = useRef(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState('');

  useEffect(() => {
    if (editorRef.current && editorRef.current.innerHTML !== value) {
      editorRef.current.innerHTML = value;
    }
  }, [value]);

  function sync() {
    onChange(sanitizeRichHtml(editorRef.current?.innerHTML || ''));
  }

  function exec(command) {
    editorRef.current?.focus();
    document.execCommand(command, false, null);
    sync();
  }

  function insertHTML(html) {
    editorRef.current?.focus();
    document.execCommand('insertHTML', false, html);
    sync();
  }

  function addLink() {
    const url = window.prompt('Paste URL');
    if (!url) return;
    const safeURL = url.trim();
    if (!safeURL.startsWith('http://') && !safeURL.startsWith('https://')) {
      setUploadError('Only http/https links are accepted.');
      return;
    }
    insertHTML(`<a href="${safeURL}" target="_blank" rel="noreferrer">${safeURL}</a>`);
  }

  async function uploadImage(file) {
    if (!file || !file.type.startsWith('image/')) return;
    setUploading(true);
    setUploadError('');
    try {
      const formData = new FormData();
      formData.append('file', file);
      const uploaded = await api('/api/v1/uploads', { method: 'POST', body: formData });
      const url = `${API_BASE}${uploaded.url_path}`;
      const alt = uploaded.filename || 'screenshot';
      insertHTML(`<p><img src="${url}" alt="${alt}" title="${alt}"></p>`);
    } catch (err) {
      setUploadError(err.message);
    } finally {
      setUploading(false);
    }
  }

  async function uploadFiles(files) {
    for (const file of [...files]) {
      await uploadImage(file);
    }
  }

  function onPaste(e) {
    const imageFiles = [...(e.clipboardData?.files || [])].filter((file) => file.type.startsWith('image/'));
    if (imageFiles.length > 0) {
      e.preventDefault();
      uploadFiles(imageFiles);
    }
  }

  function onDrop(e) {
    const imageFiles = [...(e.dataTransfer?.files || [])].filter((file) => file.type.startsWith('image/'));
    if (imageFiles.length > 0) {
      e.preventDefault();
      uploadFiles(imageFiles);
    }
  }

  return (
    <div className="rich-editor-shell">
      <div className="editor-toolbar" aria-label="Rich notes toolbar">
        <button type="button" title="Bold selected text" onClick={() => exec('bold')}>Bold</button>
        <button type="button" title="Italic selected text" onClick={() => exec('italic')}>Italic</button>
        <button type="button" title="Insert a clickable link" onClick={addLink}>Link</button>
        <button type="button" title="Upload or attach a local screenshot" onClick={() => fileRef.current?.click()} disabled={uploading}>
          {uploading ? 'Uploading…' : 'Screenshot'}
        </button>
        <input ref={fileRef} className="hidden-input" type="file" accept="image/png,image/jpeg,image/webp,image/gif" multiple onChange={(e) => uploadFiles(e.target.files || [])} />
      </div>
      <div
        ref={editorRef}
        className="rich-editor"
        contentEditable
        role="textbox"
        aria-multiline="true"
        title="Write notes, paste links, paste screenshots from clipboard, or click Screenshot to upload images."
        data-placeholder="Example: CI failed again after 20 min with an unclear timeout. Paste links or screenshots here."
        onInput={sync}
        onBlur={sync}
        onPaste={onPaste}
        onDrop={onDrop}
        suppressContentEditableWarning
      />
      <div className="editor-help">
        <span>Supports formatted text, links, drag/drop images, file upload, and pasted screenshots.</span>
        {uploadError && <span className="editor-error">{uploadError}</span>}
      </div>
    </div>
  );
}

function EventComposer({ onCreated }) {
  const [occurredAt, setOccurredAt] = useState(nowLocalInput());
  const [level, setLevel] = useState('yellow');
  const [notes, setNotes] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [lastEvent, setLastEvent] = useState(null);
  const [lastEnrichment, setLastEnrichment] = useState(null);

  async function submit(e) {
    e.preventDefault();
    const cleanNotes = sanitizeRichHtml(notes);
    if (stripHtml(cleanNotes).trim() === '' && !cleanNotes.includes('<img')) {
      setError('Notes are required. Describe what happened or attach a screenshot.');
      return;
    }
    setBusy(true);
    setError('');
    try {
      const payload = {
        occurred_at: toApiTime(occurredAt),
        friction_level: level,
        notes_markdown: cleanNotes,
      };
      const result = await api('/api/v1/friction-events/quick', { method: 'POST', body: JSON.stringify(payload) });
      setLastEvent(result.event);
      setLastEnrichment(result.enrichment || result.event?.enrichment || null);
      if (result.enrichment?.job_id) {
        streamEnrichment(result.event.id, result.enrichment.job_id);
      }
      setNotes('');
      setOccurredAt(nowLocalInput());
      onCreated();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function streamEnrichment(eventId, jobId) {
    const terminal = new Set(['succeeded', 'partially_succeeded', 'failed', 'timed_out', 'cancelled', 'disabled', 'not_queued']);

    if (!window.EventSource) {
      pollEnrichment(eventId, jobId, terminal);
      return;
    }

    const source = new EventSource(apiURL(`/api/v1/enrichment-jobs/${jobId}/events`));
    source.addEventListener('enrichment', async (message) => {
      try {
        const jobRes = JSON.parse(message.data);
        const eventRes = await api(`/api/v1/friction-events/${eventId}`);
        setLastEvent(eventRes.event);
        const nextEnrichment = mergeEnrichmentWithJob(eventRes.event?.enrichment, jobRes, jobId);
        setLastEnrichment(nextEnrichment);
        onCreated();
        if (terminal.has(nextEnrichment.llm_status)) {
          source.close();
        }
      } catch (err) {
        source.close();
        setLastEnrichment((current) => current ? { ...current, user_message: err.message } : current);
      }
    });
    source.addEventListener('error', () => {
      source.close();
      pollEnrichment(eventId, jobId, terminal);
    });
  }

  async function pollEnrichment(eventId, jobId, terminal = new Set(['succeeded', 'partially_succeeded', 'failed', 'timed_out', 'cancelled', 'disabled', 'not_queued'])) {
    for (let attempt = 0; attempt < 20; attempt += 1) {
      await new Promise((resolve) => setTimeout(resolve, 750));
      try {
        const [eventRes, jobRes] = await Promise.all([
          api(`/api/v1/friction-events/${eventId}`),
          api(`/api/v1/enrichment-jobs/${jobId}`),
        ]);
        setLastEvent(eventRes.event);
        const nextEnrichment = mergeEnrichmentWithJob(eventRes.event?.enrichment, jobRes, jobId);
        setLastEnrichment(nextEnrichment);
        onCreated();
        if (terminal.has(nextEnrichment.llm_status)) return;
      } catch (err) {
        setLastEnrichment((current) => current ? { ...current, user_message: err.message } : current);
        return;
      }
    }
  }

  return (
    <section className="card composer-card">
      <div className="composer-heading">
        <div>
          <h2>Log friction</h2>
          <p>Three fields only. Logarift infers workflow, layer, type, time loss, and tags locally with deterministic rules.</p>
        </div>
        <span className="pill subtle" title="Only date, color, and notes are required.">3 fields</span>
      </div>
      <form onSubmit={submit} className="composer-form">
        <label className="date-field">
          <FieldLabel tooltip="When the friction happened. Defaults to now so logging stays fast.">When?</FieldLabel>
          <input type="datetime-local" value={occurredAt} onChange={(e) => setOccurredAt(e.target.value)} required title="Timestamp for this friction log." />
        </label>

        <fieldset className="level-picker">
          <legend><FieldLabel tooltip="A simple color frustration score. The backend maps it to severity, cognitive load, and emotion valence.">How bad was it?</FieldLabel></legend>
          <div className="level-options">
            {levels.map((option) => (
              <button
                type="button"
                key={option.value}
                className={`level-button ${option.value} ${level === option.value ? 'selected' : ''}`}
                onClick={() => setLevel(option.value)}
                aria-pressed={level === option.value}
                title={option.tooltip}
              >
                <span className="level-dot" />
                <strong>{option.label}</strong>
                <small>{option.description}</small>
              </button>
            ))}
          </div>
        </fieldset>

        <label className="notes-field">
          <FieldLabel tooltip="Write what happened. Links and screenshots help the deterministic rules infer better metadata.">Notes</FieldLabel>
          <RichNotesEditor value={notes} onChange={setNotes} />
        </label>

        {error && <p className="error full">{error}</p>}
        <button className="primary composer-submit" disabled={busy} title="Save the friction log and run local deterministic enrichment.">{busy ? 'Saving…' : 'Save friction'}</button>
      </form>
      {lastEvent && <InferencePreview event={lastEvent} enrichment={lastEnrichment || lastEvent.enrichment} />}
    </section>
  );
}

function InferencePreview({ event, enrichment }) {
  return (
    <div className="inference-preview" title="These fields were inferred locally from the color and notes. They can be corrected later in advanced editing.">
      <div>
        <strong>Inferred locally</strong>
        <p>{event.workflow_stage} · {event.friction_layer} · {event.friction_type} · ~{event.time_lost_minutes} min lost</p>
      </div>
      <div className="preview-badges">
        <ConfidenceBadge event={event} />
        <EnrichmentStatus enrichment={enrichment || event.enrichment} />
      </div>
    </div>
  );
}

function mergeEnrichmentWithJob(enrichment, jobResponse, jobId) {
  const job = jobResponse?.job || {};
  const mergeSummary = enrichment?.merge_summary || job.merge_summary || jobResponse?.merge_summary || null;
  return {
    ...(enrichment || {}),
    llm_status: enrichment?.llm_status || job.status || 'not_requested',
    job_id: enrichment?.job_id || job.id || jobId,
    merge_summary: mergeSummary || undefined,
  };
}

function countLabel(count, singular, plural = `${singular}s`) {
  if (!Number.isFinite(count)) return null;
  return `${count} ${count === 1 ? singular : plural}`;
}

function fieldCountParts(summary) {
  if (!summary) return [];
  return [
    countLabel(summary.accepted_field_count, 'accepted field'),
    countLabel(summary.rejected_field_count, 'rejected field'),
    countLabel(summary.fallback_field_count, 'fallback field'),
  ].filter(Boolean);
}

function enrichmentDisplay(enrichment) {
  const status = enrichment?.llm_status || 'not_requested';
  const summary = enrichment?.merge_summary;
  const rejectedCount = summary?.rejected_field_count || 0;
  const fallbackCount = summary?.fallback_field_count || 0;
  const hasFallback = status === 'partially_succeeded' || rejectedCount > 0 || fallbackCount > 0;
  const counts = fieldCountParts(summary);
  const countSuffix = counts.length > 0 ? ` (${counts.join(', ')})` : '';

  if ((status === 'succeeded' || status === 'partially_succeeded') && hasFallback) {
    return { label: `applied with fallback${countSuffix}`, titleStatus: 'applied with fallback' };
  }
  if (status === 'succeeded') return { label: `applied${countSuffix}`, titleStatus: 'applied' };
  if (status === 'partially_succeeded') return { label: `applied with fallback${countSuffix}`, titleStatus: 'applied with fallback' };

  return { label: status.replaceAll('_', ' '), titleStatus: status.replaceAll('_', ' ') };
}

function EnrichmentStatus({ enrichment }) {
  const status = enrichment?.llm_status || 'not_requested';
  const display = enrichmentDisplay(enrichment);
  const title = enrichment?.user_message || `LLM enrichment status: ${display.titleStatus}`;
  return <span className={`pill enrichment-${status}`} title={title}>LLM: {display.label}</span>;
}

function ConfidenceBadge({ event }) {
  const confidence = averageConfidence(event);
  if (confidence === null) return <span className="pill subtle" title="This event was created through advanced/manual fields.">advanced</span>;
  return <span className="pill" title="Average confidence of locally inferred fields.">{Math.round(confidence * 100)}% confidence</span>;
}

function averageConfidence(event) {
  const fields = event?.inference?.fields;
  if (!fields) return null;
  const values = Object.values(fields).map((field) => Number(field.confidence)).filter((value) => Number.isFinite(value));
  if (!values.length) return null;
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function GoalSessionPanel({ onChanged, onClose }) {
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
    <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
      <section className="card modal-card" role="dialog" aria-modal="true" aria-labelledby="context-title" onMouseDown={(e) => e.stopPropagation()}>
        <div className="section-heading">
          <div>
            <h2 id="context-title">Optional context</h2>
            <p className="muted">Goals and sessions are optional. They stay out of the main logging flow.</p>
          </div>
          <button type="button" title="Close optional context" onClick={onClose}>Close</button>
        </div>
        <form onSubmit={createGoal} className="mini-form">
          <label>
            <FieldLabel tooltip="A goal is a meaningful work outcome, such as fixing a bug or implementing a feature.">New goal</FieldLabel>
            <input value={goalTitle} onChange={(e) => setGoalTitle(e.target.value)} placeholder="Implement dashboard filters" title="Goal title" />
          </label>
          <label>
            <FieldLabel tooltip="Goal status is optional metadata for later filtering.">Goal status</FieldLabel>
            <select value={goalStatus} onChange={(e) => setGoalStatus(e.target.value)} title="Goal status">
              {goalStatuses.map((s) => <option key={s}>{s}</option>)}
            </select>
          </label>
          <button title="Create a work goal">Create goal</button>
        </form>
        <form onSubmit={createSession} className="mini-form">
          <label>
            <FieldLabel tooltip="A session is a bounded block of work. Quick logging does not require selecting one.">New session</FieldLabel>
            <input value={sessionTitle} onChange={(e) => setSessionTitle(e.target.value)} placeholder="Morning debugging" title="Session title" />
          </label>
          <button title="Create a work session starting now">Create session</button>
        </form>
        {error && <p className="error">{error}</p>}
      </section>
    </div>
  );
}

function Metric({ value, label, tooltip }) {
  return (
    <div className="metric" title={tooltip}>
      <strong>{value}</strong>
      <span>{label} <InfoTip text={tooltip} /></span>
    </div>
  );
}

function Dashboard({ events, snapshots, onCalculate }) {
  const totals = useMemo(() => {
    const byLayer = {};
    const byType = {};
    let time = 0;
    let cognitive = 0;
    let inferred = 0;
    let confidenceSum = 0;
    for (const event of events) {
      time += event.time_lost_minutes || 0;
      cognitive += event.cognitive_load_self || 0;
      byLayer[event.friction_layer] = (byLayer[event.friction_layer] || 0) + 1;
      byType[event.friction_type] = (byType[event.friction_type] || 0) + (event.time_lost_minutes || 0);
      const confidence = averageConfidence(event);
      if (confidence !== null) {
        inferred += 1;
        confidenceSum += confidence;
      }
    }
    return {
      time,
      avgCognitive: events.length ? (cognitive / events.length).toFixed(1) : '0.0',
      byLayer,
      byType,
      inferred,
      avgConfidence: inferred ? `${Math.round((confidenceSum / inferred) * 100)}%` : '—',
    };
  }, [events]);
  const latest = snapshots[0];

  return (
    <section className="card dashboard-card">
      <div className="section-heading">
        <div>
          <h2>Dashboard</h2>
          <p className="muted">Analytics are separated from logging so the first screen remains fast.</p>
        </div>
        <button onClick={onCalculate} title="Ask the Go API to send current events to the C++ math-engine service and store a score snapshot.">Calculate current scores</button>
      </div>
      <div className="cards-row">
        <Metric value={events.length} label="events" tooltip="Number of friction logs currently loaded in the dashboard." />
        <Metric value={totals.time} label="minutes lost" tooltip="Sum of estimated or observed time_lost_minutes across loaded events." />
        <Metric value={totals.avgCognitive} label="avg cognitive load" tooltip="Average self or inferred cognitive-load score. 1 is low effort; 5 is exhausting or deeply confusing." />
        <Metric value={totals.avgConfidence} label="avg inference confidence" tooltip="Average confidence for fields inferred by the deterministic local rules engine." />
        <Metric value={latest?.scores?.cla?.toFixed?.(1) ?? '—'} label="CLA" tooltip="Cognitive Load Accumulator. A decayed score estimating accumulated mental pressure from severity, cognitive load, interruptions, resume time, and recovery." />
        <Metric value={latest?.scores?.fci?.toFixed?.(1) ?? '—'} label="FCI" tooltip="Friction Compounding Index. A time-decayed score estimating whether friction events cluster and compound near the scoring period end." />
        <Metric value={latest?.scores?.sdc?.toFixed?.(2) ?? '—'} label="SDC" tooltip="Systemic Drag Coefficient. Wait-like friction time divided by active work time; higher means more recorded waiting burden." />
      </div>
      <div className="analytics-grid">
        <Breakdown title="Events by layer" tooltip="Friction layer is inferred from notes, for example technical, temporal, cognitive, or social/process." data={totals.byLayer} />
        <Breakdown title="Time lost by type" tooltip="Friction type is a more specific classification, such as failed_feedback or waiting_for_review." data={totals.byType} suffix=" min" />
      </div>
    </section>
  );
}

function Breakdown({ title, tooltip, data, suffix = '' }) {
  const entries = Object.entries(data).sort((a, b) => b[1] - a[1]).slice(0, 8);
  return (
    <div className="breakdown" title={tooltip}>
      <h3>{title} <InfoTip text={tooltip} /></h3>
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
    <section className="card recent-card">
      <h2>Recent logs</h2>
      {events.length === 0 && <p className="muted">No friction logs yet.</p>}
      <div className="timeline">
        {events.slice(0, 12).map((event) => (
          <article key={event.id} className={`event-item level-${event.observed?.friction_level || 'advanced'}`} title="A recent friction log with canonical fields inferred for analytics.">
            <div className="event-main">
              <div className="event-title-row">
                <strong>{event.friction_type}</strong>
                <ConfidenceBadge event={event} />
                <EnrichmentStatus enrichment={event.enrichment} />
              </div>
              <p title="Inferred canonical metadata used by charts and scoring.">{event.workflow_stage} · {event.friction_layer} · severity {event.severity_self} · load {event.cognitive_load_self}</p>
              {event.notes && <div className="event-notes" dangerouslySetInnerHTML={{ __html: sanitizeRichHtml(event.notes) }} />}
              {event.tags?.length > 0 && (
                <div className="tag-row" title="Tags inferred from notes and links.">{event.tags.map((tag) => <span key={tag}>{tag}</span>)}</div>
              )}
            </div>
            <time title="When the friction was recorded as happening.">{new Date(event.timestamp_start).toLocaleString()}</time>
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
  const [activeTab, setActiveTab] = useState('log');
  const [contextOpen, setContextOpen] = useState(false);

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
        <div className="header-actions">
          <button onClick={() => setContextOpen(true)} title="Create optional goals or sessions in a modal. This does not interrupt quick logging.">Optional context</button>
          <button onClick={refresh} title="Reload backend status, recent logs, goals, sessions, and score snapshots.">Refresh</button>
        </div>
      </header>

      <nav className="tabs" aria-label="Main sections">
        <button className={activeTab === 'log' ? 'active' : ''} onClick={() => setActiveTab('log')} title="Fast friction logging and recent logs.">Log</button>
        <button className={activeTab === 'dashboard' ? 'active' : ''} onClick={() => setActiveTab('dashboard')} title="Analytics, score cards, and breakdowns.">Dashboard</button>
      </nav>

      <p className={`status ${status}`} title="Backend readiness based on the API status endpoint.">Backend status: {status}</p>
      {error && <p className="error">{error}</p>}

      {activeTab === 'log' && (
        <div className="single-column">
          <EventComposer onCreated={refresh} />
          <Timeline events={events} />
        </div>
      )}

      {activeTab === 'dashboard' && <Dashboard events={events} snapshots={snapshots} onCalculate={calculateScores} />}

      {contextOpen && <GoalSessionPanel goals={goals} sessions={sessions} onChanged={refresh} onClose={() => setContextOpen(false)} />}
    </main>
  );
}

createRoot(document.getElementById('root')).render(<App />);
