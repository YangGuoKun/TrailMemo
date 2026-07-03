interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: number
}

interface ChatRequest {
  message: string
  session_id?: string
}

interface ChatLoopResponse {
  content: string
  steps: number
  tool_calls: number
  total_tokens: number
  latency_ms: number
  finish_reason: string
}

interface RouteDraftRequest {
  session_id?: string
  query: string
  start_city?: string
  target_city?: string
  days?: number
  budget?: number
  travel_styles?: string[]
}

interface CheckpointDraftData {
  name: string
  description: string
  city: string
  address: string
  latitude: number
  longitude: number
  sequence: number
  arrive_time: string
  stay_duration: number
}

interface RouteDraftData {
  title: string
  summary: string
  start_city: string
  end_city: string
  estimated_budget: number
  estimated_hours: number
  checkpoints: CheckpointDraftData[]
}

interface RouteDraftResponse {
  run_id: string
  artifact_id: string
  route_draft: RouteDraftData
  warnings?: string[]
  approval_required: boolean
  next_action: string
}

interface RecommendRequest {
  query: string
  days?: number
  budget?: number
  interests?: string[]
  travel_type?: string
}

interface RecommendItem {
  title: string
  city: string
  reason: string
  estimated_budget: number
  days: number
  tags: string[]
}

interface RecommendResponse {
  run_id: string
  artifact_id: string
  items: RecommendItem[]
  fallback: boolean
  warnings?: string[]
}

interface TravelNoteRequest {
  route_id: number
  style?: string
  include_checkin_content?: boolean
  include_images?: boolean
}

interface TravelNoteResponse {
  run_id: string
  artifact_id: string
  title: string
  content: string
  suggested_tags: string[]
  image_refs?: string[]
  warnings?: string[]
  approval_required: boolean
  next_action: string
}

interface RemixRequest {
  query: string
  days?: number
  budget?: number
  travel_styles?: string[]
}

interface RemixChangeItem {
  action: string
  point: string
  reason: string
}

interface RemixResponse {
  run_id: string
  source_route_id: number
  artifact_id: string
  change_summary: RemixChangeItem[]
  route_draft: RouteDraftData
  warnings?: string[]
  approval_required: boolean
  next_action: string
}

interface ArtifactCommitRequest {
  commit_type: 'create_route' | 'create_post'
  idempotency_key: string
  is_public?: number
}

interface ArtifactCommitResponse {
  artifact_id: string
  status: string
  entity_type: string
  entity_id: number
}

interface ArtifactApprovalResponse {
  artifact_id: string
  status: string
}

interface RunStepInfo {
  index: number
  type: string
  name: string
  status: string
  latency_ms: number
  created_at: string
}

interface RunArtifactInfo {
  artifact_id: string
  type: string
  status: string
  committed_entity_type?: string
  committed_entity_id?: number
  created_at: string
}

interface RunDetailResponse {
  run_id: string
  user_id: number
  session_id?: string
  intent: string
  mode: string
  status: string
  model: string
  prompt_version?: string
  total_tokens: number
  latency_ms: number
  error_code?: string
  error_message?: string
  created_at: string
  updated_at: string
  steps: RunStepInfo[]
  artifacts: RunArtifactInfo[]
}

interface StreamChunk {
  content?: string
  error?: string
}

type QuickQuestion = {
  label: string
  icon: string
  query: string
  type?: string
}

interface PreferencesResponse {
  user_id: number
  travel_styles: string[]
  preferred_cities: string[]
  budget_range: [number, number]
  preferred_days: [number, number]
  interests: string[]
  avoid_list: string[]
  created_at: string
  updated_at: string
}

interface PreferencesUpdateRequest {
  travel_styles?: string[]
  preferred_cities?: string[]
  budget_range?: [number, number]
  preferred_days?: [number, number]
  interests?: string[]
  avoid_list?: string[]
}

interface HealthResponse {
  status: string
  enabled: boolean
  stage: string
  llm_configured: boolean
  default_mode: string
  request_timeout: string
  stream_timeout: string
}

interface CapabilityResponse {
  enabled: boolean
  default_mode: string
  max_steps: number
  max_tool_calls: number
  intents: string[]
  tools: ToolDescriptor[]
  stage: string
}

interface ToolDescriptor {
  name: string
  description: string
  parameters: Record<string, unknown>
}

interface SessionInfo {
  session_id: string
  title: string
  model: string
  message_count: number
  last_message_at?: string
  created_at: string
}

interface SessionListResponse {
  sessions: SessionInfo[]
  total: number
}

interface SessionMsg {
  role: string
  content: string
}

interface SessionDetailResponse {
  session_id: string
  title: string
  messages: SessionMsg[]
  message_count: number
  expired: boolean
}

interface RenameSessionRequest {
  title: string
}
