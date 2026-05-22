import type { ExperimentAssignment, ExperimentExposure } from './types';

export interface IExperimentProvider {
  name: string;
  fetchAssignment(
    experimentKey: string,
    userId: string,
    attributes?: Record<string, unknown>,
  ): Promise<ExperimentAssignment | null>;
  fetchAllAssignments(
    userId: string,
    attributes?: Record<string, unknown>,
  ): Promise<ExperimentAssignment[]>;
  reportExposure(exposure: ExperimentExposure, userId: string): Promise<void>;
  reportMetric(
    experimentKey: string,
    userId: string,
    metricKey: string,
    value: number,
  ): Promise<void>;
}
