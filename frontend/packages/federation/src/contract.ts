export interface IRemoteModule {
  mount(
    container: HTMLElement,
    props?: Record<string, unknown>,
  ): void | Promise<void>;

  unmount(): void | Promise<void>;

  onThemeChange?(preset: unknown): void;
  onLocaleChange?(locale: string): void;
  onRouteChange?(route: string): void;
}

export interface RemoteModuleConfig {
  url: string;
  scope: string;
  module: string;
  props?: Record<string, unknown>;
}
