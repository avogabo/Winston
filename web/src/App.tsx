import { useEffect, useMemo, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { AlertTriangle, CheckCircle2, Film, FolderTree, LoaderCircle, Search, Settings2, ShieldCheck, Sparkles, Tv, Upload, Wand2 } from 'lucide-react'
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
  status?: string
}

type ReviewListResponse = { items: ReviewItem[] }

type Settings = {
  altmount_base_url: string
  altmount_api_key: string
  altmount_path_from: string
  altmount_path_to: string
  plex_base_url: string
  plex_token: string
  plex_path_from: string
  plex_path_to: string
  default_mode: string
  sleep_between_imports: string
  movies_template: string
  series_template: string
  filebot_movie_format: string
  filebot_series_format: string
  filebot_db: string
  filebot_binary: string
  filebot_home: string
  auto_import_medium: boolean
}

type FileBotStatus = {
  enabled: boolean
  available: boolean
  mode: string
  binary: string
  home: string
  db: string
  license_present: boolean
}

const defaultSettings: Settings = {
  altmount_base_url: '', altmount_api_key: '', altmount_path_from: '', altmount_path_to: '', plex_base_url: '', plex_token: '', plex_path_from: '', plex_path_to: '',
  default_mode: 'preserve', sleep_between_imports: '3s', movies_template: '', series_template: '', filebot_movie_format: '', filebot_series_format: '', filebot_db: 'TheMovieDB', filebot_binary: '/usr/local/bin/filebot', filebot_home: '/config/filebot', auto_import_medium: true,
}

