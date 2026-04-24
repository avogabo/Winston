import { useMemo, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { AlertTriangle, CheckCircle2, Film, FolderTree, Search, ShieldCheck, Sparkles, Tv, Wand2 } from 'lucide-react'
import './app.css'

type ReviewItem = {
  id: string
  source: string
  title: string
  year?: number
  kind: 'movie' | 'series'
  confidence: 'high' | 'medium' | 'low'
  state: 'needs_review' | 'approved' | 'imported'
  proposedPath: string
  reason: string
  candidates: { label: string; year?: number; tmdbId?: number }[]
}

const seed: ReviewItem[] = [
  {
    id: '1',
    source: '/nzb/series/el-internado.nzb',
    title: 'El Internado',
    year: 2007,
    kind: 'series',
    confidence: 'low',
    state: 'needs_review',
    proposedPath: 'Series/E/El Internado Las Cumbres (2021)/Temporada 01/El Internado Las Cumbres - 01x01',
    reason: 'Título ambiguo, varias coincidencias plausibles',
    candidates: [
      { label: 'El Internado', year: 2007, tmdbId: 37952 },
      { label: 'El Internado: Las Cumbres', year: 2021, tmdbId: 112233 },
    ],
  },
  {
    id: '2',
    source: '/nzb/pelis/1080/avatar.nzb',
    title: 'Avatar',
    year: 2009,
    kind: 'movie',
    confidence: 'high',
    state: 'approved',
    proposedPath: 'Peliculas/1080/A/Avatar (2009)/Avatar (2009).mkv',
    reason: 'tmdb_id explícito',
    candidates: [],
  },
]

export default function App() {
  const [query, setQuery] = useState('')
  const [selected, setSelected] = useState<ReviewItem | null>(seed[0])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return seed
    return seed.filter((item) => [item.title, item.source, item.proposedPath, item.reason].join(' ').toLowerCase().includes(q))
  }, [query])

  return (
    <div className="app-shell">
      <section className="hero">
        <div className="hero-glow" />
        <div className="hero-copy">
          <span className="badge"><Sparkles size={14} /> Winston Review Center</span>
          <h1>Preview, corrección y control antes de publicar en AltMount</h1>
          <p>
            Winston autoimporta cuando está claro, frena cuando duda y te deja corregir por TMDB ID,
            candidato o path final.
          </p>
          <div className="hero-stats">
            <Stat icon={<AlertTriangle size={16} />} label="en revisión" value="1" />
            <Stat icon={<CheckCircle2 size={16} />} label="aprobados" value="1" />
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

      <section className="layout">
        <div className="list-panel glass">
          <div className="panel-head">
            <h2>Items</h2>
            <span>{filtered.length} visibles</span>
          </div>
          <div className="review-list">
            {filtered.map((item) => (
              <button key={item.id} className={`review-row ${selected?.id === item.id ? 'active' : ''}`} onClick={() => setSelected(item)}>
                <div className="review-row-top">
                  <span className={`pill ${item.state}`}>{item.state}</span>
                  <span className={`pill confidence ${item.confidence}`}>{item.confidence}</span>
                </div>
                <div className="review-title">
                  {item.kind === 'series' ? <Tv size={16} /> : <Film size={16} />}
                  <strong>{item.title}</strong>
                </div>
                <div className="review-source"><FolderTree size={14} /> {item.source}</div>
              </button>
            ))}
          </div>
        </div>

        <AnimatePresence mode="wait">
          {selected && (
            <motion.div key={selected.id} initial={{ opacity: 0, y: 14 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -14 }} className="detail-panel glass">
              <div className="panel-head detail-head">
                <div>
                  <h2>{selected.title}</h2>
                  <p>{selected.reason}</p>
                </div>
                <button className="primary-btn"><Wand2 size={16} /> Recalcular</button>
              </div>

              <div className="grid-two">
                <Card title="Preview actual">
                  <PreviewRow label="Origen" value={selected.source} />
                  <PreviewRow label="Tipo" value={selected.kind} />
                  <PreviewRow label="Confianza" value={selected.confidence} />
                  <PreviewRow label="Ruta propuesta" value={selected.proposedPath} />
                </Card>

                <Card title="Corrección rápida">
                  <label className="field">
                    <span>TMDB ID</span>
                    <input placeholder="37952" />
                  </label>
                  <label className="field">
                    <span>Path override</span>
                    <input placeholder="Series/E/El Internado (2007)/Temporada 01/..." />
                  </label>
                  <div className="actions-row">
                    <button className="secondary-btn">Aplicar corrección</button>
                    <button className="ghost-btn">Aprobar e importar</button>
                  </div>
                </Card>
              </div>

              <Card title="Coincidencias sugeridas">
                <div className="candidate-list">
                  {selected.candidates.map((candidate) => (
                    <button key={`${candidate.label}-${candidate.tmdbId}`} className="candidate-row">
                      <div>
                        <strong>{candidate.label}</strong>
                        <span>{candidate.year ?? 's/f'} · tmdb {candidate.tmdbId ?? '-'}</span>
                      </div>
                      <span className="ghost-btn compact">Elegir</span>
                    </button>
                  ))}
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
