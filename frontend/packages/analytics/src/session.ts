const SESSION_ID_KEY = 'voxera_session_id';
const SESSION_START_KEY = 'voxera_session_start';
const SESSION_LAST_ACTIVE_KEY = 'voxera_session_last_active';

export class SessionManager {
  private sessionId: string;
  private sessionStart: number;
  private lastActive: number;
  private timeoutMs: number;

  constructor(timeoutMs = 30 * 60 * 1000) {
    this.timeoutMs = timeoutMs;

    const stored = this.restore();
    if (stored && Date.now() - stored.lastActive < this.timeoutMs) {
      this.sessionId = stored.sessionId;
      this.sessionStart = stored.sessionStart;
      this.lastActive = stored.lastActive;
    } else {
      this.sessionId = crypto.randomUUID();
      this.sessionStart = Date.now();
      this.lastActive = Date.now();
      this.persist();
    }
  }

  getSessionId(): string {
    this.rotateIfExpired();
    return this.sessionId;
  }

  getSessionStart(): number {
    return this.sessionStart;
  }

  getSessionDuration(): number {
    return Date.now() - this.sessionStart;
  }

  touch(): void {
    this.rotateIfExpired();
    this.lastActive = Date.now();
    this.persist();
  }

  reset(): void {
    this.sessionId = crypto.randomUUID();
    this.sessionStart = Date.now();
    this.lastActive = Date.now();
    this.persist();
  }

  private rotateIfExpired(): void {
    if (Date.now() - this.lastActive >= this.timeoutMs) {
      this.reset();
    }
  }

  private persist(): void {
    try {
      sessionStorage.setItem(SESSION_ID_KEY, this.sessionId);
      sessionStorage.setItem(SESSION_START_KEY, String(this.sessionStart));
      sessionStorage.setItem(SESSION_LAST_ACTIVE_KEY, String(this.lastActive));
    } catch {
      // sessionStorage unavailable (SSR, privacy mode, etc.)
    }
  }

  private restore(): { sessionId: string; sessionStart: number; lastActive: number } | null {
    try {
      const sessionId = sessionStorage.getItem(SESSION_ID_KEY);
      const sessionStart = sessionStorage.getItem(SESSION_START_KEY);
      const lastActive = sessionStorage.getItem(SESSION_LAST_ACTIVE_KEY);
      if (sessionId && sessionStart && lastActive) {
        return {
          sessionId,
          sessionStart: Number(sessionStart),
          lastActive: Number(lastActive),
        };
      }
    } catch {
      // sessionStorage unavailable
    }
    return null;
  }
}