export default function App() {
  const [tab, setTab] = useState<'review' | 'settings'>('review')
  const [items, setItems] = useState<ReviewItem[]>([])
  const [query, setQuery] = useState('')
  const [selectedSource, setSelectedSource] = useState('')
  const [selected, setSelected] = useState<ReviewItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [reloading, setReloading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [importing, setImporting] = useState(false)
  const [savingSettings, setSavingSettings] = useState(false)
  const [error, setError] = useState('')
  const [settingsMessage, setSettingsMessage] = useState('')
  const [fileBotStatus, setFileBotStatus] = useState<FileBotStatus | null>(null)
  const [tmdbId, setTmdbId] = useState('')
  const [pathOverride, setPathOverride] = useState('')
  const [settings, setSettings] = useState<Settings>(defaultSettings)

  useEffect(() => { void loadItems(); void loadSettings(); void loadFileBotStatus() }, [])
  useEffect(() => { if (!selectedSource && items.length > 0) setSelectedSource(items[0].source_nzb_path) }, [items, selectedSource])
  useEffect(() => { if (!selectedSource) { setSelected(null); return }; void loadItem(selectedSource) }, [selectedSource])
  useEffect(() => { if (!selected) return; setTmdbId(selected.metadata.tmdb_id ? String(selected.metadata.tmdb_id) : ''); setPathOverride(selected.metadata.relative_path_override || '') }, [selected])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return items
    return items.filter((item) => [item.metadata.title, item.source_nzb_path, item.proposed_path, item.reason, item.state, item.confidence].join(' ').toLowerCase().includes(q))
  }, [items, query])

  async function loadItems() {
    setLoading(true); setError('')
    try {
      const res = await fetch('/api/review/items')
      if (!res.ok) throw new Error('No pude cargar la lista')
      const data: ReviewListResponse = await res.json()
      setItems(data.items)
    } catch (err) { setError(err instanceof Error ? err.message : 'Error cargando items') } finally { setLoading(false) }
  }

  async function loadItem(source: string) {
    try {
      const res = await fetch(`/api/review/item?source=${encodeURIComponent(source)}`)
      if (!res.ok) throw new Error('No pude cargar el detalle')
      const data: ReviewItem = await res.json()
      setSelected(data)
    } catch (err) { setError(err instanceof Error ? err.message : 'Error cargando detalle') }
  }

  async function reloadSelected() {
    if (!selected) return
    setReloading(true); setError('')
    try {
      await loadItems()
      await loadItem(selected.source_nzb_path)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No pude recargar el item')
    } finally {
      setReloading(false)
    }
  }

  async function loadSettings() {
    try {
      const res = await fetch('/api/settings')
      if (!res.ok) throw new Error('No pude cargar ajustes')
      const data: Settings = await res.json()
      setSettings(data)
    } catch (err) { setError(err instanceof Error ? err.message : 'Error cargando ajustes') }
  }

  async function loadFileBotStatus() {
	try {
	  const res = await fetch('/api/filebot/status')
	  if (!res.ok) throw new Error('No pude cargar estado de FileBot')
	  const data: FileBotStatus = await res.json()
	  setFileBotStatus(data)
	} catch {
	  setFileBotStatus(null)
	}
  }

  async function postJSON(url: string, payload?: Record<string, unknown>) {
    const res = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload ?? {}) })
    if (!res.ok) throw new Error(await res.text())
    return res.json().catch(() => null)
  }

  async function applyCorrection(payload: Record<string, unknown>) {
    if (!selected) return
    setSaving(true); setError('')
    try {
      await postJSON(`/api/review/correct?source=${encodeURIComponent(selected.source_nzb_path)}`, payload)
      await loadItems(); await loadItem(selected.source_nzb_path)
    } catch (err) { setError(err instanceof Error ? err.message : 'No pude aplicar la corrección') } finally { setSaving(false) }
  }

  async function approveSelected() {
    if (!selected) return
    setSaving(true); setError('')
    try {
      await postJSON(`/api/review/approve?source=${encodeURIComponent(selected.source_nzb_path)}`)
      await loadItems(); await loadItem(selected.source_nzb_path)
    } catch (err) { setError(err instanceof Error ? err.message : 'No pude aprobar el item') } finally { setSaving(false) }
  }

  async function importSelected() {
    if (!selected) return
    setImporting(true); setError('')
    try {
      await postJSON(`/api/review/import?source=${encodeURIComponent(selected.source_nzb_path)}`)
      await loadItems(); await loadItem(selected.source_nzb_path)
    } catch (err) { setError(err instanceof Error ? err.message : 'No pude importar el item') } finally { setImporting(false) }
  }

  async function saveSettings() {
    setSavingSettings(true); setSettingsMessage(''); setError('')
    try {
      await postJSON('/api/settings', settings as unknown as Record<string, unknown>)
      setSettingsMessage('Ajustes guardados en la configuración persistente de Winston.')
      await loadFileBotStatus()
    } catch (err) { setError(err instanceof Error ? err.message : 'No pude guardar ajustes') } finally { setSavingSettings(false) }
  }

  const reviewCount = items.filter((item) => item.state === 'needs_review').length
  const approvedCount = items.filter((item) => item.state === 'approved' || item.state === 'corrected').length

  return (
    <div className="app-shell">
      <section className="hero"><div className="hero-glow" /><div className="hero-copy"><span className="badge"><Sparkles size={14} /> Winston Review Center</span><h1>Preview, corrección y control antes de publicar en AltMount</h1><p>Winston autoimporta cuando está claro, frena cuando duda y te deja corregir, aprobar e importar desde la propia UI.</p><div className="hero-stats"><Stat icon={<AlertTriangle size={16} />} label="en revisión" value={String(reviewCount)} /><Stat icon={<CheckCircle2 size={16} />} label="aprobados/corregidos" value={String(approvedCount)} /><Stat icon={<ShieldCheck size={16} />} label="NZBs borrados" value="0" /></div></div></section>
      <section className="tabs-row"><button className={`tab-btn ${tab === 'review' ? 'active' : ''}`} onClick={() => setTab('review')}>Revisión</button><button className={`tab-btn ${tab === 'settings' ? 'active' : ''}`} onClick={() => setTab('settings')}><Settings2 size={16} /> Ajustes</button></section>
      {tab === 'review' && <><section className="toolbar"><div className="search-box"><Search size={18} /><input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Busca por título, ruta, motivo..." /></div></section>{error && <div className="error-banner">{error}</div>}<section className="layout"><div className="list-panel glass"><div className="panel-head"><h2>Items</h2><span>{loading ? 'cargando...' : `${filtered.length} visibles`}</span></div><div className="review-list">{loading ? <div className="empty-state"><LoaderCircle className="spin" size={18} /> Cargando items...</div> : filtered.length === 0 ? <div className="empty-state">No hay items todavía en el estado de Winston.</div> : filtered.map((item) => (<button key={item.source_nzb_path} className={`review-row ${selectedSource === item.source_nzb_path ? 'active' : ''}`} onClick={() => setSelectedSource(item.source_nzb_path)}><div className="review-row-top"><span className={`pill ${item.state}`}>{item.state}</span><span className={`pill confidence ${item.confidence}`}>{item.confidence}</span></div><div className="review-title">{item.metadata.kind === 'series' ? <Tv size={16} /> : <Film size={16} />}<strong>{item.metadata.title || item.source_nzb_path.split('/').pop()}</strong></div><div className="review-source"><FolderTree size={14} /> {item.source_nzb_path}</div></button>))}</div></div><AnimatePresence mode="wait">{selected && <motion.div key={selected.source_nzb_path} initial={{ opacity: 0, y: 14 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -14 }} className="detail-panel glass"><div className="panel-head detail-head"><div><h2>{selected.metadata.title || selected.source_nzb_path.split('/').pop()}</h2><p>{selected.reason || 'Sin motivo calculado todavía'}</p></div><div className="inline-actions"><button className="secondary-btn" onClick={() => void approveSelected()} disabled={saving || importing || reloading}>Aprobar</button><button className="primary-btn" onClick={() => void importSelected()} disabled={saving || importing || reloading}><Upload size={16} /> {importing ? 'Importando...' : 'Importar'}</button><button className="ghost-btn" onClick={() => void reloadSelected()} disabled={saving || importing || reloading}><Wand2 size={16} /> {reloading ? 'Recargando...' : 'Recargar'}</button></div></div><div className="grid-two"><Card title="Preview actual"><PreviewRow label="Origen" value={selected.source_nzb_path} /><PreviewRow label="Tipo" value={selected.metadata.kind || '-'} /><PreviewRow label="Confianza" value={selected.confidence || '-'} /><PreviewRow label="Estado" value={selected.state || '-'} /><PreviewRow label="Ruta propuesta" value={selected.proposed_path || '-'} /></Card><Card title="Corrección rápida"><label className="field"><span>TMDB ID</span><input value={tmdbId} onChange={(e) => setTmdbId(e.target.value)} placeholder="37952" /></label><button className="secondary-btn" disabled={saving || importing || reloading || !tmdbId.trim()} onClick={() => void applyCorrection({ tmdb_id: Number(tmdbId) })}>{saving ? 'Aplicando...' : 'Aplicar TMDB ID'}</button><label className="field"><span>Path override</span><input value={pathOverride} onChange={(e) => setPathOverride(e.target.value)} placeholder="Series/E/El Internado (2007)/Temporada 01/..." /></label><button className="ghost-btn" disabled={saving || importing || reloading || !pathOverride.trim()} onClick={() => void applyCorrection({ relative_path_override: pathOverride })}>{saving ? 'Aplicando...' : 'Aplicar path override'}</button></Card></div><Card title="Coincidencias sugeridas"><div className="candidate-list">{selected.candidates?.length ? selected.candidates.map((candidate) => (<button key={`${candidate.label}-${candidate.tmdb_id}-${candidate.year}`} className="candidate-row" onClick={() => candidate.tmdb_id && void applyCorrection({ tmdb_id: candidate.tmdb_id })}><div><strong>{candidate.label}</strong><span>{candidate.year || 's/f'} · tmdb {candidate.tmdb_id || '-'} · {candidate.reason}</span></div><span className="ghost-btn compact">Elegir</span></button>)) : <div className="empty-state">No hay candidatos alternativos para este item.</div>}</div></Card></motion.div>}</AnimatePresence></section></>}
      {tab === 'settings' && <section className="settings-layout glass"><div className="panel-head"><div><h2>Ajustes</h2><p>Configuración persistente para AltMount, Plex, FileBot, naming y política de import.</p></div><button className="primary-btn" onClick={() => void saveSettings()} disabled={savingSettings}>{savingSettings ? 'Guardando...' : 'Guardar ajustes'}</button></div>{error && <div className="error-banner">{error}</div>}{settingsMessage && <div className="success-banner">{settingsMessage}</div>}<div className="settings-grid"><SettingsCard title="AltMount"><Field label="Base URL" value={settings.altmount_base_url} onChange={(value) => setSettings({ ...settings, altmount_base_url: value })} placeholder="http://192.168.1.100:8989" /><Field label="API Key" value={settings.altmount_api_key} onChange={(value) => setSettings({ ...settings, altmount_api_key: value })} placeholder="token" /><Field label="Ruta Winston NZB" value={settings.altmount_path_from} onChange={(value) => setSettings({ ...settings, altmount_path_from: value })} placeholder="/data/nzb" /><Field label="Ruta visible en AltMount" value={settings.altmount_path_to} onChange={(value) => setSettings({ ...settings, altmount_path_to: value })} placeholder="/config/.nzbs" /><Field label="Modo por defecto" value={settings.default_mode} onChange={(value) => setSettings({ ...settings, default_mode: value })} placeholder="preserve | template | filebot" /><Field label="Sleep entre imports" value={settings.sleep_between_imports} onChange={(value) => setSettings({ ...settings, sleep_between_imports: value })} placeholder="3s" /></SettingsCard><SettingsCard title="Plex"><Field label="Base URL" value={settings.plex_base_url} onChange={(value) => setSettings({ ...settings, plex_base_url: value })} placeholder="http://plex:32400" /><Field label="Token" value={settings.plex_token} onChange={(value) => setSettings({ ...settings, plex_token: value })} placeholder="plex-token" /><Field label="Path from" value={settings.plex_path_from} onChange={(value) => setSettings({ ...settings, plex_path_from: value })} placeholder="/home" /><Field label="Path to" value={settings.plex_path_to} onChange={(value) => setSettings({ ...settings, plex_path_to: value })} placeholder="/media/biblioteca" /></SettingsCard><SettingsCard title="FileBot"><Field label="Modo de import" value={settings.default_mode} onChange={(value) => setSettings({ ...settings, default_mode: value })} placeholder="filebot" /><Field label="FileBot DB" value={settings.filebot_db} onChange={(value) => setSettings({ ...settings, filebot_db: value })} placeholder="TheMovieDB" /><Field label="FileBot binary" value={settings.filebot_binary} onChange={(value) => setSettings({ ...settings, filebot_binary: value })} placeholder="/usr/local/bin/filebot" /><Field label="FileBot home" value={settings.filebot_home} onChange={(value) => setSettings({ ...settings, filebot_home: value })} placeholder="/config/filebot" /><Field label="FileBot movie format" value={settings.filebot_movie_format} onChange={(value) => setSettings({ ...settings, filebot_movie_format: value })} placeholder="Peliculas/{plex}" /><Field label="FileBot series format" value={settings.filebot_series_format} onChange={(value) => setSettings({ ...settings, filebot_series_format: value })} placeholder="Series/{alpha}/{series} ({year}) tvdb {id}/{episode.special ? 'Especiales' : 'Temporada ' + s00}/{series} - {s00e00} - {t}" />{fileBotStatus && <div className="preview-row"><span>Estado</span><code>{fileBotStatus.enabled ? 'modo filebot' : 'modo no-filebot'} · {fileBotStatus.available ? 'binario OK' : 'binario no disponible'} · {fileBotStatus.license_present ? 'licencia detectada' : 'licencia no detectada'}</code></div>} {fileBotStatus && <div className="preview-row"><span>Persistencia licencia</span><code>{fileBotStatus.home || '/config/filebot'}</code></div>}</SettingsCard><SettingsCard title="Naming fallback"><Field label="Movies template" value={settings.movies_template} onChange={(value) => setSettings({ ...settings, movies_template: value })} placeholder="Peliculas/{quality}/{alpha}/{title} ({year})" /><Field label="Series template" value={settings.series_template} onChange={(value) => setSettings({ ...settings, series_template: value })} placeholder="Series/{alpha}/{series}/Temporada {season}/{series} - {episode}" /></SettingsCard><SettingsCard title="Política de matching"><label className="check-row"><input type="checkbox" checked={settings.auto_import_medium} onChange={(e) => setSettings({ ...settings, auto_import_medium: e.target.checked })} /><span>Autoimportar `medium` si no hay conflicto</span></label></SettingsCard></div></section>}
    </div>
  )
}

function Stat({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) { return <div className="stat-card glass-soft"><span>{icon}</span><strong>{value}</strong><small>{label}</small></div> }
function Card({ title, children }: { title: string; children: React.ReactNode }) { return <div className="card glass-soft"><div className="card-head"><h3>{title}</h3></div><div className="card-body">{children}</div></div> }
function SettingsCard({ title, children }: { title: string; children: React.ReactNode }) { return <div className="card glass-soft"><div className="card-head"><h3>{title}</h3></div><div className="card-body">{children}</div></div> }
function Field({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (value: string) => void; placeholder?: string }) { return <label className="field"><span>{label}</span><input value={value} onChange={(e) => onChange(e.target.value)} placeholder={placeholder} /></label> }
function PreviewRow({ label, value }: { label: string; value: string }) { return <div className="preview-row"><span>{label}</span><code>{value}</code></div> }
