export type {
  ExperimentConfig,
  ExperimentVariant,
  ExperimentAssignment,
  ExperimentExposure,
} from './types';

export type { IExperimentProvider } from './provider';

export { ExperimentClient } from './client';
export { HttpProvider } from './http-provider';
export { PostHogProvider } from './posthog-provider';
export { createExperimentHook, createVariantHook } from './hooks';
