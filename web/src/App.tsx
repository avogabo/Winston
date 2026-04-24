import { useEffect, useMemo, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { AlertTriangle, CheckCircle2, Film, FolderTree, LoaderCircle, Search, ShieldCheck, Sparkles, Tv, Wand2 } from 'lucide-react'
import './app.css'

type ItemMetadata = {
  tmdb_id: number
  tvdb_id: number
  imdb_id: string
  kind: string
  title: string
  year: number
  season: number
  episode: number
  quality: string
  relative_path_override: string
}

type CandidateMatch = {
  label: string
  kind: string
  tmdb_id: number
  tvdb_id: number
  imdb_id: string
  year: number
  reason: string
  score: number
}

type ReviewItem = {
  source_nzb_path: string
  state: 'needs_review' | 'approved' | 'imported' | 'corrected' | 'failed' | 'importing'
  confidence: 'high' | 'medium' | 'low'
  metadata: ItemMetadata
  proposed_path: string
  reason: string
  candidates: CandidateMatch[]
  queue_id?: number
  status?: string
}

type ReviewListResponse = { items: ReviewItem[] }

export default function App() {
  const [items, setItems] = useState<ReviewItem[]>([])
  const [query, setQuery] = useState('')
  const [selectedSource, setSelectedSource] = useState('')
  const [selected, setSelected] = useState<ReviewItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [tmdbId, setTmdbId] = useState('')
  const [pathOverride, setPathOverride] = useState('')

  useEffect(() => {
    void loadItems()
  }, [])

  useEffect(() => {
    if (!selectedSource && items.length > 0) {
      setSelectedSource(items[0].source_nzb_path)
    }
  }, [items, selectedSource])

  useEffect(() => {
    if (!selectedSource) {
      setSelected(null)
      return
    }
    void loadItem(selectedSource)
  }, [selectedSource])

  useEffect(() => {
    if (!selected) return
    setTmdbId(selected.metadata.tmdb_id ? String(selected.metadata.tmdb_id) : '')
    setPathOverride(selected.metadata.relative_path_override || '')
  }, [selected])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return items
    return items.filter((item) =>
      [item.metadata.title, item.source_nzb_path, item.proposed_path, item.reason, item.state, item.confidence]
        .join(' ')
        .toLowerCase()
        .includes(q),
    )
  }, [items, query])

  async function loadItems() {
    setLoading(true)
    setError('')
    try {
      const res = await fetch('/api/review/items')
      if (!res.ok) throw new Error('No pude cargar la lista')
      const data: ReviewListResponse = await res.json()
      setItems(data.items)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error cargando items')
    } finally {
      setLoading(false)
    }
  }

  async function loadItem(source: string) {
    try {
      const res = await fetch(`/api/review/item?source=${encodeURIComponent(source)}`)
      if (!res.ok) throw new Error('No pude cargar el detalle')
      const data: ReviewItem = await res.json()
      setSelected(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error cargando detalle')
    }
  }

  async function applyCorrection(payload: Record<string, unknown>) {
    if (!selected) return
    setSaving(true)
    setError('')
    try {
      const res = await fetch(`/api/review/correct?source=${encodeURIComponent(selected.source_nzb_path)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
      if (!res.ok) throw new Error(await res.text())
      await loadItems()
      await loadItem(selected.source_nzb_path)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No pude aplicar la corrección')
    } finally {
      setSaving(false)
    }
  }

  const reviewCount = items.filter((item) => item.state === 'needs_review').length
  const approvedCount = items.filter((item) => item.state === 'approved' || item.state === 'corrected').length

  return (
    <div className="app-shell">
      <section className="hero">
        <div className="hero-glow" />
        <div className="hero-copy">
          <span className="badge"><Sparkles size={14} /> Winston Review Center</span>
          <h1>Preview, corrección y control antes de publicar en AltMount</h1>
          <p>Winston autoimporta cuando está claro, frena cuando duda y te deja corregir por TMDB ID o path final.</p>
          <div className="hero-stats">
            <Stat icon={<AlertTriangle size={16} />} label="en revisión" value={String(reviewCount)} />
            <Stat icon={<CheckCircle2 size={16} />} label="aprobados/corregidos" value={String(approvedCount)} />
            <Stat icon={<ShieldCheck size={16} />} label="NZBs borrados" value="0" />
          </div>
        </div>
      </section>

      <section className="toolbar">
        <div className="search-box">
          <Search size={18} />
          <input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Busca por título, ruta, motivo..." />
        </div>
      </section>

      {error && <div className="error-banner">{error}</div>}

      <section className="layout">
        <div className="list-panel glass">
          <div className="panel-head">
            <h2>Items</h2>
            <span>{loading ? 'cargando...' : `${filtered.length} visibles`}</span>
          </div>
          <div className="review-list">
            {loading ? (
              <div className="empty-state"><LoaderCircle className="spin" size={18} /> Cargando items...</div>
            ) : filtered.length === 0 ? (
              <div className="empty-state">No hay items todavía en el estado de Winston.</div>
            ) : (
              filtered.map((item) => (
                <button key={item.source_nzb_path} className={`review-row ${selectedSource === item.source_nzb_path ? 'active' : ''}`} onClick={() => setSelectedSource(item.source_nzb_path)}>
                  <div className="review-row-top">
                    <span className={`pill ${item.state}`}>{item.state}</span>
                    <span className={`pill confidence ${item.confidence}`}>{item.confidence}</span>
                  </div>
                  <div className="review-title">
                    {item.metadata.kind === 'series' ? <Tv size={16} /> : <Film size={16} />}
                    <strong>{item.metadata.title || item.source_nzb_path.split('/').pop()}</strong>
                  </div>
                  <div className="review-source"><FolderTree size={14} /> {item.source_nzb_path}</div>
                </button>
              ))
            )}
          </div>
        </div>

        <AnimatePresence mode="wait">
          {selected && (
            <motion.div key={selected.source_nzb_path} initial={{ opacity: 0, y: 14 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -14 }} className="detail-panel glass">
              <div className="panel-head detail-head">
                <div>
                  <h2>{selected.metadata.title || selected.source_nzb_path.split('/').pop()}</h2>
                  <p>{selected.reason || 'Sin motivo calculado todavía'}</p>
                </div>
                <button className="primary-btn" onClick={() => void loadItem(selected.source_nzb_path)}><Wand2 size={16} /> Recargar</button>
              </div>

              <div className="grid-two">
                <Card title="Preview actual">
                  <PreviewRow label="Origen" value={selected.source_nzb_path} />
                  <PreviewRow label="Tipo" value={selected.metadata.kind || '-'} />
                  <PreviewRow label="Confianza" value={selected.confidence || '-'} />
                  <PreviewRow label="Ruta propuesta" value={selected.proposed_path || '-'} />
                </Card>

                <Card title="Corrección rápida">
                  <label className="field">
                    <span>TMDB ID</span>
                    <input value={tmdbId} onChange={(e) => setTmdbId(e.target.value)} placeholder="37952" />
                  </label>
                  <button className="secondary-btn" disabled={saving || !tmdbId.trim()} onClick={() => void applyCorrection({ tmdb_id: Number(tmdbId) })}>
                    {saving ? 'Aplicando...' : 'Aplicar TMDB ID'}
                  </button>
                  <label className="field">
                    <span>Path override</span>
                    <input value={pathOverride} onChange={(e) => setPathOverride(e.target.value)} placeholder="Series/E/El Internado (2007)/Temporada 01/..." />
                  </label>
                  <button className="ghost-btn" disabled={saving || !pathOverride.trim()} onClick={() => void applyCorrection({ relative_path_override: pathOverride })}>
                    {saving ? 'Aplicando...' : 'Aplicar path override'}
                  </button>
                </Card>
              </div>

              <Card title="Coincidencias sugeridas">
                <div className="candidate-list">
                  {selected.candidates?.length ? selected.candidates.map((candidate) => (
                    <button key={`${candidate.label}-${candidate.tmdb_id}-${candidate.year}`} className="candidate-row" onClick={() => candidate.tmdb_id && void applyCorrection({ tmdb_id: candidate.tmdb_id })}>
                      <div>
                        <strong>{candidate.label}</strong>
                        <span>{candidate.year || 's/f'} · tmdb {candidate.tmdb_id || '-'} · {candidate.reason}</span>
                      </div>
                      <span className="ghost-btn compact">Elegir</span>
                    </button>
                  )) : <div className="empty-state">No hay candidatos alternativos para este item.</div>}
                </div>
              </Card>
            </motion.div>
          )}
        </AnimatePresence>
      </section>
    </div>
  )
}

function Stat({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="stat-card glass-soft">
      <span>{icon}</span>
      <strong>{value}</strong>
      <small>{label}</small>
    </div>
  )
}

function Card({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="card glass-soft">
      <div className="card-head">
        <h3>{title}</h3>
      </div>
      <div className="card-body">{children}</div>
    </div>
  )
}

function PreviewRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="preview-row">
      <span>{label}</span>
      <code>{value}</code>
    </div>
  )
}
