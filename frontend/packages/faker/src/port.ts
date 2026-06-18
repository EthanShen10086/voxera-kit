/** Pluggable fake-data port for tests and fixtures. */
export interface FakerPort {
  uuid(): string;
  email(): string;
  firstName(): string;
  lastName(): string;
  companyName(): string;
  sentence(wordCount?: number): string;
  int(min?: number, max?: number): number;
}

export type FakerProvider = "fakerjs" | "lightweight";

export interface CreateFakerOptions {
  provider?: FakerProvider;
  seed?: number;
}
