import type { FakerPort } from "./port.js";

/** Deterministic PRNG (mulberry32). */
function mulberry32(seed: number): () => number {
  let t = seed >>> 0;
  return () => {
    t += 0x6d2b79f5;
    let r = Math.imul(t ^ (t >>> 15), 1 | t);
    r ^= r + Math.imul(r ^ (r >>> 7), 61 | r);
    return ((r ^ (r >>> 14)) >>> 0) / 4294967296;
  };
}

const LOREM = [
  "alpha",
  "beta",
  "gamma",
  "delta",
  "epsilon",
  "zeta",
  "eta",
  "theta",
];

/** Zero-dependency faker for CI and minimal bundles. */
export function createLightweightFaker(seed = 42): FakerPort {
  const rnd = mulberry32(seed);
  let seq = 0;

  const pick = <T>(items: T[]): T => items[Math.floor(rnd() * items.length)]!;

  return {
    uuid(): string {
      seq += 1;
      return `00000000-0000-4000-8000-${String(seq).padStart(12, "0")}`;
    },
    email(): string {
      return `user${Math.floor(rnd() * 1e6)}@example.com`;
    },
    firstName(): string {
      return pick(["Ada", "Grace", "Linus", "Ken", "Margaret"]);
    },
    lastName(): string {
      return pick(["Lovelace", "Hopper", "Torvalds", "Thompson", "Hamilton"]);
    },
    companyName(): string {
      return `${pick(["Acme", "Globex", "Initech", "Umbrella"])} ${pick(["Labs", "Inc", "Corp"])}`;
    },
    sentence(wordCount = 6): string {
      const words: string[] = [];
      for (let i = 0; i < wordCount; i += 1) {
        words.push(pick(LOREM));
      }
      words[0] = words[0]!.charAt(0).toUpperCase() + words[0]!.slice(1);
      return `${words.join(" ")}.`;
    },
    int(min = 0, max = 100): number {
      return Math.floor(rnd() * (max - min + 1)) + min;
    },
  };
}
