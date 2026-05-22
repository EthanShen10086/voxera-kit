import type { ExperimentClient } from './client';
import type { ExperimentVariant } from './types';

/**
 * Creates a reactive experiment hook factory.
 * Usage with React: const useExperiment = createExperimentHook(client, React.useState, React.useEffect);
 */
export function createExperimentHook(
  client: ExperimentClient,
  useState: <T>(init: T) => [T, (v: T) => void],
  useEffect: (fn: () => void | (() => void), deps: unknown[]) => void,
) {
  return function useExperiment(key: string): {
    variant: ExperimentVariant | null;
    isLoading: boolean;
  } {
    const [variant, setVariant] = useState<ExperimentVariant | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
      let cancelled = false;

      client.getVariant(key).then((v) => {
        if (cancelled) return;
        setVariant(v);
        setIsLoading(false);
        if (v) client.exposure(key);
      }).catch(() => {
        if (cancelled) return;
        setIsLoading(false);
      });

      return () => {
        cancelled = true;
      };
    }, [key]);

    return { variant, isLoading };
  };
}

/**
 * Creates a hook for checking if user is in a specific variant.
 */
export function createVariantHook(
  client: ExperimentClient,
  useState: <T>(init: T) => [T, (v: T) => void],
  useEffect: (fn: () => void | (() => void), deps: unknown[]) => void,
) {
  return function useVariant(key: string, variantKey: string): boolean {
    const [isMatch, setIsMatch] = useState(false);

    useEffect(() => {
      let cancelled = false;

      client.getVariant(key).then((v) => {
        if (cancelled) return;
        setIsMatch(v?.key === variantKey);
      }).catch(() => {
        // variant fetch failed, remain false
      });

      return () => {
        cancelled = true;
      };
    }, [key, variantKey]);

    return isMatch;
  };
}
