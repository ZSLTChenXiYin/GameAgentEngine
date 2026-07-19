import fetch from "node-fetch";

export class Client {
  private baseURL: string;
  private apiKey: string;
  private timeout: number;

  constructor(baseURL: string, apiKey: string, timeout = 30000) {
    this.baseURL = baseURL.replace(/\/+$/, "");
    this.apiKey = apiKey;
    this.timeout = timeout;
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-API-Key": this.apiKey,
    };
    const res = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
      signal: AbortSignal.timeout(this.timeout),
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`HTTP ${res.status} ${path}: ${text}`);
    }
    return res.json() as Promise<T>;
  }

  // === Health ===
  async health(): Promise<void> {
    await this.request<unknown>("GET", "/api/v1/health");
  }

  // === Nodes ===
  async getNodes(worldID: string, opts?: { limit?: number; offset?: number; nodeType?: string }): Promise<any[]> {
    const params = new URLSearchParams();
    if (opts?.limit) params.set("limit", String(opts.limit));
    if (opts?.offset) params.set("offset", String(opts.offset));
    if (opts?.nodeType) params.set("node_type", opts.nodeType);
    const qs = params.toString();
    return this.request<any[]>("GET", `/api/v1/worlds/${worldID}/nodes${qs ? "?" + qs : ""}`);
  }

  async getNode(id: string): Promise<any> {
    return this.request<any>("GET", `/api/v1/nodes/${id}`);
  }

  async createNode(worldID: string, name: string, nodeType: string, parentID?: string): Promise<any> {
    return this.request<any>("POST", `/api/v1/worlds/${worldID}/nodes`, { name, node_type: nodeType, parent_id: parentID });
  }

  async updateNode(id: string, data: { name?: string; node_type?: string; parent_id?: string | null }): Promise<any> {
    return this.request<any>("PUT", `/api/v1/nodes/${id}`, data);
  }

  async deleteNode(id: string): Promise<void> {
    await this.request<void>("DELETE", `/api/v1/nodes/${id}`);
  }

  // === Components ===
  async addComponent(nodeID: string, compType: string, data: string): Promise<any> {
    return this.request<any>("POST", `/api/v1/nodes/${nodeID}/components`, { component_type: compType, data });
  }

  async getComponents(nodeID: string): Promise<any[]> {
    return this.request<any[]>("GET", `/api/v1/nodes/${nodeID}/components`);
  }

  async getComponent(id: string): Promise<any> {
    return this.request<any>("GET", `/api/v1/components/${id}`);
  }

  async updateComponent(id: string, data: { component_type?: string; data?: string }): Promise<any> {
    return this.request<any>("PUT", `/api/v1/components/${id}`, data);
  }

  async deleteComponent(id: string): Promise<void> {
    await this.request<void>("DELETE", `/api/v1/components/${id}`);
  }

  // === Memories ===
  async addMemory(nodeID: string, content: string, level: string, tags?: string): Promise<any> {
    return this.request<any>("POST", `/api/v1/nodes/${nodeID}/memories`, { content, level, tags });
  }

  async getMemories(nodeID: string): Promise<any[]> {
    return this.request<any[]>("GET", `/api/v1/nodes/${nodeID}/memories`);
  }

  // === Relations ===
  async getRelations(worldID: string, opts?: { limit?: number; offset?: number; relationType?: string }): Promise<any[]> {
    const params = new URLSearchParams();
    if (opts?.limit) params.set("limit", String(opts.limit));
    if (opts?.offset) params.set("offset", String(opts.offset));
    if (opts?.relationType) params.set("relation_type", opts.relationType);
    const qs = params.toString();
    return this.request<any[]>("GET", `/api/v1/worlds/${worldID}/relations${qs ? "?" + qs : ""}`);
  }

  async createRelation(worldID: string, sourceID: string, targetID: string, relationType: string, weight?: number): Promise<any> {
    return this.request<any>("POST", `/api/v1/worlds/${worldID}/relations`, { source_id: sourceID, target_id: targetID, relation_type: relationType, weight: weight ?? 0 });
  }

  // === World ===
  async getWorlds(): Promise<any[]> {
    return this.request<any[]>("GET", "/api/v1/worlds");
  }

  async updateWorld(worldID: string, name: string): Promise<any> {
    return this.request<any>("PUT", `/api/v1/worlds/${worldID}`, { name });
  }

  // === Invoke ===
  async invoke(req: any): Promise<any> {
    return this.request<any>("POST", "/api/v1/invoke", req);
  }

  async executeInteraction(req: any): Promise<any> {
    return this.request<any>("POST", "/api/v1/interaction/execute", req);
  }

  async interpretPlayerInput(req: any): Promise<any> {
    return this.request<any>("POST", "/api/v1/player/input/interpret", req);
  }

  // === Tick ===
  async advanceTick(worldID: string, tickType?: string, gameTime?: string, autonomousLimit?: number): Promise<any> {
    return this.request<any>("POST", `/api/v1/worlds/${worldID}/tick`, {
      tick_type: tickType ?? "world_tick",
      game_time: gameTime,
      autonomous_limit: autonomousLimit,
    });
  }

  // === Autonomous ===
  async runAutonomousNode(worldID: string, nodeID: string): Promise<any> {
    return this.request<any>("POST", `/api/v1/worlds/${worldID}/autonomous`, { node_id: nodeID });
  }

  // === Auth Query ===
  async dispatchAuthorityQuery(worldID: string, query: any): Promise<any> {
    return this.request<any>("POST", `/api/v1/worlds/${worldID}/authority/query`, query);
  }

  async handleAuthorityCallback(worldID: string, taskID: string, response: any): Promise<any> {
    return this.request<any>("POST", `/api/v1/worlds/${worldID}/authority/callback/${taskID}`, response);
  }
}
